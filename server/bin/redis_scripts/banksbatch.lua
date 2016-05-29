local reqPayload = ARGV[1]

if not reqPayload then
    return redis.error_reply('no such json payload')
end

local req = cjson.decode(reqPayload)

if not req.banks then
    return redis.error_reply('no such required argument: banks')
end

local result = {}
for _, bankId in pairs(req.banks) do
    local bankJsonData = redis.call('GET', 'bank:' .. tostring(bankId))
    if bankJsonData then
        result[#result + 1] = cjson.decode(bankJsonData)
    end
end

if #result == 0 then
    return '[]'
end

return cjson.encode(result)
