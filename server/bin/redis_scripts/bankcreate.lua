local bankPayload = ARGV[1]
local lang = ARGV[2] or 'en'

local tr = function(msg)
    return redis.call('HGET', 'msg:' .. msg, lang) or msg
end

if not bankPayload then
    return redis.error_reply('no such json payload')
end

local bank = cjson.decode(bankPayload)

if not bank.name then
    return redis.error_reply('no such required argument: name')
end

if not bank.name_tr then
    return redis.error_reply('no such required argument: name_tr')
end

if string.len(bank.name) == 0 then
    return tr('Empty bank name')
end

if string.len(bank.name_tr) == 0 then
    return tr('Empty bank transliterated name')
end

if redis.call('HEXISTS', 'bank_ids', bank.name_tr) == 1 then
    return tr('Bank already exists with name') .. ': ' .. bank.name
end

local bankId = redis.call('INCR', 'bank_next_id')
redis.call('HMSET', 'bank_ids', bank.name, bankId)
redis.call('SET', 'bank:' .. bankId, bankPayload)

return true
