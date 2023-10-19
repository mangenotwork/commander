package handler

import (
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"

	"github.com/gin-gonic/gin"
)

func TaskList(c *gin.Context) {
	pg := c.Query("pg")
	data := new(dao.DaoTask).GetALLPage(utils.Any2Int(pg))
	APIOutPut(c, 1, 1, data, "")
	return
}

func TaskDelete(c *gin.Context) {
	id := c.Query("id")
	err := new(dao.DaoTask).Del(id)
	if err != nil {
		APIOutPut(c, 1, 0, "", err.Error())
		return
	}
	APIOutPut(c, 0, 0, "", "ok")
	return
}
