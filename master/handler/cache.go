package handler

import (
	"fmt"
	"gitee.com/mangenotework/commander/common/utils"
	"os"

	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/protocol"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

// CacheSize 持久化缓存文件大小
func CacheSize(c *gin.Context) {
	f, err := os.Open(conf.MasterConf.DBPath.Data)
	if err != nil {
		APIOutPut(c, 1, 0, err.Error(), "")
		return
	}
	fi, err := f.Stat()
	if err != nil {
		APIOutPut(c, 1, 0, err.Error(), "")
		return
	}
	APIOutPut(c, 1, 0, fi.Size(), "")
	return
}

// CacheList 持久化缓存列表
func CacheList(c *gin.Context) {
	list := dao.TableAll
	slaveList := protocol.AllUdpClient.GetAllKey()
	for _, v := range slaveList {
		tableName := fmt.Sprintf(dao.PerformanceBucketName, v)
		if dao.HasTable(tableName) {
			list = append(list, tableName)
		}
	}
	APIOutPut(c, 0, 1, list, "")
	return
}

// CacheDelete 删除持久化缓存
func CacheDelete(c *gin.Context) {
	name := c.Query("name")
	err := dao.ClearTable(name)
	if err != nil {
		APIOutPut(c, 0, 1, "", "刪除失敗:"+err.Error())
		return
	}
	err = dao.CreateTable(name)
	if err != nil {
		APIOutPut(c, 0, 1, "", "刪除失敗:"+err.Error())
		return
	}
	APIOutPut(c, 0, 1, "", "刪除成功!")
	return
}

// OperateList 操作记录
func OperateList(c *gin.Context) {
	pgStr := c.Query("pg")
	pg := utils.Num2Int(pgStr)
	data := new(dao.DaoOperate).GetALLPage(pg)
	APIOutPut(c, 0, 1, data, "")
	return
}

// OperateDelete 删除操作记录
func OperateDelete(c *gin.Context) {
	date := c.Query("date")
	err := new(dao.DaoOperate).Del(date)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "ok")
	return
}
