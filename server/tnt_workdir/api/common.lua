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

function malformedRequest(err, func)
    if func then
        func = func .. ": "
    else
        func = ""
    end
    return { code = 400, reason =  func .. err }
end

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

    return _cashpointTupleToTable(t[COL_ID]), t
end

function getCashpointById(cpId)
    local cp = _getCashpointById(cpId)
    if cp then
        return json.encode(cp)
    end
end

function getQuadKey(longitude, latitude, zoom)
    if not longitude or not latitude then
        return ''
    end

    local geoRectPart = function(minLon, maxLon, minLat, maxLat, lon, lat)
        local midLon = (minLon + maxLon) * 0.5
        local midLat = (minLat + maxLat) * 0.5

        local quad = ""
        if lat < midLat then
            maxLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = '0'
            else
                minLon = midLon
                quad = '1'
            end
        else
            minLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = '2'
            else
                minLon = midLon
                quad = '3'
            end
        end

        return minLon, maxLon, minLat, maxLat, quad
    end

    local minLon = -180.0
    local maxLon = 180.0

    local minLat = -90.0
    local maxLat = 90.0

    if longitude < minLon or maxLon < longitude then
        return ""
    end

    if latitude < minLat or maxLat < latitude then
        return ""
    end

    quadKey = ""
    for currentZoom = 0, zoom do
        local q = ""
        minLon, maxLon, minLat, maxLat, q = geoRectPart(minLon, maxLon, minLat, maxLat, longitude, latitude)
        quadKey = quadKey .. q
    end
    return quadKey
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

    local func = "createFilterChain"

    local chain = ""
    if filter.bank_id then
        if type(filter.bank_id) ~= 'table' then
            return "", malformedRequest('filter.bank_id must be an array', func)
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
                return "",  malformedRequest('filter.bank_id contains non-numerical value', func)
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
        return malformedRequest("empty request", func)
    end

    local missingReqired = "missing required request field"
    if not req.topLeft then
        return malformedRequest(missingReqired .. ": topLeft", func)
    end

    if not req.topLeft.longitude then
        return malformedRequest(missingReqired .. ": topLeft.longitude", func)
    end

    if not req.topLeft.latitude then
        return malformedRequest(missingReqired .. ": topLeft.latitude", func)
    end

    if not req.bottomRight then
        return malformedRequest(missingReqired .. ": bottomRight", func)
    end

    if not req.bottomRight.longitude then
        return malformedRequest(missingReqired .. ": bottomRight.longitude", func)
    end

    if not req.bottomRight.latitude then
        return malformedRequest(missingReqired .. ": bottomRight.latitude", func)
    end

    return nil
end

local function isValidCashpointType(cpType)
    local avaliableTypes = { "atm", "office", "branch", "cash" }
    for _, v in ipairs(avaliableTypes) do
        if cpType == v then
            return true
        end
    end
    return false
end

local function isValidCoordinate(longitude, latitude)
    if type(longitude) ~= 'number' or type(latitude) ~= 'number' then
        return false
    end

    if math.abs(latitude) > 90.0 or math.abs(longitude) > 180.0 then
        return false
    end

    return true
end

local function validateCashpointFields(cp, checkRequired, func)
    local allowedFields = {
        id = { type = 'number', required = false },
        type = { type = 'string', required = true },
        bank_id = { type = 'number', required = true },
        town_id = { type = 'number', required = true },
        longitude = { type = 'number', required = true },
        latitude = { type = 'number', required = true },
        address = { type = 'string', required = false },
        address_comment = { type = 'string', required = false },
        metro_name = { type = 'string', required = false },
        free_access = { type = 'boolean', required = true },
        main_office = { type = 'boolean', required = true },
        without_weekend = { type = 'boolean', required = true },
        round_the_clock = { type = 'boolean', required = true },
        works_as_shop = { type = 'boolean', required = true },
        schedule = { type = 'string', required = false },
        tel = { type = 'string', required = false },
        additional = { type = 'string', required = false },
        rub = { type = 'boolean', required = true },
        usd = { type = 'boolean', required = true },
        eur = { type = 'boolean', required = true },
        cash_in = { type = 'boolean', required = true },
        version = { type = 'number', required = false },
        timestamp = { type = 'number', required = false },
    }

    if checkRequired then
        for k, v in pairs(allowedFields) do
            if v.required and cp[k] == nil then
                return malformedRequest("missing required cashpoint field '" .. tostring(k) .. "'")
            end
            if cp[k] ~= nil then
                local cpKType = type(cp[k])
                if cpKType ~= v.type then
                    return malformedRequest("wrong type of cashpoint field '" .. tostring(k) .. "'. " ..
                                            "Expected '" .. v.type .. "' but got '" .. cpKType .. "'", func)
                end
            end
        end
    end

    for k, v in pairs(cp) do
        if allowedFields[k] == nil then -- unknown field
            return malformedRequest("unknown cashpoint field '" .. tostring(k) .. "'", func)
        end
        local fieldType = type(v)
        local expectedType = allowedFields[k].type
        if fieldType ~= expectedType then -- type missmatch
            return malformedRequest("wrong type of cashpoint field '" .. tostring(k) .. "'. Expected '" .. expectedType .. "' but got '" .. fieldType .. "'", func)
        end
    end
end

function validateCashpoint(cp, checkRequired, func)
    local err = validateCashpointFields(cp, checkRequired, func)
    if err then
        return err
    end

    if cp.bank_id then
        local t = box.space.banks.index[0]:select{ cp.bank_id }
        if #t == 0 then
            return malformedRequest("no such bank_id: " .. tostring(cp.bank_id), func)
        end
    end

    if cp.town_id then
        local t = box.space.towns.index[0]:select{ cp.town_id }
        if #t == 0 then
            return malformedRequest("no such town_id: " .. tostring(cp.town_id), func)
        end
    end

    if cp.type then
        if not isValidCashpointType(cp.type) then
            return malformedRequest("wrong cashpoint type '" .. tostring(cp.type) .. "'", func)
        end
    end

    if checkRequired or (cp.longitude and cp.latitude) then
        if not isValidCoordinate(cp.longitude, cp.latitude) then
            return malformedRequest("invalid cashpoint coordinate (" .. tostring(cp.longitude) .. ", " .. tostring(cp.latitude) .. ")", func)
        end
    end

    return nil
end
