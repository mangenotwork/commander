package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

)

// ResponseJson 统一接口输出
type ResponseJson struct {
	Code      int64       `json:"code"`
	Msg       string      `json:"msg"`
	Body      interface{} `json:"body"`
	TimeStamp int64       `json:"timeStamp"`
}

// Body 具体数据模型
type Body struct {
	Count int `json:"count"`
	Data  interface{} `json:"data"`
}

// APIOutPut 统一接口输出方法
func APIOutPut(c *gin.Context, code int64, count int, data interface{}, msg string, ) {
	resp := &ResponseJson{
		Code:  code,
		Msg:   msg,
		TimeStamp: time.Now().Unix(),
	}
	resp.Body = Body{
		Count: count,
		Data:  data,
	}
	c.IndentedJSON(http.StatusOK, resp)
	return
}
