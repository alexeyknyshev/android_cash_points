json = require('json')

local COL_TOWN_ID = 1
local COL_TOWN_COORD = 2
local COL_TOWN_NAME = 3
local COL_TOWN_NAME_TR = 4
local COL_TOWN_REGION_ID = 5
local COL_TOWN_REGIONAL_CENTER = 6
local COL_TOWN_ZOOM = 7
local COL_TOWN_BIG = 8
local COL_TOWN_CP_COUNT = 9

local MAX_TOWNS_BATCH_SIZE = 1024

local function _getTownById(townId)
    local t = box.space.towns.index[0]:select(townId)
    if #t == 0 then
        return nil
    end

    t = t[1]

    local town = {
        id = t[COL_TOWN_ID],
        longitude = t[COL_TOWN_COORD][1],
        latitude = t[COL_TOWN_COORD][2],
        name = t[COL_TOWN_NAME],
        name_tr = t[COL_TOWN_NAME_TR],
        region_id = t[COL_TOWN_REGION_ID],
        regional_center = t[COL_TOWN_REGIONAL_CENTER],
        zoom = t[COL_TOWN_ZOOM],
        big = t[COL_TOWN_BIG],
    }

    return town
end

function getTownById(townId)
    local town = _getTownById(townId)
    if town then
        return json.encode(town)
    end

    return ""
end

function getTownsBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.towns then
        box.error{ code = 400, reason = "getTownsBatch: malformed request" }
        return nil
    end

    local result = {}
    for _, townId in pairs(req.towns) do
        result[#result + 1] = _getTownById(townId)
        if #result == MAX_TOWNS_BATCH_SIZE then
            break
        end
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end

function getTownsList()
    local t = box.space.towns.index[0]:select{}

    local result = {}
    for _, tuple in ipairs(t) do
        result[#result + 1] = tuple[COL_TOWN_ID]
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
