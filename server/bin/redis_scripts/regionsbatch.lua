local result = {}
local regionIds = redis.call('ZRANGE', 'regions', 0, -1)
for _, regionId in pairs(regionIds) do
    local regionJsonData = redis.call('GET', 'region:' .. tostring(regionId))
    if regionJsonData then
        result[#result + 1] = cjson.decode(regionJsonData)
    end
end

if #result == 0 then
    return '[]'
end

return cjson.encode(result)
