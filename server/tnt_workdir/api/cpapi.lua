json = require('json')
local common = require('common')

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

local MAX_CASHPOINTS_BATCH_SIZE = 1024
local MAX_COORD_DELTA = 0.01

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

function getNearbyCashpoints(reqJson)
    local req = json.decode(reqJson)

    local err = validateRequest(req, "getNearbyCashpoints")
    if err then
        box.error(err)
        return nil
    end

    req.filter = req.filter or {}

    local t = box.space.cashpoints.index[1]:select({ req.topLeft.longitude, req.topLeft.latitude,
                                                     req.bottomRight.longitude, req.bottomRight.latitude },
                                                   { iterator = "le" })

    local tooBigRegion = "too big region size in request"
    if math.abs(req.topLeft.longitude - req.bottomRight.longitude) > MAX_COORD_DELTA then
        box.error{ code = 400, reason = func .. ": " .. tooBigRegion .. ": longitude" }
        return nil
    end

    if math.abs(req.topLeft.latitude - req.bottomRight.latitude) > MAX_COORD_DELTA then
        box.error{ code = 400, reason = func .. ": " .. tooBigRegion .. ": latitude" }
        return nil
    end

    local filtersList = {
        matchingBankFilter,
        matchingTypeFilter,
        matchingRubFilter,
        matchingUsdFilter,
        matchingEurFilter,
        matchingRoundTheClock,
        matchingWithoutWeekend,
        matchingFreeAccess,
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
