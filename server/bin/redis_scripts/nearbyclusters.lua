local reqPayload = ARGV[1]

local getQuadKey = function(longitude, latitude, zoom)
    local geoRectPart = function(minLon, maxLon, minLat, maxLat, lon, lat)
        local midLon = (minLon + maxLon) * 0.5
        local midLat = (minLat + maxLat) * 0.5

        local quad = ""
        if lat < midLat then
            maxLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = "0"
            else
                minLon = midLon
                quad = "1"
            end
        else
            minLat = midLat
            if lon < midLon then
                maxLon = midLon
                quad = "2"
            else
                minLon = midLon
                quad = "3"
            end
        end

        return minLon, maxLon, minLat, maxLat, quad
    end

    local minLon = -180.0
    local maxLon = 180.0

    local minLat = -85.0
    local maxLat = 85.0

    quadKey = ""
    for currentZoom = 0, zoom do
        local q = ""
        minLon, maxLon, minLat, maxLat, q = geoRectPart(minLon, maxLon, minLat, maxLat, longitude, latitude)
        quadKey = quadKey .. q
    end
    return quadKey
end

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

if not req.zoom then
    return redis.error_reply('no such required argument: zoom')
end

if req.zoom < 0 then
    return redis.error_reply('zoom must be positive')
end

if req.zoom > 16 then
    return redis.error_reply('zoom is too large')
end

req.zoom = math.floor(req.zoom + 0.5)
local quadKey = getQuadKey(req.longitude, req.latitude, req.zoom)

return quadKey
