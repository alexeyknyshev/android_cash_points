reqPayload = ARGV[1] 
uuid = ARGV[2]

if not reqPayload then
    return 'No json data'
end

if not uuid then
    return 'No uuid data'
end

if uuid == '' then
    return 'Empty uuid data'
end

req = cjson.decode(reqPayload)

if not req.login then
    return 'No such login'
end

if not req.password then
    return 'No such password'
end

userJson = redis.call('GET', 'user:' .. user.login)
if not userJson then
    return 'No such user account'
end

user = cjson.decode(userJson)
if req.password ~= user.password then
    return 'Invalid password'
end

-- after this timeout user will be automatically logged out
UUID_TTL = redis.call('HGET', 'settings', 'uuid_ttl') or 250
redis.call('SETEX', 'user:' .. user.login .. ':session', UUID_TTL, uuid)
return ''
