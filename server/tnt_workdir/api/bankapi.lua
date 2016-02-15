json = require('json')

local MAX_BANKS_BATCH_SIZE = 256

local function _getBankById(bankId)
    local t = box.space.banks.index[0]:select(bankId)
    if #t == 0 then
        return nil
    end

    t = t[1]

    local partnersTuple = t[5]
    local partners = {}
    for i = 1, #partnersTuple do
        partners[#partners + 1] = partnersTuple[i]
    end

    local bank = {
        id = t[1],
        name = t[2],
        name_tr = t[3],
        name_tr_alt = t[4],
        partners = partners,
        licence = t[7],
        rating = t[8],
        tel = t[9],
    }

    return bank
end

function getBankById(bankId)
    local bank = _getBankById(bankId)
    if bank then
        return json.encode(bank)
    end
end

function getBanksBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.banks then
        box.error{ code = 400, reason = "getBanksBatch: malformed request" }
        return nil
    end

    local result = {}
    for _, bankId in pairs(req.banks) do
        local bank = _getBankById(bankId)
        if bank then
            result[#result + 1] = bank
        end
        if #result == MAX_BANKS_BATCH_SIZE then
            break
        end
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end

function getBanksList()
    local t = box.space.banks.index[0]:select{}

    local result = {}
    for _, tuple in ipairs(t) do
        result[#result + 1] = tuple[1]
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
