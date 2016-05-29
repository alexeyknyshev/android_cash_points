local reqPayload = ARGV[1]
local countLimit = tonumber(ARGV[2]) or 32
local setExpireTime = tonumber(ARGV[3]) or 300

if not reqPayload then
    return redis.error_reply('no such json payload')
end

local req = cjson.decode(reqPayload)

if not req.longitude then
    return redis.error_reply('no such required argument: longitude')
end

if not req.latitude then
    return redis.error_reply('no such required argument: latitude')
end

if not req.radius then
    return redis.error_reply('no such required argument: radius')
end

if req.radius <= 0.0 then
    return redis.error_reply('search radius must be positive')
end

if not req.zoom then
    return redis.error_reply('no such required argument: zoom')
end

req.radius = math.floor(req.radius) + 1

req.filter = req.filter or {}
--req.filter = { bank_id = { 322, 325 } }
--req.filter = { bank_id = { 325, 4045, 77487 } }
--req.filter = { bank_id = { 4045 } }
--req.filter = { bank_id = { 77487, 322, 4045 }, round_the_clock = true }
--req.filter = { bank_id = { 322 }, type = "atm" }
--req.filter = { bank_id = { 322, 325 }, type = "office" }
--req.filter = req.filter or {}

local supportedFilters = {
    type = "",
    free_access = true,
    main_office = true,
    without_weekend = true,
    round_the_clock = true,
    works_as_shop = true,
    rub = true,
    usd = true,
    eur = true,
    cash_in = true,
    bank_id = 0
}
local supportedFiltersOrder = {
    "type",
    "free_access",
    "main_office",
    "without_weekend",
    "round_the_clock",
    "works_as_shop",
    "rub",
    "usd",
    "eur",
    "cash_in",
    "bank_id"
}

local enabledFilters = {}
for k, _ in pairs(supportedFilters) do
    if req.filter[k] ~= nil then
        enabledFilters[k] = req.filter[k]
    end
end

if countLimit <= 0 then
    return redis.error_reply('towns count limit must be positive')
end

