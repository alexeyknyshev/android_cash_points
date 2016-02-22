json = require('json')
local fiber = require('fiber')
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

local COL_TOWN_CP_COUNT = 9

local COL_CP_PATCH_ID = 1
local COL_CP_PATCH_CASHPOINT_ID = 2
local COL_CP_PATCH_CASHPOINT_USER_ID = 3
local COL_CP_PATCH_DATA = 4
--local COL_CP_PATCH_TIMESTAMP = 5

local COL_CP_PATCH_VOTE_ID = 1
local COL_CP_PATCH_VOTE_PATCH_ID = 2
local COL_CP_PATCH_VOTE_USER_ID = 3
local COL_CP_PATCH_VOTE_DATA = 4
--local COL_CP_PATCH_VOTE_TIMESTAMP = 5

local PATCH_APPROVE_VOTES = 5

local MAX_CASHPOINTS_BATCH_SIZE = 1024
local MAX_COORD_DELTA = 0.01

local INT32_MAX = 2147483647

function getCashpointsBatch(reqJson)
    local func = "getCashpointsBatch"
    local req = json.decode(reqJson)
    if not req or not req.cashpoints then
        box.error(malformedRequest("malformed request json", func))
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
    local func = "getNearbyCashpoints"
    local req = json.decode(reqJson)

    local err = validateRequest(req, func)
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
        box.error(malformedRequest(tooBigRegion .. ": longitude", func))
        return nil
    end

    if math.abs(req.topLeft.latitude - req.bottomRight.latitude) > MAX_COORD_DELTA then
        box.error(malformedRequest(tooBigRegion .. ": latitude", func))
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

local function updateOldCp(old, new)
    local dataChanged = false
    for k, v in pairs(new) do
        if old[k] ~= nil and old[k] ~= v then
            old[k] = v
            dataChanged = true
        end
    end
    return old, dataChanged
end

function cashpointCommit(reqJson)
    local func = "cashpointCommit"

    local cp = json.decode(reqJson)

    if cp.id then -- editing existing cashpoint
        local err = validateCashpoint(cp, false, func)
        if err then
            box.error(err)
            return false
        end

        local oldCp, t = _getCashpointById(cp.id)
        if not oldCp then
            box.error(malformedRequest("attempt to edit non existing cashpoint with id: " .. tostring(cp.id)), func)
            return false
        end

        if cp.longitude and cp.latitude then -- coordinates changed
            local oldQuadKey = getQuadKey(oldCp.longitude, oldCp.latitude, 15)
            local newQuadKey = getQuadKey(cp.longitude, cp.latitude, 15)

            if newQuadKey:len() == 0 then
                box.error(malformedRequest("invalid coordinates of cashpoint: (" .. tostring(cp.longitude) .. ", " .. tostring(cp.latitude) .. ")"), func)
                return false
            end

            if oldQuadKey ~= newQuadKey then -- quadkey changed
                -- TODO: update quadkeys
                box.error{ code = 404, reason = "quadkey update is not implemented yet!" }
                return false
            end
        end

        if cp.town_id ~= oldCp.town_id then -- update towns' cp count
            box.space.towns:update(oldCp.town_id, {{ '-',  COL_TOWN_CP_COUNT, 1 }})
            box.space.towns:update(cp.town_id, {{ '+', COL_TOWN_CP_COUNT, 1 }})
        end

        local updated = false
        cp, updated = updateOldCp(oldCp, cp)
        if not updated then
            return false
        end
    else -- creating new cashpoint
        local err = validateCashpoint(cp, true, func)
        if err then
            box.error(err)
            return false
        end

        local quadkey = getQuadKey(cp.longitude, cp.latitude, 15)
        if quadkey:len() == 0 then
            box.error(malformedRequest("invalid coordinates of cashpoint: (" .. tostring(cp.longitude) .. ", " .. tostring(cp.latitude) .. ")"), func)
            return false
        end

        -- TODO: update quadtree

        box.space.towns:update(cp.town_id, {{ '+', COL_TOWN_CP_COUNT, 1 }})
        -- TODO: creating new cashpoint
    end

    return true
end

