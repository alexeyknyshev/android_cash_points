local json = require('json')

function _getSpaceMetrics()
    return {
        banks = box.space.banks.index[0]:count(),
        towns = box.space.towns.index[0]:count(),
        regions = box.space.regions.index[0]:count(),
        metro = box.space.metro.index[0]:count(),
        cashpoints = box.space.cashpoints.index[0]:count(),
        cashpoints_patches = box.space.cashpoints_patches.index[0]:count(),
        cashpoints_patches_votes = box.space.cashpoints_patches_votes.index[0]:count(),
        clusters = box.space.clusters.index[0]:count(),
        clusters_cache = box.space.clusters_cache.index[0]:count(),
    }
end

function getSpaceMetrics()
    local metrcis = _getSpaceMetrics()
    return json.encode(setmetatable(metrcis, { __serialize = "map" }))
end
