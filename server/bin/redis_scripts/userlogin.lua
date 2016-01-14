local reqPayload = ARGV[1]
local uuid = ARGV[2]
local lang = ARGV[3] or 'en'

local tr = function(msg)
    return redis.call('HGET', 'msg:' .. msg, lang) or msg
end

if not reqPayload then
    return redis.error_reply('no such json payload')
end

if not uuid then
    return redis.error_reply('no such uuid data')
end

if string.len(uuid) == 0 then
    return redis.error_reply('empty uuid data')
end

local req = cjson.decode(reqPayload)

if not req.login then
    return redis.error_reply('no such required argument: login')
end

if not req.password then
    return redis.error_reply('no such required argument: password')
end

local loginFailedMsg = 'Wrong user login or password'

local userJson = redis.call('GET', 'user:' .. req.login)
if not userJson then
    return tr(loginFailedMsg)
end

local user = cjson.decode(userJson)
if req.password ~= user.password then
    return tr(loginFailedMsg)
end

-- after this timeout user will be automatically logged out
local UUID_TTL = redis.call('HGET', 'settings', 'uuid_ttl') or 250
redis.call('SETEX', 'user:' .. user.login .. ':session', UUID_TTL, uuid)

return true
