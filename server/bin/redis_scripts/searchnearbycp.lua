local reqPayload = ARGV[1]

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

req.filter = req.filter or {}
--req.filter = { bank_id = { 325, 322 }, round_the_clock = true }

local supportedFilters = {
    type = "",
    free_access = true,
    main_office = true,
    without_weekend = true,
    round_the_clock = true,
    works_as_shop = true,
    rub = true,
    usd = true,
    eur = true,
    cash_in = true,
    bank_id = 0
}
local supportedFiltersOrder = {
    "type",
    "free_access",
    "main_office",
    "without_weekend",
    "round_the_clock",
    "works_as_shop",
    "rub",
    "usd",
    "eur",
    "cash_in",
    "bank_id"
}

local enabledFilters = {}
for k, _ in pairs(supportedFilters) do
    if req.filter[k] ~= nil then
        enabledFilters[k] = req.filter[k]
    end
end

if req.filter.bank_id then
    table.sort(req.filter.bank_id, function(a, b) return a < b end)

    -- remove duplicates
    local prevBankId = 0
    local i = 1
    while i <= #req.filter.bank_id do
        if req.filter.bank_id[i] == prevBankId then
            table.remove(req.filter.bank_id, i)
        else
            prevBankId = req.filter.bank_id[i]
            i = i + 1
        end
    end
end

local filterData = function(idList, filter)
    if next(filter) == nil then
        return idList
    end

    local result = {}
    for _, id in pairs(idList) do
        local cpJsonData = redis.call('GET', 'cp:' .. id)
        if cpJsonData then
            local cp = cjson.decode(cpJsonData) or {}
            local matches = true
            for k, v in pairs(filter) do
                local filterType = type(v)
                if filterType == 'table' then
                    -- match any policy for nested array (variants) filters
                    matches = false
                    for _, var in ipairs(v) do
                        if cp[k] == var then
                            matches = true
                            break
                        end
                    end
                else
                    -- match all policy
                    if cp[k] ~= v then
                        matches = false
                    end
                end
                -- match all policy for all filters
                if not matches then
                    break
                end
            end
            if matches then
                result[#result + 1] = id
            end
        end
    end
    return result
end

local idList = redis.call('GEORADIUS', 'cashpoints', req.longitude, req.latitude, req.radius, 'm')
idList = filterData(idList, enabledFilters)

if #idList == 0 then
    return '[]'
end

for i, id in ipairs(idList) do
    idList[i] = tonumber(id)
end

return cjson.encode(idList)