function cashpointProposePatch(reqJson)
    local func = "cashpointProposePatch"
    local timestamp = fiber.time64()

    local req = json.decode(reqJson)

    local userId = req.user_id
    if not userId then
        return false
    end
    -- TODO: check session exists

    local cp = req.data
    if not cp then
        return false
    end

    local cpId = INT32_MAX

    if cp.id then -- editing existing cashpoint
        local err = validateCashpoint(cp, false, func)
        if err then
            return false
        end

        cpId = cp.id
        cp.id = nil -- don't save cashpoint id in patch

        local oldCp, t = _getCashpointById(cp.id)
        if not oldCp then
            return false
        end

        local updated = false
        updatedCp, updated = updateOldCp(oldCp, cp)
        if not updated then
            return false
        end
    else -- creating new cashpoint
        local err = validateCashpoint(cp, true, func)
        if err then
            return false
        end
    end

    local t = box.space.cashpoints_patches.index[1]:select{ cpId }

    reqJson = json.encode(cp)
    for _, tuple in pairs(t) do
        if reqJson == tuple[COL_CP_PATCH_DATA] then -- same patch already exists
            return false
        end
    end

--     print('before auto_increment: ' .. tostring(cpId) .. ', type = ' .. type(cpId))
--     if cp then
--         return true
--     end
    box.space.cashpoints_patches:auto_increment{ cpId, userId, reqJson, timestamp }
    return true
end

function getCashpointPatches(cpId)
    if not cpId then
        return nil
    end

    local t = box.space.cashpoints_patches.index[1]:select{ cpId }

    local result = {}
    for _, tuple in pairs(t) do
        local patchId = tuple[1]
        local cpPatchJson = tuple[3]
        local cp = json.decode(cpPatchJson)
        if cp then
            result[patchId] = cp
        end
    end

    return json.encode(setmetatable(result, { __serialize = "map" }))
end

function _getCashpointPatchVotes(patchId)
    if not patchId then
        return nil
    end

    local t = box.space.cashpoints_patches_votes.index[1]:select{ patchId }

    local result = {}
    for _, tuple in pairs(t) do
        local patchVote = {
            user_id = tuple[COL_CP_PATCH_VOTE_USER_ID],
            score = tuple[COL_CP_PATCH_VOTE_DATA],
            timestamp = tuple[COL_CP_PATCH_VOTE_TIMESTAMP],
        }
        result[#result + 1] = patchVote
    end
    return result
end

function getCashpointPatchVotes(patchId)
    local result = _getCashpointPatchVotes(patchId)
    if result then
        return json.encode(setmetatable(result, { __serialize = "seq" }))
    end
end

local function validateVote(userId, score, func)
    if not userId then
        return malformedRequest("missing user_id for vote", func)
    end
    -- TODO: vote user_id validation

    if not score then
        return malformedRequest("missing vote score", func)
    end

    if math.abs(score) ~= 1 then
        return malformedRequest("wrong vote score", func)
    end
end

local function isCashpointPatchApproved(patchId)
    local votes = _getCashpointPatchVotes(patchId)
    if votes then
        local score = 0
        for _, vote in ipairs(votes) do
            score = score + vote.score
        end
        return score >= PATCH_APPROVE_VOTES
    end
    return false
end

-- vote struct:
--    patch_id
--    user_id
--    vote
function cashpointVotePatch(reqJson)
    local func = "cashpointVotePatch"
    local vote = json.decode(reqJson)
    if type(vote.patch_id) ~= 'number' then
        box.error(malformedRequest("missing patch id for vote", func))
        return false
    end

    local t = box.space.cashpoints_patches.index[0]:select{ vote.patch_id }
    if #t == 0 then
        box.error(malformedRequest("no such patch id for vote", func))
        return false
    end

    local patchTuple = t[1]

    local err = validateVote(vote.user_id, vote.score)
    if err then
        box.error(err)
        return false
    end

    box.space.cashpoints_patches_votes:auto_increment{ vote.patch_id, vote.user_id, vote.score }
    if isCashpointPatchApproved(vote.patch_id) then
        if not cashpointCommit(patchTuple[COL_CP_PATCH_DATA]) then
            box.error{ code = 400, reason = "cannot commit approved cashpoint patch" }
            return false
        end
    end
    return true
end
