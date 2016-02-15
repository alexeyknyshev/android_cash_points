json = require('json')

local MAX_TOWNS_BATCH_SIZE = 1024

local function _getTownById(townId)
    local t = box.space.towns.index[0]:select(townId)
    if #t == 0 then
        return nil
    end

    t = t[1]

    local town = {
        id = t[1],
        longitude = t[2][1],
        latitude = t[2][2],
        name = t[3],
        name_tr = t[4],
        region_id = t[5],
        regional_center = t[6],
        zoom = t[7],
        big = t[8],
    }

    return town
end

function getTownById(townId)
    local town = _getTownById(townId)
    if town then
        return json.encode(town)
    end
end

function getTownsBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.towns then
        box.error{ code = 400, reason = "getTownsBatch: malformed request" }
        return nil
    end

    local result = {}
    for _, townId in pairs(req.towns) do
        local town = _getTownById(townId)
        if town then
            result[#result + 1] = town
        end
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
        result[#result + 1] = tuple[1]
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
