json = require('json')

local MAX_CASHPOINTS_BATCH_SIZE = 1024
local MAX_COORD_DELTA = 0.01

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

local function _getCashpointById(cpId)
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


function getCashpointsBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.cashpoints then
        box.error{ code = 400, reason = "getCashpointsBatch: malformed request" }
        return nil
    end

    local result = {}
    for _, cpId in pairs(req.cashpoints) do
        local cp = _getCashpointById(cpId)
        if cp then
            result[#result + 1] = cp
        end
        if #result == MAX_CASHPOINTS_BATCH_SIZE then
            break
        end
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end

local function _matchingBankFilter(tuple, filter)
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

local function _matchingTypeFilter(tuple, filter)
    if not filter.type then
        return true
    end

    return tuple[COL_TYPE] == filter.type
end

local function _matchingFreeAccess(tuple, filter)
    if filter.free_access ~= nil then
        return tuple[COL_FREE_ACCESS] == filter.free_access
    end
    return true
end

local function _matchingRubFilter(tuple, filter)
    if filter.rub ~= nil then
        return tuple[COL_RUB] == filter.rub
    end
    return true
end

local function _matchingUsdFilter(tuple, filter)
    if filter.usd ~= nil then
        return tuple[COL_USD] == filter.usd
    end
    return true
end

local function _matchingEurFilter(tuple, filter)
    if filter.eur ~= nil then
        return tuple[COL_EUR] == filter.eur
    end
    return true
end

local function _matchingRoundTheClock(tuple, filter)
    if filter.round_the_clock ~= nil then
        return tuple[COL_ROUND_THE_CLOCK] == filter.round_the_clock
    end
    return true
end

local function _matchingWithoutWeekend(tuple, filter)
    if filter.without_weekend ~= nil then
        return tuple[COL_WITHOUT_WEEKEND] == filter.without_weekend
    end
    return true
end

function getNearbyCashpoints(reqJson)
    local func = "getNearbyCashpoints"
    local req = json.decode(reqJson)
    if not req then
        box.error{ code = 400, reason = func .. ": malformed request" }
        return nil
    end

    req.filter = req.filter or {}

    local missingReqired = "missing required request field"
    if not req.topLeft then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": topLeft" }
        return nil
    end

    if not req.topLeft.longitude then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": topLeft.longitude" }
        return nil
    end

    if not req.topLeft.latitude then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": topLeft.latitude" }
        return nil
    end

    if not req.bottomRight then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight" }
        return nil
    end

    if not req.bottomRight.longitude then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight.longitude" }
        return nil
    end

    if not req.bottomRight.latitude then
        box.error{ code = 400, reason = func .. ": " .. missingReqired .. ": bottomRight.latitude" }
        return nil
    end

    local tooBigRegion = "too big region size in request"
    if math.abs(req.topLeft.longitude - req.bottomRight.longitude) > MAX_COORD_DELTA then
        box.error{ code = 400, reason = func .. ": " .. tooBigRegion .. ": longitude" }
        return nil
    end

    if math.abs(req.topLeft.latitude - req.bottomRight.latitude) > MAX_COORD_DELTA then
        box.error{ code = 400, reason = func .. ": " .. tooBigRegion .. ": latitude" }
        return nil
    end

    local t = box.space.cashpoints.index[1]:select({ req.topLeft.longitude, req.topLeft.latitude,
                                                     req.bottomRight.longitude, req.bottomRight.latitude },
                                                   { iterator = "le" })

    local filtersList = {
        _matchingBankFilter,
        _matchingTypeFilter,
        _matchingRubFilter,
        _matchingUsdFilter,
        _matchingEurFilter,
        _matchingRoundTheClock,
        _matchingWithoutWeekend,
        _matchingFreeAccess,
    } 

    local result = {}

    for _, tuple in pairs(t) do
        local matching = true
        if tuple then
            for _, filter in ipairs(filtersList) do
                matching = filter(tuple, req.filter)
                if not matching then
                    break
                end
            end
        else
            matching = false
        end

        if matching then
            result[#result + 1] = tuple[COL_ID]
        end
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
