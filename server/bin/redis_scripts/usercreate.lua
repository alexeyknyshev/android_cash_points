local userPayload = ARGV[1]
local lang = ARGV[2] or 'en'

local tr = function(msg)
    return redis.call('HGET', 'msg:' .. msg, lang) or msg
end

local checkStr = function(str, pattern)
    local strLen = string.len(str)
    local s, e = string.find(str, pattern)
    if not s or s > 1 then
        return string.sub(str, 1, (s or strLen + 1) - 1)
    elseif e < strLen then
        local startOfValidStr = string.find(str, pattern, e + 1) or strLen + 1
        return string.sub(str, e + 1, startOfValidStr - 1)
    end
    return nil
end

if not userPayload then
    return redis.error_reply('no such json payload')
end

local user = cjson.decode(userPayload)

if not user.login then
    return redis.error_reply('no such required key: login')
end

if not user.password then
    return redis.error_reply('no such required: password')
end

local loginLen = string.len(user.login)
local USER_LOGIN_MIN_LEN = tonumber(redis.call('HGET', 'settings', 'user_login_min_length')) or 4
if loginLen < USER_LOGIN_MIN_LEN then
    return tr('User login length cannot be less than') .. ': ' .. tostring(USER_LOGIN_MIN_LEN)
end

local passwordLen = string.len(user.password)
local USER_PASSWORD_MIN_LEN = tonumber(redis.call('HGET', 'settings', 'user_password_min_length')) or 4
if passwordLen < USER_PASSWORD_MIN_LEN then
    return tr('User password length cannot be less than') .. ': ' .. tostring(USER_PASSWORD_MIN_LEN)
end

-- check login is alpha numeric
local errStr = checkStr(user.login, "[_%a][_%w]*")
if errStr then
    return tr('User login contains invalid character') .. ': ' .. errStr
end


-- check password is alpha numeric
errStr = checkStr(user.password, "[_%w]+")
if errStr then
    return tr('User password contains invalid character') .. ': ' .. errStr
end

-- check user already exists
if redis.call('EXISTS', 'user:' .. user.login) == 1 then
    return tr('User already exists with login') .. ': ' .. user.login
end

-- save user data
redis.call('SET', 'user:' .. user.login, userPayload)

return true
