local cashpointPayload = ARGV[1]

if not cashpointPayload then
    return 'No json data'
end

local cp = cjson.decode(bankPayload)

if not cp.latitude then
end

if not cp.longitude then
end

if not cp.bank_id then
end

