module gitee.com/mangenotework/commander/slve_linux

go 1.16

replace gitee.com/mangenotework/commander/common => ../common

require (
	gitee.com/mangenotework/commander/common v0.0.0-00010101000000-000000000000
	github.com/boltdb/bolt v1.3.1
	github.com/docker/docker v20.10.17+incompatible
	github.com/gin-gonic/gin v1.8.1
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8
)
