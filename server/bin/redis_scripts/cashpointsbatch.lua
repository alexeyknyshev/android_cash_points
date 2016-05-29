local userPayload = ARGV[1]

if not userPayload then
    return redis.error_reply('no such json payload')
end

local req = cjson.decode(userPayload)

if not req.cashpoints then
    return redis.error_reply('no such required argument: cashpoints')
end

local result = {}
for _, cpid in pairs(req.cashpoints) do
    local cpkey = 'cp:' .. tostring(cpid)
    local cpdata = redis.call('GET', cpkey)
    if cpdata then
        local cp = cjson.decode(cpdata)
        result[#result + 1] = cp
    end
end

if #result == 0 then
    return '[]'
end

return cjson.encode(result)
