json = require('json')

local COL_BANK_ID = 1
local COL_BANK_NAME = 2
local COL_BANK_NAME_TR = 3
local COL_BANK_NAME_TR_ALT = 4
local COL_BANK_PARTNERS = 5
local COL_BANK_LICENCE = 7
local COL_BANK_RATING = 8
local COL_BANK_TEL = 9

local MAX_BANKS_BATCH_SIZE = 256

local function _getBankById(bankId)
    local t = box.space.banks.index[0]:select(bankId)
    if #t == 0 then
        return nil
    end

    t = t[1]

    local partnersTuple = t[COL_BANK_PARTNERS]
    local partners = {}
    for i = 1, #partnersTuple do
        partners[#partners + 1] = partnersTuple[i]
    end

    local bank = {
        id = t[COL_BANK_ID],
        name = t[COL_BANK_NAME],
        name_tr = t[COL_BANK_NAME_TR],
        name_tr_alt = t[COL_BANK_NAME_TR_ALT],
        partners = partners,
        licence = t[COL_BANK_LICENCE],
        rating = t[COL_BANK_RATING],
        tel = t[COL_BANK_TEL],
    }

    return bank
end

function getBankById(bankId)
    local bank = _getBankById(bankId)
    if bank then
        return json.encode(bank)
    end

    return ""
end

function getBanksBatch(reqJson)
    local req = json.decode(reqJson)
    if not req or not req.banks then
        box.error{ code = 400, reason = "getBanksBatch: malformed request" }
        return ""
    end

    local result = {}
    for _, bankId in pairs(req.banks) do
        result[#result + 1] = _getBankById(bankId)
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
        result[#result + 1] = tuple[COL_BANK_ID]
    end

    return json.encode(setmetatable(result, { __serialize = "seq" }))
end
