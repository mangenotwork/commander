package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Int642Str int64 -> string
func Int642Str(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Get16MD5Encode 返回一个16位md5加密后的字符串
func Get16MD5Encode(data string) string {
	return GetMD5Encode(data)[8:24]
}

// GetMD5Encode 获取Md5编码
func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// DeletePreAndSufSpace 删除字符串前后两端的所有空格
func DeletePreAndSufSpace(str string) string {
	strList := []byte(str)
	spaceCount, count := 0, len(strList)
	for i := 0; i <= len(strList)-1; i++ {
		if strList[i] == 32 {
			spaceCount++
		} else {
			break
		}
	}
	strList = strList[spaceCount:]
	spaceCount, count = 0, len(strList)
	for i := count - 1; i >= 0; i-- {
		if strList[i] == 32 {
			spaceCount++
		} else {
			break
		}
	}
	return string(strList[:count-spaceCount])
}

// Str2Int64 字符串转int64
func Str2Int64(s string) int64 {
	reg := regexp.MustCompile(`[0-9]+`)
	sList := reg.FindAllString(s, -1)
	if len(sList) == 0 {
		return 0
	}

	int64num, err := strconv.ParseInt(sList[0], 10, 64)
	if err != nil {
		return 0
	}
	return int64num
}

// Num2Int 数字类字符串 转 int
func Num2Int(s string) int {
	innum, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return innum
}

// Str2Int 字符串转int
func Str2Int(s string) int {
	reg := regexp.MustCompile(`[0-9]+`)
	sList := reg.FindAllString(s, -1)
	if len(sList) == 0 {
		return 0
	}

	int64num, err := strconv.Atoi(sList[0])
	if err != nil {
		return 0
	}
	return int64num
}

// Any2Json interface{} -> json string
func Any2Json(data interface{}) (string, error) {
	jsonStr, err := json.Marshal(data)
	return string(jsonStr), err
}

// Any2JsonB interface{} -> json string
func Any2JsonB(data interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(data)
	return jsonStr, err
}

// StringValue 任何类型返回值字符串形式
func StringValue(i interface{}) string {
	var buf bytes.Buffer
	stringValue(reflect.ValueOf(i), 0, &buf)
	return buf.String()
}

// stringValue 任何类型返回值字符串形式的实现方法，私有
func stringValue(v reflect.Value, indent int, buf *bytes.Buffer) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		buf.WriteString("{\n")
		for i := 0; i < v.Type().NumField(); i++ {
			ft := v.Type().Field(i)
			fv := v.Field(i)
			if ft.Name[0:1] == strings.ToLower(ft.Name[0:1]) {
				// ignore unexported fields
				continue
			}
			if (fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Slice) && fv.IsNil() {
				// ignore unset fields
				continue
			}
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(ft.Name + ": ")
			if tag := ft.Tag.Get("sensitive"); tag == "true" {
				buf.WriteString("<sensitive>")
			} else {
				stringValue(fv, indent+2, buf)
			}
			buf.WriteString(",\n")
		}
		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	case reflect.Slice:
		nl, id, id2 := "", "", ""
		if v.Len() > 3 {
			nl, id, id2 = "\n", strings.Repeat(" ", indent), strings.Repeat(" ", indent+2)
		}
		buf.WriteString("[" + nl)
		for i := 0; i < v.Len(); i++ {
			buf.WriteString(id2)
			stringValue(v.Index(i), indent+2, buf)

			if i < v.Len()-1 {
				buf.WriteString("," + nl)
			}
		}
		buf.WriteString(nl + id + "]")
	case reflect.Map:
		buf.WriteString("{\n")
		for i, k := range v.MapKeys() {
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(k.String() + ": ")
			stringValue(v.MapIndex(k), indent+2, buf)

			if i < v.Len()-1 {
				buf.WriteString(",\n")
			}
		}
		buf.WriteString("\n" + strings.Repeat(" ", indent) + "}")
	default:
		format := "%v"
		switch v.Interface().(type) {
		case string:
			format = "%q"
		}
		fmt.Fprintf(buf, format, v.Interface())
	}
}

// Any2Map interface{} -> map[string]interface{}
func Any2Map(data interface{}) map[string]interface{} {
	if v, ok := data.(map[string]interface{}); ok {
		return v
	}
	if reflect.ValueOf(data).Kind() == reflect.String {
		dataMap, err := Json2Map(data.(string))
		if err == nil {
			return dataMap
		}
	}
	return nil
}

// Json2Map json -> map
func Json2Map(str string) (map[string]interface{}, error) {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		return nil, err
	}
	return tempMap, nil
}

// Any2Arr interface{} -> []interface{}
func Any2Arr(data interface{}) []interface{} {
	if v, ok := data.([]interface{}); ok {
		return v
	}
	return nil
}

// Any2Int interface{} -> int
func Any2Int(data interface{}) int {
	var t2 int
	switch data.(type) {
	case uint:
		t2 = int(data.(uint))
		break
	case int8:
		t2 = int(data.(int8))
		break
	case uint8:
		t2 = int(data.(uint8))
		break
	case int16:
		t2 = int(data.(int16))
		break
	case uint16:
		t2 = int(data.(uint16))
		break
	case int32:
		t2 = int(data.(int32))
		break
	case uint32:
		t2 = int(data.(uint32))
		break
	case int64:
		t2 = int(data.(int64))
		break
	case uint64:
		t2 = int(data.(uint64))
		break
	case float32:
		t2 = int(data.(float32))
		break
	case float64:
		t2 = int(data.(float64))
		break
	case string:
		t2, _ = strconv.Atoi(data.(string))
		break
	default:
		t2 = data.(int)
		break
	}
	return t2
}

// Any2Map interface{} -> map[string]interface{}
func interface2Map(data interface{}) (map[string]interface{}, error) {
	if v, ok := data.(map[string]interface{}); ok {
		return v, nil
	}
	if reflect.ValueOf(data).Kind() == reflect.String {
		return Json2Map(data.(string))
	}
	return nil, fmt.Errorf("not map type")
}

// IsHaveKey map[string]interface{} 是否存在 输入的key
func IsHaveKey(data map[string]interface{}, key string) bool {
	_, ok := data[key]
	return ok
}

// PathExists 文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//IsNotExist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

// Num2Int64 数字类字符串 转 int64
func Num2Int64(s string) int64 {
	int64num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return int64num
}

// Mkdir 创建目录
func Mkdir(dir string) error {
	return os.Mkdir(dir, 0666)
}
