box.cfg{ listen = 3301, logger = 'tnt.log' }

package.path = package.path .. ';./api/?.lua'

function getSpaceId(spaceName)
    local s = box.space[spaceName]
    if s then
        return s.id
    end
    return 0
end

function spaceTruncate(space)
    local s = box.space[space]
    if s then
        s:truncate()
        return true
    end
    return false
end

local cpapi = require('cpapi')
local townapi = require('townapi')
local bankapi = require('bankapi')
local clusterapi = require('clusterapi')

if not box.space.banks then
    local banks = box.schema.space.create('banks')
    banks:create_index('primary', {
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    print('created space: banks')
else
    print('space already exists: banks')
end

if not box.space.towns then
    local towns = box.schema.space.create('towns')
    towns:create_index('primary', {
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    towns:create_index('spatial', {
        type = 'RTREE',
        parts = { 2, 'ARRAY' },
        unique = false,
    })
    print('created space: towns')
else
    print('space already exists: towns')
end

if not box.space.regions then
    local regions = box.schema.space.create('regions')
    regions:create_index('primary', {
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    print('created space: regions')
else
    print('space already exists: regions')
end

if not box.space.cashpoints then
    local cashpoints = box.schema.space.create('cashpoints')
    cashpoints:create_index('primary', {
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    cashpoints:create_index('spatial', {
        type = 'RTREE',
        parts = { 2, 'ARRAY' },
        unique = false,
    })
--     cashpoints:create_index('secondary', {
--         type = 'TREE',
--         parts = { 4, 'NUM',
--         unique = false,
--     })
    print('created space: cashpoints')
else
    print('space already exists: cashpoints')
end

-- [patch_id] [cp_id] [user_id] [json_data_string] [timestamp]
if not box.space.cashpoints_patches then
    local patches = box.schema.space.create('cashpoints_patches')
    patches:create_index('primary', { -- patch_id
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    patches:create_index('target', { -- cp_id
        type = 'TREE',
        parts = { 2, 'NUM', 3, 'NUM' },
        unique = false,
    })
    print('created space: cashpoints_patches')
else
    print('space already exists: cashpoints_patches')
end

-- [vote_id] [patch_id] [user_id] [vote] [timestamp]
if not box.space.cashpoints_patches_votes then
    local votes = box.schema.space.create('cashpoints_patches_votes')
    votes:create_index('primary', { -- vote_id
        type = 'TREE',
        parts = { 1, 'NUM' },
    })
    votes:create_index('votes', {
        type = 'TREE',
        parts = { 2, 'NUM', 3, 'NUM' },
    })
    print('created space: cashpoints_patches_votes')
else
    print('space already exists: cashpoints_patches_votes')
end

if not box.space.clusters then
    local clusters = box.schema.space.create('clusters')
    clusters:create_index('primary', {
        type = 'HASH',
        parts = { 1, 'STR' },
    })
    clusters:create_index('spatial', {
        type = 'RTREE',
        parts = { 2, 'ARRAY' },
        unique = false,
    })
    print('created space: clusters')
else
    print('space already exists: clusters')
end

if not box.space.clusters_cache then
     local clusters_cache = box.schema.space.create('clusters_cache')
     clusters_cache:create_index('primary', {
         type = 'HASH',
         parts = { 1, 'STR' }
     })
     print('created space: clusters_cache')
else
     print('space already exists: clusters_cache')
end

local console = require('console')
console.listen('127.0.0.1:3302')
