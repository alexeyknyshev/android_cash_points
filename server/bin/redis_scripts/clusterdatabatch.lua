local result = {}

for i, quadKey in ipairs(ARGV) do
    local clusterData = {
        longitude = 0.0,
        latitude = 0.0,
        size = 0,
        id = quadKey
    }

    local cachedData = redis.call('GET', 'cluster:' .. quadKey .. ':data')
    if cachedData then
        result[#result + 1] = cjson.decode(cachedData)
    else
        local members = redis.call('SMEMBERS', 'cluster:' .. quadKey)

        if #members > 0 then
            clusterData.size = #members
            for _, cpid in pairs(members) do
                local cpkey = 'cp:' .. tostring(cpid)
                local cpdata = redis.call('GET', cpkey)
                if cpdata then
                    local cp = cjson.decode(cpdata)
                    clusterData.longitude = clusterData.longitude + cp.longitude
                    clusterData.latitude = clusterData.latitude + cp.latitude
                else
                    return redis.error_reply('cannot unpack json data for ' .. cpkey)
                end
            end

            clusterData.longitude = clusterData.longitude / clusterData.size
            clusterData.latitude = clusterData.latitude / clusterData.size

            result[#result + 1] = clusterData
            redis.call('SET', 'cluster:' .. quadKey .. ':data', cjson.encode(clusterData))
        end
    end
end

if #result == 0 then
    return '[]'
end
return cjson.encode(result)
