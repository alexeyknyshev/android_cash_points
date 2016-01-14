local cashpointPayload = ARGV[1]
local timestamp = ARGV[2]
local lang = ARGV[3] or 'en'

if not cashpointPayload then
    return redis.error_reply('no such json payload')
end

if not timestamp then
    return redis.error_reply('no timestamp specified')
end

local cpData = cjson.decode(cashpointPayload)

local allCpProperties = {
    id = 0,
    type = "",
    bank_id = 0,
    town_id = 0,
    longitude = 0.0,
    latitude = 0.0,
    address = "",
    address_comment = "",
    metro_name = "",
    free_access = false,
    main_office = false,
    without_weekend = false,
    round_the_clock = false,
    works_as_shop = false,
    schedule = "",
    tel = "",
    additional = "",
    rub = false,
    usd = false,
    eur = false,
    cash_in = false
}
local allCpTypes = {
    atm = true,
    cash = true,
    office = true
}

local tr = function(msg)
    return redis.call('HGET', 'msg:' .. msg, lang) or msg
end

local getQuadKey = function(longitude, latitude, zoom)
    if not longitude or not latitude then
        return ''
    end

    local geoRectPart = function(minLon, maxLon, minLat, maxLat, lon, lat)
        local midLon = (minLon + maxLon) * 0.5
        local midLat = (minLat + maxLat) * 0.5

        local quad = ""
        if lat < midLat then
            maxLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = '0'
            else
                minLon = midLon
                quad = '1'
            end
        else
            minLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = '2'
            else
                minLon = midLon
                quad = '3'
            end
        end

        return minLon, maxLon, minLat, maxLat, quad
    end

    local minLon = -180.0
    local maxLon = 180.0

    local minLat = -85.0
    local maxLat = 85.0

    if longitude < minLon or maxLon < longitude then
        return ""
    end

    if latitude < minLat or maxLat < latitude then
        return ""
    end

    quadKey = ""
    for currentZoom = 0, zoom do
        local q = ""
        minLon, maxLon, minLat, maxLat, q = geoRectPart(minLon, maxLon, minLat, maxLat, longitude, latitude)
        quadKey = quadKey .. q
    end
    return quadKey
end

local validateExistingFields = function(cpData)
    for k, v in pairs(cpData) do
        local property = allCpProperties[k]
        if property == nil then
            return tr('Unknown field') .. ': ' .. tostring(k)
        end

        local propertyType = type(property)
        local cpFieldType = type(v) 
        if propertyType ~= cpFieldType then
            return tr('Wrong type of field') .. ': ' .. tostring(property) .. '. ' .. tr('Expected') ..
                   ' "' .. tostring(propertyType) .. '" ' .. tr('but got') .. ' "' .. tostring(cpFieldType) .. '.'
        end
    end

    if cpData.type then
        if not allCpTypes[cpData.type] then
            return tr('Invalid cashpoint type') .. ':' .. tostring(cpData.type)
        end
    end

    local err = validateBankId(cpData)
    if err then
        return err
    end

    return validateTownId(cpData)
end

local creationValidate = function(cpData)
    for k, v in pairs(allCpProperties) do
        if cpData[k] == nil then
            return tr('Missing required field') .. ': ' .. tostring(k)
        end
    end
end

local validateBankId = function(cpData)
    if cpData.bank_id then
        local bankIdStr = tostring(cpData.bank_id)
        if redis.call('EXISTS', 'bank:' .. bankIdStr) == 0 then
            return tr('Bank does not exist with id') .. ': ' .. bankIdStr
        end
    end
end

local validateTownId = function(cpData)
    if cpData.town_id then
        local townIdStr = tostring(cpData.town_id)
        if redis.call('EXISTS', 'town:' .. townIdStr) == 0 then
            return tr('Town does not exist with id') .. ': ' .. townIdStr
        end
    end
end

local err = validateExistingFields(cpData)
if err then
    return err
end

local eqCpData = function(a, b)
    for k, v in pairs(a) do
        if b[k] ~= v then
            return false
        end
    end
    return true
end

