local reqPayload = ARGV[1]

if not reqPayload then
    return 'No json data'
end

local req = cjson.decode(reqPayload)

if not req.from then
    return 'No from'
end

if not req.to then
    return 'No to'
end

local fromType = type(req.from)
if fromType ~= 'number' then
    return 'Wrong type of from: ' .. fromType
end

local toType = type(req.to)
if toType ~= 'number' then
    return 'Wrong type of to: ' .. toType
end

if req.to - req.from > 512 then
    req.to = req.from + 512
end

return redis.call('ZREVRANGE', 'cp:history', req.from, req.to)
