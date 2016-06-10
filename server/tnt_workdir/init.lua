box.cfg{
    listen = 3301,
    logger = 'tnt.log',
    snap_dir = 'snap',
    wal_dir = 'wal',
}

package.path = package.path .. ';./api/?.lua'

local init = require("init_common")

init()