if cpData.id then -- editing existing cashpoint

    local cpDataIdStr = tostring(cpData.id)
    local oldCpDataJson = redis.call('GET', 'cp:' .. cpDataIdStr)
    if not oldCpDataJson then
        return tr('Cashpoint does not exist with id') .. ': ' .. cpDataIdStr
    end
    local oldCpData = cjson.decode(oldCpDataJson)

    if eqCpData(cpData, oldCpData) then
        return false -- signal: no data changed
    end

    local oldBankId = oldCpData.bank_id
    local oldTownId = oldCpData.town_id

    local oldQuadKey = getQuadKey(oldCpData.longitude, oldCpData.latitude, 15)
    local newQuadKey = getQuadKey(cpData.longitude, cpData.latitude, 15)

    if string.len(newQuadKey) == 0 then
        return tr('Coordinates of cashpoint are out of range')
    end

    for k, v in pairs(cpData) do
        oldCpData[k] = v
    end
    oldCpData.version = oldCpData.version + 1
    oldCpData.timestamp = timestamp

    -- Coordinates are changed => update geoindex
    if cpData.longitude and cpData.latitude then
        local res = redis.pcall('GEOADD', 'cashpoints', cpData.longitude, cpData.latitude, cpData.id)
        if type(res) == 'table' and res.err then
            return tr('Coordinates of cashpoint are out of range')
        end
    end

    -- Bank id change proposed => check if it has been really changed and then update 
    if cpData.bank_id and cpData.bank_id ~= oldBankId then
        redis.call('SREM', 'cp:bank:' .. tostring(oldBankId))
        redis.call('SADD', 'cp:bank:' .. tostring(cpData.bank_id))
    end

    -- Town id change proposed => check if it has been really changed and then update
    if cpData.town_id and cpData.town_id ~= oldTownId then
        redis.call('SREM', 'cp:town:' .. tostring(oldTownId))
        redis.call('SADD', 'cp:town:' .. tostring(cpData.town_id))
    end

    redis.call('SET', 'cp:' .. cpDataIdStr, cjson.encode(oldCpData))

    -- TODO: update clusters
    if oldQuadKey ~= newQuadKey then -- need to update clusters
        for zoom = 10, 16 do
            local currentOldQuadKey = string.sub(oldQuadKey, 1, zoom)
            local currentNewQuadKey = string.sub(newQuadKey, 1, zoom)
            if currentNewQuadKey ~= currentOldQuadKey then
                redis.call('DEL', 'cluster:' .. currentOldQuadKey .. ':data')
                redis.call('SREM', 'cluster:' .. currentOldQuadKey, cpData.id)
                redis.call('SADD', 'cluster:' .. currentNewQuadKey, cpData.id)
            end
        end
    end

    redis.call('ZADD', 'cp:history', timestamp, cpData.id)
else -- adding new cashpoint
    err = creationValidate(cpData)
    if err then
        return err
    end

    local quadKey = getQuadKey(cp.longitude, cp.latitude, 15)
    if string.len(quadKey) == 0 then
        return tr('Coordinates of cashpoint are out of range')
    end

    local res = redis.pcall('GEOADD', 'cashpoints', cpData.longitude, cpData.latitude, cpData.id)
    if type(res) == 'table' and res.err then
        return tr('Coordinates of cashpoint are out of range')
    end

    redis.call('SADD', 'cp:town:' .. tostring(cpData.town_id))
    redis.call('SADD', 'cp:bank:' .. tostring(cpData.bank_id))

    local nextCpIdKey = 'cp_next_id'
    if redis.call('HGET', 'settings', 'testing_mode') == '1' then
        nextCpIdKey = 'test_cp_next_id'
    end

    local nextId = redis.call('INCR', nextCpIdKey)
    redis.call('SET', 'cp:' .. tostring(nextId), cashpointPayload)

    -- TODO: update clusters
    for zoom = 10, 16 do
        local currentQuadKey = string.sub(quadKey, 1, zoom)
        redis.call('DEL', 'cluster:' .. currentQuadKey .. ':data')
        --redis.call('SREM', 'cluster:' .. currentQuadKey, )
        redis.call('SADD', 'cluster:' .. currentQuadKey, nextId)
    end

    redis.call('ZADD', 'cp:history', timestamp, nextId)
end

return true
