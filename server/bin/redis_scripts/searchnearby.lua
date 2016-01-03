local geoset = KEYS[1]
local reqPayload = ARGV[1]

if not geoset then
    return redis.error_reply('no geoset specified')
end

if not reqPayload then
    return redis.error_reply('no json request payload')
end

local req = cjson.decode(reqPayload)

if not req.longitude then
    return redis.error_reply('no json request key: longitude')
end

if not req.latitude then
    return redis.error_reply('no json request key: latitude')
end

if not req.radius then
    return redis.error_reply('no json request key: radius')
end

if req.radius <= 0.0 then
   return redis.error_reply('search radius must be positive')
end

req.radius = math.floor(req.radius) + 1

return redis.call('GEORADIUS', geoset, req.longitude, req.latitude, req.radius, 'm')
