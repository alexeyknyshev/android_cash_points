local login = ARGV[1]
local uuid_old = ARGV[2]
local uuid_new = ARGV[3]
local lang = ARGV[4] or 'en'

local tr = function(msg)
    return redis.call('HGET', 'msg:' .. msg, lang) or msg
end

if not login then 
    return redis.error_reply('no such required argument: login')
end

if string.len(login) == 0 then
    return redis.error_reply('login is empty')
end

local sessionKey = 'user:' .. user.login .. ':session'

-- uuid_old is not valid => drop session
if not uuid_old then
    redis.call('DEL', sessionKey)
    return redis.error_reply('no such required argument: uuid_old')
end

-- uuid_new is not valid => drop session
if not uuid_new then
    redis.call('DEL', sessionKey)
    return redis.error_reply('no such required argument: uuid_new')
end

-- uuid is not valid => drop session
if string.len(uuid_old) == 0 then
    redis.call('DEL', sessionKey)
    return redis.error_reply('uuid_old is empty')
end

-- uuid is not valid => drop session
if string.len(uuid_new) == 0 then
    redis.call('DEL', sessionKey)
    return redis.error_reply('uuid_new is empty')
end

-- validate old session key
local sessionKeyVal = redis.call('GET', sessionKey)
if sessionKeyVal == nil then
    return tr('You are not logged in')
end

if sessionKeyVal ~= uuid_old then
    return tr('Invalid session token')
end

-- after this timeout user will be automatically logged out
local UUID_TTL = redis.call('HGET', 'settings', 'uuid_ttl') or 250
redis.call('SETEX', sessionKey, UUID_TTL, uuid_new)

return redis.status_reply('OK')
