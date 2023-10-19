module gitee.com/mangenotework/commander/master

go 1.16

replace gitee.com/mangenotework/commander/common => ../common

require (
	gitee.com/mangenotework/commander/common v0.0.0-00010101000000-000000000000
	github.com/boltdb/bolt v1.3.1
	github.com/docker/docker v20.10.17+incompatible
	github.com/gin-gonic/gin v1.8.1
	github.com/gorilla/websocket v1.5.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
)
