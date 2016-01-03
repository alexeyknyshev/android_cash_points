local reqPayload = ARGV[1]

if not reqPayload then
    return redis.error_reply('no json request payload')
end

local req = cjson.decode(reqPayload)

if not req.towns then
    return redis.error_reply('no json request key: towns')
end

local result = {}
for _, townid in pairs(req.towns) do
    result[#result + 1] = redis.call('GET', 'town:' .. tostring(townid))
end

return result
