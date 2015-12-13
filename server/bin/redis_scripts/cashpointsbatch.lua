local userPayload = ARGV[1]

if not userPayload then
    return 'No json data'
end

local req = cjson.decode(userPayload)

if not req.cashpoints then
    return 'No such key: cashpoints'
end

local result = { cashpoints = { } }
for _, cpid in pairs(req.cashpoints) do
    local cpkey = 'cp:' .. tostring(cpid)
    local cpdata = redis.call('GET', cpkey)
    if cpdata then
        local cp = cjson.decode(cpdata)
        cp.version = redis.call('GET', cpkey .. ':version') or 0
        cp.timestamp = redis.call('GET', cpkey .. ':timestamp') or 0
        cp.owner = 0
        result.cashpoints[#result.cashpoints + 1] = cp
    end
end

return cjson.encode(result)
