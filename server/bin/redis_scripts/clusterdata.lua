local reqPayload = ARGV[1]

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

local result = {}

local zclusterName = 'zcluster:' .. tostring(req.zoom)
local quadKeyList = redis.call('GEORADIUS', zclusterName, req.longitude, req.latitude, req.radius, 'm')
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
        local cpCount = redis.call('SCARD', 'cluster:' .. tostring(quadKey))
        --return cpCount
        if cpCount and tonumber(cpCount) > 0 then
            cpCount = tonumber(cpCount)
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
end

if #result == 0 then
    return '[]'
end
return cjson.encode(result)

