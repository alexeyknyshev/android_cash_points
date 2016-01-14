local reqPayload = ARGV[1]

if not reqPayload then
    return redis.error_reply('no such json payload')
end

local req = cjson.decode(reqPayload)

if not req.from then
    return redis.error_reply('no such required argument: from')
end

if not req.to then
    return redis.error_reply('no such required argument: to')
end

local fromType = type(req.from)
if fromType ~= 'number' then
    return redis.error_reply('wrong type of argument "from": ' .. fromType)
end

local toType = type(req.to)
if toType ~= 'number' then
    return redis.error_reply('wrong type of argument "to": ' .. toType)
end

if req.to - req.from > 512 then
    req.to = req.from + 512
end

local cpIdList = redis.call('ZREVRANGE', 'cp:history', req.from, req.to)

if #cpIdList == 0 then
    return '[]'
end

return cjson.encode(cpIdList)
