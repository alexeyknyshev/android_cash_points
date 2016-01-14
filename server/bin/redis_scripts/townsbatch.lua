local reqPayload = ARGV[1]

if not reqPayload then
    return redis.error_reply('no json request payload')
end

local req = cjson.decode(reqPayload)

if not req.towns then
    return redis.error_reply('no json request key: towns')
end

local result = {}
for _, townId in pairs(req.towns) do
    local townJsonData = redis.call('GET', 'town:' .. tostring(townId))
    if townJsonData then
        result[#result + 1] = cjson.decode(townJsonData)
    end
end

if #result == 0 then
    return '[]'
end

return cjson.encode(result)
