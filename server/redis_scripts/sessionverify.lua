login = ARGV[1]
uuid_old = ARGV[2]
uuid_new = ARGV[3]

if not login then 
    return 'No login'
end

if login == '' then
    return 'Empty login'
end

sessionKey = 'user:' .. user.login .. ':session'

-- uuid_old is not valid => drop session
if not uuid_old then
    redis.call('DEL', sessionKey)
    return 'No uuid_old'
end

-- uuid_new is not valid => drop session
if not uuid_new then
    redis.call('DEL', sessionKey)
    return 'No uuid_new'
end

-- uuid is not valid => drop session
if uuid_old == '' then
    redis.call('DEL', sessionKey)
    return 'No uuid'
end

-- uuid is not valid => drop session
if uuid_new == '' then
    
end

-- after this timeout user will be automatically logged out
UUID_TTL = redis.call('HGET', 'settings', 'uuid_ttl') or 250