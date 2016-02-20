local COL_ID = 1
local COL_COORD = 2
local COL_TYPE = 3
local COL_BANK_ID = 4
local COL_TOWN_ID = 5
local COL_ADDRESS = 6
local COL_ADDRESS_COMMENT = 7
local COL_METRO_NAME = 8
local COL_FREE_ACCESS = 9
local COL_MAIN_OFFICE = 10
local COL_WITHOUT_WEEKEND = 11
local COL_ROUND_THE_CLOCK = 12
local COL_WORKS_AS_SHOP = 13
local COL_SCHEDULE = 14
local COL_TEL = 15
local COL_ADDITIONAL = 16
local COL_RUB = 17
local COL_USD = 18
local COL_EUR = 19
local COL_CASH_IN = 20
local COL_VERSION = 21


local function _cashpointTupleToTable(t)
    local cp = {
        id = t[COL_ID],
        longitude = t[COL_COORD][1],
        latitude = t[COL_COORD][2],
        type = t[COL_TYPE],
        bank_id = t[COL_BANK_ID],
        town_id = t[COL_TOWN_ID],
        address = t[COL_ADDRESS],
        address_comment = t[COL_ADDRESS_COMMENT],
        metro_name = t[COL_METRO_NAME],
        free_access = t[COL_FREE_ACCESS],
        main_office = t[COL_MAIN_OFFICE],
        without_weekend = t[COL_WITHOUT_WEEKEND],
        round_the_clock = t[COL_ROUND_THE_CLOCK],
        works_as_shop = t[COL_WORKS_AS_SHOP],
        schedule = t[COL_SCHEDULE],
        tel = t[COL_TEL],
        additional = t[COL_ADDITIONAL],
        rub = t[COL_RUB],
        usd = t[COL_USD],
        eur = t[COL_EUR],
        cash_in = t[COL_CASH_IN],
        version = t[COL_VERSION],
-- TODO: timestamp
    }

    return cp
end

function _getCashpointById(cpId)
    local t = box.space.cashpoints.index[0]:select(cpId)
    if #t == 0 then
        return nil
    end

    return _cashpointTupleToTable(t[COL_ID])
end

function getCashpointById(cpId)
    local cp = _getCashpointById(cpId)
    if cp then
        return json.encode(cp)
    end
end


function getSupportedFilters()
    return {
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
end

function getSupportedFiltersOrder()
    return {
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
end

function createFilterChain(filter)
    if next(filter) == nil then
        return ""
    end

    local chain = ""
    if filter.bank_id then
        if type(filter.bank_id) ~= 'table' then
            return "", { code = 400, reason = 'filter.bank_id must be an array' }
        end

        -- sort to order ids => prevent bankIdChain variations
        table.sort(filter.bank_id, function(a, b) return a < b end)

        -- remove duplicates
        local prevBankId = 0
        local i = 1
        while i <= #filter.bank_id do
            if filter.bank_id[i] == prevBankId then
                table.remove(filter.bank_id, i)
            else
                prevBankId = filter.bank_id[i]
                i = i + 1
            end
        end

        for _, bankId in ipairs(filter.bank_id) do
            if type(bankId) == 'number' then
                chain = ':' .. tostring(math.floor(bankId))
            else
                return "", { code = 400, reason = 'filter.bank_id contains non-numerical value' }
            end
        end

        if string.len(chain) > 0 then
            chain = ':bank' .. chain
        end
    end

    for i, field in ipairs(getSupportedFiltersOrder()) do
        if field ~= 'bank_id' and filter[field] ~= nil then
            chain = chain .. ':' .. field .. ':' .. tostring(filter[field])
        end
    end

    return chain
end

function matchingBankFilter(tuple, filter)
    if not filter.bank_id then
        return true
    end

    for _, id in ipairs(filter.bank_id) do
        if tuple[COL_BANK_ID] == id then
            return true
        end
    end

    return false
end

function matchingTypeFilter(tuple, filter)
    if filter.type ~= nil then
        return tuple[COL_TYPE] == filter.type
    end
    return true
end

function matchingFreeAccess(tuple, filter)
    if filter.free_access ~= nil then
        return tuple[COL_FREE_ACCESS] == filter.free_access
    end
    return true
end

function matchingRubFilter(tuple, filter)
    if filter.rub ~= nil then
        return tuple[COL_RUB] == filter.rub
    end
    return true
end

function matchingUsdFilter(tuple, filter)
    if filter.usd ~= nil then
        return tuple[COL_USD] == filter.usd
    end
    return true
end

function matchingEurFilter(tuple, filter)
    if filter.eur ~= nil then
        return tuple[COL_EUR] == filter.eur
    end
    return true
end

function matchingRoundTheClock(tuple, filter)
    if filter.round_the_clock ~= nil then
        return tuple[COL_ROUND_THE_CLOCK] == filter.round_the_clock
    end
    return true
end

function matchingWithoutWeekend(tuple, filter)
    if filter.without_weekend ~= nil then
        return tuple[COL_WITHOUT_WEEKEND] == filter.without_weekend
    end
    return true
end


function validateRequest(req, func)
    if not req then
        return { code = 400, reason = func .. ": malformed request" }
    end

    local missingReqired = "missing required request field"
    if not req.topLeft then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": topLeft" }
    end

    if not req.topLeft.longitude then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": topLeft.longitude" }
    end

    if not req.topLeft.latitude then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": topLeft.latitude" }
    end

    if not req.bottomRight then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight" }
    end

    if not req.bottomRight.longitude then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight.longitude" }
    end

    if not req.bottomRight.latitude then
        return { code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight.latitude" }
    end

    return nil
end
