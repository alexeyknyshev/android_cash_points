local reqPayload = ARGV[1]
local countLimit = ARGV[2] or 32

if not reqPayload then
    return redis.error_reply('no such json payload')
end

local req = cjson.decode(reqPayload)

if not req.longitude then
    return redis.error_reply('no such required argument: longitude')
end

if not req.latitude then
    return redis.error_reply('no such required argument: latitude')
end

if not req.radius then
    return redis.error_reply('no such required argument: radius')
end

if req.radius <= 0.0 then
    return redis.error_reply('search radius must be positive')
end

req.radius = math.floor(req.radius) + 1

req.filter = req.filter or {}
--req.filter = { bankId = 322 }

countLimit = tonumber(countLimit)

if countLimit <= 0 then
    return redis.error_reply('towns count limit must be positive')
end

local result = {}

local townIdList = redis.call('GEORADIUS', 'towns', req.longitude, req.latitude, req.radius, 'm')
for _, townId in pairs(townIdList) do
    local townJsonData = redis.call('GET', 'town:' .. tostring(townId))
    if townJsonData then
        local town = cjson.decode(townJsonData)

        if town then
            local townIdStr = 'town:' .. tostring(townId)
            local cpCount = 0
            if req.filter.bankId then
                cpCount = redis.call('SINTERSTORE', townIdStr .. ':bank:' .. req.filter.bankId .. ':cp',
                                     townIdStr .. ':cp', 'bank:' .. req.filter.bankId .. ':cp')
            else
                cpCount = redis.call('SCARD', townIdStr .. ':cp')
            end

            local clusterData = {
                longitude = town.longitude,
                latitude = town.latitude,
                size = cpCount,
                id = townId
            }

            result[#result + 1] = clusterData
        end
    end
end

if #result == 0 then
    return '[]'
end

table.sort(result, function(a, b) return a.size > b.size end)

while #result > countLimit * 4 do
    table.remove(result)
end
--[[
local dist = function(a, b)
    local r = 63781370
    local dlon = (a.longitude - b.longitude) * math.pi / 180.0
    local dlat = (a.latitude - b.latitude) * math.pi / 180.0
    local sdlat = math.sin(dlat * 0.5)
    local sdlon = math.sin(dlon * 0.5)
    local a = sdlat * sdlat + math.cos(a.longitude * math.pi / 180.0) * math.cos(b.longitude * math.pi / 180.0) * sdlon * sdlon
    local c = 2 * math.atan2(math.sqrt(a), math.sqrt(1.0 - a))
    return c * r;
end
]]--
local checkDist = function(a, b, minDist)
    local abDist = redis.call('GEODIST', 'towns', a.id, b.id, 'm')
    if abDist then
        abDist = tonumber(abDist)
        if abDist < minDist then
            if a.size < b.size then
                return 1
            end
            return 2
        end
    end
    return 0
end

local minDist = req.radius * 0.1
local minPointsCount = (result[2] or result[1]).size * 0.01

for i = #result, 1, -1 do
    if result[i].size < minPointsCount then
        table.remove(result, i)
    end
end

local i = 1
while i < #result do
    local j = i + 1
    while j <= #result do
        local mark = checkDist(result[i], result[j], minDist)
        if mark == 1 then
            result[j].size = result[i].size + result[j].size
            table.remove(result, i)
            i = i - 1
        elseif mark == 2 then
            result[i].size = result[i].size + result[j].size
            table.remove(result, j)
            j = j - 1
        end
        j = j + 1
    end
    i = i + 1
end

while #result > countLimit do
    table.remove(result)
end

return cjson.encode(result)