local createBankCpUnion = function(bankIdList)
    if type(bankIdList) ~= 'table' then
        return "", redis.error_reply('filter.bank_id must be an array')
    end

    -- sort to order ids => prevent bankIdChain variations
    table.sort(bankIdList, function(a, b) return a < b end)

    -- remove duplicates
    local prevBankId = 0
    local i = 1
    while i <= #bankIdList do
        if bankIdList[i] == prevBankId then
            table.remove(bankIdList, i)
        else
            prevBankId = bankIdList[i]
            i = i + 1
        end
    end

    local bankCpSetList = {}
    for _, bankId in ipairs(bankIdList) do
        if type(bankId) == 'number' then
            bankCpSetList[#bankCpSetList + 1] = 'cp:bank:' .. tostring(math.floor(bankId))
        else
            return "", redis.error_reply('filter.bank_id contains non-numerical value') 
        end
    end

    local bankCpSetListSize = #bankCpSetList
    if bankCpSetListSize > 0 then
        local bankIdChain = table.concat(bankIdList, ':')
        local bankCpUnion = 'cp:bank:' .. bankIdChain
        if bankCpSetListSize > 1 and redis.call('EXISTS', bankCpUnion) == 0 then
            redis.call('SUNIONSTORE', bankCpUnion, unpack(bankCpSetList))
            redis.call('EXPIRE', bankCpUnion, setExpireTime)
        end
        return bankIdChain
    end
end

local filterSet = function(townRecList, filter)
   local interSetList = {}
   local interChain = ""

   if next(filter) ~= nil then
       if filter.bank_id then
           local bankIdChain, err = createBankCpUnion(filter.bank_id)
           if err then
               return {}, filter, err
           end
           filter.bank_id = nil

           if bankIdChain then
               interSetList[#interSetList + 1] = 'cp:bank:' .. bankIdChain
               interChain = interChain .. ':bank:' .. bankIdChain
           end
       end
   end

   local interSetListSize = #interSetList

   local resultSetList = {}

   for _, townRec in pairs(townRecList) do
       local townIdStr = townRec[1]
       local lon = townRec[2][1]
       local lat = townRec[2][2]

       local townId = tonumber(townIdStr)
       local townCpSet = 'cp:town:' .. townIdStr

       local clusterData = {
           longitude = tonumber(lon),
           latitude = tonumber(lat),
           id = townId
       }

       if interSetListSize > 0 then
           clusterData.set = townCpSet .. interChain
           if redis.call('EXISTS', clusterData.set) == 0 then
               clusterData.size = redis.call('SINTERSTORE', clusterData.set, townCpSet, unpack(interSetList))
               redis.call('EXPIRE', clusterData.set, setExpireTime)
           else
               clusterData.size = redis.call('SCARD', clusterData.set)
           end
       else
           clusterData.set = townCpSet
           clusterData.size = redis.call('SCARD', clusterData.set)
       end

       -- intersection can produce empty set => don't return empty clusters
       if clusterData.size > 0 then
           resultSetList[#resultSetList + 1] = clusterData
       end
   end

   return resultSetList, filter
end

local filterData = function(clusterDataList, filter)
    if next(filter) == nil then
        return clusterDataList
    end

    local chain = ""
    for i, field in ipairs(supportedFiltersOrder) do
        if filter[field] ~= nil then
            chain = chain .. ':' .. field .. ':' .. tostring(filter[field])
        end
    end

--    if chain then
--       return {}, redis.error_reply(chain)
--    end

    if string.len(chain) == 0 then
        return clusterDataList
    end

    local result = {}
    for _, clusterData in ipairs(clusterDataList) do
        local expectedSetData = clusterData.set .. chain .. ':data'
        local expectedSetDataJson = redis.call('GET', expectedSetData)
        if not expectedSetDataJson then
            local count = 0

            local avgLon = 0.0
            local avgLat = 0.0

            local cpIdList = redis.call('SMEMBERS', clusterData.set)
            for _, id in pairs(cpIdList) do
                local cpJsonData = redis.call('GET', 'cp:' .. tostring(id))
                if cpJsonData then
                    local cp = cjson.decode(cpJsonData)
                    if cp then
                        local matches = true
                        for k, v in pairs(filter) do
                            if cp[k] ~= v then
                                matches = false
                                break
                            end
                        end
                        if matches then
                            avgLon = avgLon + cp.longitude
                            avgLat = avgLat + cp.latitude
                            count = count + 1
                        end
                    end
                end
            end

            if count > 0 then
                avgLon = avgLon / count
                avgLat = avgLat / count
            end

            local data = {
                size = count,
                longitude = avgLon,
                latitude = avgLat,
            }
            redis.call('SETEX', expectedSetData, setExpireTime, cjson.encode(data))
            clusterData.longitude = avgLon
            clusterData.latitude = avgLat
            clusterData.size = count
        else
--          if clusterData then
--              return {}, redis.error_reply(type(expectedSetDataJson))
--          end
            local edata = cjson.decode(expectedSetDataJson)
            if edata then
                clusterData.longitude = edata.longitude
                clusterData.latitude = edata.latitude
                clusterData.size = edata.size or 0
            end
        end

        if clusterData.size > 0 then
            clusterData.set = expectedSetData
            result[#result + 1] = clusterData
        end
    end
    return result
end

local geoset = 'towns'
if req.zoom <= 8.0 or req.radius > 150000 then
    geoset = geoset .. ':big'
end

local townIdList = redis.call('GEORADIUS', geoset, req.longitude, req.latitude, req.radius, 'm', 'WITHCOORD')
local clusterDataList, enabledFilters, err = filterSet(townIdList, enabledFilters)
if err then
    return err
end
clusterDataList, err = filterData(clusterDataList, enabledFilters)
if err then
    return err
end

if #clusterDataList == 0 then
    return '[]'
end

table.sort(clusterDataList, function(a, b) return a.size > b.size end)

local idMappingIndex = {}
for index, clusterData in ipairs(clusterDataList) do
    idMappingIndex[clusterData.id] = index
end

--if result then
--    return cjson.encode(result)
--end

local minDist = req.radius * 0.2

--if #result > countLimit then
--    local minPointsCount = (result[2] or result[1]).size * 0.05
--    for i = #result, 1, -1 do
--        if result[i].size < minPointsCount then
--            table.remove(result, i)
--        end
--    end
--end

for i = 1, #clusterDataList do
    if clusterDataList[i].size ~= 0 then
        for j = #clusterDataList, i + 1, -1 do
            if clusterDataList[j].size ~= 0 then
                local dist = redis.call('GEODIST', geoset, clusterDataList[i].id, clusterDataList[j].id, 'm')
                if dist then
                    dist = tonumber(dist)
                    if dist < minDist then
                        clusterDataList[i].size = clusterDataList[i].size + clusterDataList[j].size
                        clusterDataList[j].size = 0
                    end
                else
                    clusterDataList[i].size = 0
                    clusterDataList[j].size = 0
                end
            end
        end
    end
end

for i = #clusterDataList, 1, -1 do
    if clusterDataList[i].size == 0 then
        table.remove(clusterDataList, i)
    end
end

while #clusterDataList > countLimit do
    table.remove(clusterDataList)
end

-- remove redis cached set info
for i = 1, #clusterDataList do
    clusterDataList[i].set = nil
end

return cjson.encode(clusterDataList)
