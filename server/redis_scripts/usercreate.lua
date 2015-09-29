USER_LOGIN_MIN_LEN = 4
USER_PASSWORD_MIN_LEN = 4

userPayload = ARGV[1]

if userPayload == nil then
    return 'No json data'
end

user = cjson.decode(userPayload)

if not user.login then
    return 'No such key: login'
end

if not user.password then
    return 'No such key: password'
end

if string.len(user.login) < USER_LOGIN_MIN_LEN then
    return 'User login must be ' .. tostring(USER_LOGIN_MIN_LEN) .. ' characters at least'
end

if string.len(user.password) < USER_PASSWORD_MIN_LEN then
    return 'User password must be ' .. tostring(USER_PASSWORD_MIN_LEN) .. ' characters at least'
end

-- check user already exists
if redis.call('EXISTS', 'user:' .. user.login) == 1 then
  return 'User with login: "' .. user.login .. '" already exists'
end

-- save user data
redis.call('SET', 'user:' .. user.login, userPayload)

return ""
