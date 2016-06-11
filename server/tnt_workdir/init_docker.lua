#!/usr/bin/tarantool
box.cfg{
    listen = 3301,
    logger = '/var/log/tarantool/cashpoints.log',
    work_dir = '/var/lib/cpsrv',
    snap_dir = 'snap',
    wal_dir = 'wal'
}

package.path = package.path .. ';/var/lib/cpsrv/lua/?.lua;/var/lib/cpsrv/lua/api/?.lua'

box.schema.user.passwd('admin', 'admin')

local init = require('init_common')

init()
