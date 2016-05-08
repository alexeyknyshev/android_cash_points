local COL_METRO_ID = 1
local COL_METRO_COORD = 2
local COL_METRO_TOWN_ID = 3
local COL_METRO_BRANCH_ID = 4
local COL_METRO_STATION_NAME = 5
local COL_METRO_STATION_EXIT_NAME = 6

local MAX_METRO_BATCH_SIZE = 1024

json = require("json")

function _getMetroById(metroId)
    local t = box.space.metro.index[0]:select(metroId)
    if #t == 0 then
        return nil
    end
    local tuple = t[1]
    local metro = {
        id = tuple[COL_METRO_ID],
        longitude = tuple[COL_METRO_COORD][1],
        latitude = tuple[COL_METRO_COORD][2],
        town_id = tuple[COL_METRO_TOWN_ID],
        branch_id = tuple[COL_METRO_BRANCH_ID],
        station_name = tuple[COL_METRO_STATION_NAME],
        station_exit_name = tuple[COL_METRO_STATION_EXIT_NAME],
    }

    return metro
end

function getMetroList(townId)
    local t = box.space.metro.index[2]:select(townId)

    local result = {}
    for _, tuple in ipairs(t) do
        result[#result + 1] = tuple[COL_METRO_ID]
    end
    return json.encode(setmetatable(result, {__serialize = "seq"}))
end

function getMetroById(metroId)
    local metro = _getMetroById(metroId)
    if metro then
        return json.encode(metro)
    end
    return ""
end

function getMetroBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.metro then
        box.error{ code = 400, reason = "getMetroBatch: malformed request" }
        return nil
    end

    local result = {}
    for _, metroId in pairs(req.metro) do
        local tuple = _getMetroById(metroId)
        if tuple then
            result[#result + 1] = tuple
        end
        if #result == MAX_METRO_BATCH_SIZE then
            break
        end
    end
    return json.encode(setmetatable(result, {__serialize = "seq"}))
end
