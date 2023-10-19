package utils

import (
	"encoding/json"
	"fmt"
	"gitee.com/mangenotework/commander/common/logger"
	"strings"
)

// JsonFind 按路径寻找指定json值
// find : 寻找路径，与目录的url类似， 下面是一个例子：
// json:  {a:[{b:1},{b:2}]}
// find=/a/[0]  =>   {b:1}
// find=a/[0]/b  =>   1
func JsonFind(jsonStr, find string) (interface{}, error) {
	if !IsJson(jsonStr) {
		return nil, fmt.Errorf("不是标准的Json格式")
	}
	jxList := strings.Split(find, "/")
	jxLen := len(jxList)
	var (
		data = Any2Map(jsonStr)
		value interface{}
		err error
	)
	for i:= 0; i< jxLen; i++ {
		l := len(jxList[i])
		if l > 2 && string(jxList[i][0]) == "[" && string(jxList[i][l-1]) == "]" {
			numStr := jxList[i][1:l-1]
			dataList := Any2Arr(value)
			value = dataList[Any2Int(numStr)]
			data, err = interface2Map(value)
			if err != nil {
				continue
			}
		}else{
			if IsHaveKey(data, jxList[i]) {
				value = data[jxList[i]]
				data, err = interface2Map(value)
				if err != nil {
					continue
				}
			}else{
				value = nil
			}
		}
	}
	return value, nil
}

// IsJson 是否是json格式
func IsJson(str string) bool {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		logger.Error("err = ", err)
		return false
	}
	return true
}

