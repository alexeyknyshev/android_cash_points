json = require('json')
local common = require('common')

local CP_ID = 1
local CP_COORD = 2

local CLUSTER_ID = 1
local CLUSTER_COORD = 2
local CLUSTER_MEMBERS = 3
local CLUSTER_SIZE = 4

local TOWN_ID = 1
local TOWN_COORD = 2
local TOWN_CASHPOINTS_COUNT = 9

local CLUSTER_ZOOM_MIN = 10
local CLUSTER_ZOOM_MAX = 16

function getNearbyClusters(reqJson, countLimit)
    local fund = "getNearbyClusters"
    local req = json.decode(reqJson)

    local err = validateRequest(req, func)
    if err then
        box.error(err)
        return nil
    end

    req.filter = req.filter or {}

    if not req.zoom then
        box.error{ code = 400, reason = func .. ": missing required argument => req.zoom"}
        return nil
    end

    if req.zoom < CLUSTER_ZOOM_MIN then
        return _getNearbyTownClusters(req, countLimit)
    else
        return _getNearbyQuadClusters(req)
    end
end

function _getNearbyQuadClusters(req)
    local t = box.space.clusters.index[1]:select({ req.topLeft.longitude, req.topLeft.latitude,
                                                   req.bottomRight.longitude, req.bottomRight.latitude },
                                                 { iterator = "le" })
    local filtersList = {
        matchingBankFilter,
        matchingTypeFilter,
        matchingRubFilter,
        matchingUsdFilter,
        matchingEurFilter,
        matchingRoundTheClock,
        matchingWithoutWeekend,
        matchingFreeAccess,
        matchingApproved,
    }

    local result = {}

    for _, tuple in pairs(t) do
        local clusterId = tuple[CLUSTER_ID]
        if clusterId:len() == req.zoom + 1 then
            local cluster = {
                id = clusterId,
                longitude = tuple[CLUSTER_COORD][1],
                latitude = tuple[CLUSTER_COORD][2],
                size = tuple[CLUSTER_SIZE],
            }

            local lastCpId = nil
            if next(req.filter) ~= nil then
                cluster.size = 0
                cluster.longitude = 0.0
                cluster.latitude = 0.0

                --print(json.encode(tuple[CLUSTER_MEMBERS]))

                for _, cpId in pairs(tuple[CLUSTER_MEMBERS]) do
                    lastCpId = nil
                    local cpTupleList = box.space.cashpoints.index[0]:select{ cpId }
                    if #cpTupleList > 0 then
                        local cpTuple = cpTupleList[1]
                        local matching = true
                        for _, filter in ipairs(filtersList) do
                            matching = filter(cpTuple, req.filter)
                            if not matching then
                                break
                            end
                        end
                        if matching then
                            lastCpId = cpId
                            cluster.size = cluster.size + 1
                            cluster.longitude = cluster.longitude + cpTuple[CP_COORD][1]
                            cluster.latitude = cluster.latitude + cpTuple[CP_COORD][2]
                        end
                    end
                end

                if cluster.size > 0 then
                    cluster.longitude = cluster.longitude / cluster.size
                    cluster.latitude = cluster.latitude / cluster.size
                end
            end

            if cluster.size > 0 then
                if cluster.size == 1 and lastCpId then
                    result[#result + 1] = _getCashpointById(lastCpId)
                else
                    result[#result + 1] = cluster
                end
            end
        end

--         local chain = createFilterChain(req.filter)
--         if chain:len() > 0 then
--
--         end
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end

function _getNearbyTownClusters(req, countLimit)
    local countLimit = countLimit or 32

    local t = box.space.towns.index[1]:select({ req.topLeft.longitude, req.topLeft.latitude,
                                                req.bottomRight.longitude, req.bottomRight.latitude },
                                              { iterator = "le" })
    local filtersList = {
    }

    local result = {}

    for _, tuple in pairs(t) do
        local townId = tuple[TOWN_ID]
        result[#result + 1] = {
            id = townId,
            longitude = tuple[TOWN_COORD][1],
            latitude = tuple[TOWN_COORD][2],
            size = tuple[TOWN_CASHPOINTS_COUNT],
        }
    end

    table.sort(result, function(a, b) return a.size > b.size end)
    local deltaLon = math.abs(req.topLeft.longitude - req.bottomRight.longitude)
    local deltaLat = math.abs(req.topLeft.latitude - req.bottomRight.latitude)

    local minDist = math.min(deltaLon, deltaLat) * 0.2
    local minDistSqr = minDist * minDist

    local getDistSqr = function(a, b)
        local deltaLon = a.longitude - b.longitude
        local deltaLat = a.latitude - b.latitude
        return deltaLon * deltaLon + deltaLat * deltaLat
    end

    for i = 1, #result do
        if result[i].size ~= 0 then
            for j = #result, i + 1, -1 do
                if result[j].size ~= 0 then
                    local dist = getDistSqr(result[i], result[j])
                    if dist < minDist then
                        result[i].size = result[i].size + result[j].size
                        result[j].size = 0
                    end
                end
            end
        end
    end

    for i = #result, 1, -1 do
        if result[i].size == 0 then
            table.remove(result, i)
        end
    end

    while #result > countLimit do
        table.remove(result)
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
