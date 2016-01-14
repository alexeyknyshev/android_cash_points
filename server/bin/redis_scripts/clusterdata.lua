local reqPayload = ARGV[1]
local dataExpireTime = tonumber(ARGV[2]) or 300

if not reqPayload then
    redis.error_reply('no such json payload')
end

local req = cjson.decode(reqPayload)

if not req.longitude then
    return redis.error_reply('no such required argument: longitude')
end

if not req.latitude then
    return redis.error_reply('no such required argument: latitude')
end

if not req.zoom then
    return redis.error_reply('no such required argument: zoom')
end

if not req.radius then
    return redis.error_reply('no such required argument: radius')
end

if req.zoom < 1 or req.zoom > 16 then
    return redis.error_reply('required argument zoom is out of range: ' .. tostring(req.zoom))
end

if req.radius <= 0.0 then
    return redis.error_reply('search radius must be positive')
end

req.filter = req.filter or {}

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

local createFilterChain = function(filter)
    if next(filter) == nil then
        return ""
    end

    local chain = ""
    if filter.bank_id then
        if type(filter.bank_id) ~= 'table' then
            return "", redis.error_reply('filter.bank_id must be an array')
        end

        -- sort to order ids => prevent bankIdChain variations
        table.sort(filter.bank_id, function(a, b) return a < b end)

        for _, bankId in ipairs(filter.bank_id) do
            if type(bankId) == 'number' then
                chain = ':' .. tostring(math.floor(bankId))
            else
                return "", redis.error_reply('filter.bank_id contains non-numerical value') 
            end
        end

        if string.len(chain) > 0 then
            chain = ':bank' .. chain
        end
    end

    for i, field in ipairs(supportedFiltersOrder) do
        if field ~= 'bank_id' and filter[field] ~= nil then
            chain = chain .. ':' .. field .. ':' .. tostring(filter[field])
        end
    end

    return chain
end

local filterData = function(quadKeyList, filter)
    local chain, err = createFilterChain(filter)
    if err then
        return {}, err
    end

    local result = {}
    for _, quadKey in pairs(quadKeyList) do
        local clusterJsonDataKey = 'cluster:' .. tostring(quadKey) .. chain .. ':data'
        local clusterJsonData = redis.call('GET', clusterJsonDataKey)
        if clusterJsonData then
            local clusterData = cjson.decode(clusterJsonData)
            if clusterData then
                if clusterData.size > 0 then
                    result[#result + 1] = clusterData
                end
            else
                return {}, redis.error_reply('invalid cluster json data in key: ' .. clusterJsonDataKey)
            end
        else
            local clusterData
            -- no filtering
            if string.len(chain) == 0 then
                local cpCount = tonumber(redis.call('SCARD', 'cluster:' .. tostring(quadKey)))
                if cpCount > 0 then
                    local pos = redis.call('GEOPOS', zclusterName, quadKey)
                    if pos and pos[1] then
                        local lon = tonumber(pos[1][1])
                        local lat = tonumber(pos[1][2])
                        if lon and lat then
                            clusterData = {
                                id = quadKey,
                                longitude = lon,
                                latitude = lat,
                                size = cpCount
                            }
                        end
                    end
                end
            else
                local count = 0

                local avgLon = 0.0
                local avgLat = 0.0

                local cpIdList = redis.call('SMEMBERS', 'cluster:' .. tostring(quadKey))
                for _, id in pairs(cpIdList) do
                    local cpJsonData = redis.call('GET', 'cp:' .. tostring(id))
                    if cpJsonData then
                        local cp = cjson.decode(cpJsonData)
                        if cp then
                            local matches = true

                            for k, v in pairs(filter) do
                                local filterType = type(v)
                                if filterType == 'table' then
                                    -- match any policy for nested array (variants) filters
                                    matches = false
                                    for _, var in ipairs(v) do
                                        if cp[k] == var then
                                            matches = true
                                            break
                                        end
                                    end
--                                    return {}, redis.error_reply("match: " .. cjson.encode(v))
                                else
                                    -- match all policy
                                    if cp[k] ~= v then
                                        matches = false
                                    end
                                end
                                -- match all policy for all filters
                                if not matches then
                                    break
                                end
                            end
                            if matches then
                                avgLon = avgLon + cp.longitude
                                avgLat = avgLat + cp.latitude
                                count = count + 1
                                --return {}, redis.error_reply("match: " .. cpJsonData)
                                --return {}, redis.error_reply('filter: ' .. cjson.encode(filter))
                            end
                        end
                    end
                end

                if count > 0 then
                    avgLon = avgLon / count
                    avgLat = avgLat / count
                end

                clusterData = {
                    id = quadKey,
                    longitude = avgLon,
                    latitude = avgLat,
                    size = count,
                }
            end

            if clusterData then
                redis.call('SETEX', clusterJsonDataKey, dataExpireTime, cjson.encode(clusterData))
                if clusterData.size > 0 then
                    result[#result + 1] = clusterData
                end
            end
        end
    end
    return result
end

-- local result = {}

local zclusterName = 'zcluster:' .. tostring(req.zoom)
local quadKeyList = redis.call('GEORADIUS', zclusterName, req.longitude, req.latitude, req.radius, 'm')
local clusterDataList, err = filterData(quadKeyList, req.filter)
if err then
    return err
end
-- local clusterDataList = {}

--[[
for _, quadKey in pairs(quadKeyList) do
    local clusterJsonData = redis.call('GET', 'cluster:' .. tostring(quadKey) .. ':data')
    if clusterJsonData then
        local cluster = cjson.decode(clusterJsonData)
        if cluster then
            result[#result + 1] = cluster
        else
            return redis.error_reply('invalid cluster json data for quadKey: ' .. tostring(quadKey))
        end
    else
        local cpCount = tonumber(redis.call('SCARD', 'cluster:' .. tostring(quadKey)))
        if cpCount > 0 then
            local pos = redis.call('GEOPOS', zclusterName, quadKey)
            if pos and pos[1] then
                local lon = tonumber(pos[1][1])
                local lat = tonumber(pos[1][2])
                if lon and lat then
                    clusterJsonData = {
                        id = quadKey,
                        longitude = lon,
                        latitude = lat,
                        size = cpCount
                    }
                    redis.call('SET', 'cluster:' .. tostring(quadKey) .. ':data', cjson.encode(clusterJsonData))
                    result[#result + 1] = clusterJsonData
                end
            end
        end
    end
end]]--

if #clusterDataList == 0 then
    return '[]'
end
return cjson.encode(clusterDataList)
