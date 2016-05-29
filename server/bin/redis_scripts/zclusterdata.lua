local cluster = ARGV[1]

if not cluster then
    return redis.error_reply('no such required argument: cluster')
end

if redis.call('EXISTS', cluster) == 0 then
    return redis.error_reply('no such redis key: ' .. cluster)
end

local cpIdList = redis.call('SMEMBERS', cluster)

local avgLon = 0.0
local avgLat = 0.0
for _, cpId in pairs(cpIdList) do
    local cpJson = redis.call('GET', 'cp:' .. cpId)
    if cpJson then
        local cp = cjson.decode(cpJson)
        if cp then
            avgLon = avgLon + cp.longitude
            avgLat = avgLat + cp.latitude
        end
    end
end

if #cpIdList > 0 then
    avgLon = avgLon / #cpIdList
    avgLat = avgLat / #cpIdList
end

local clusterData = {
    longitude = avgLon,
    latitude = avgLat,
    size = #cpIdList
}

return cjson.encode(clusterData)
