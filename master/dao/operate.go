package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"github.com/boltdb/bolt"
	"sort"
)

// TABLE_Operate 表  操作记录 数据持久化  key:操作时间  value: 操作数据
const TABLE_Operate = "operate"

type DaoOperate struct{}

var OperateUrlMap map[string]string = map[string]string{
	"/login":                               "登录操作",
	"/register":                            "注册操作",
	"/api/slave/process/kill":              "结束一个进程",
	"/api/docker/stop":                     "docker stop 操作",
	"/api/docker/rm":                       "docker rm 操作",
	"/api/docker/rmi":                      "docker rmi 操作",
	"/api/docker/pull":                     "docker pull 操作",
	"/api/docker/run":                      "dcoker run	操作",
	"/api/docker/top":                      "查看容器进程  docker top 操作",
	"/api/docker/rename":                   "修改容器名称 docker rename 操作",
	"/api/docker/restart":                  "容器重启  docker restart	操作",
	"/api/docker/pause":                    "容器暂停  docker pause	操作",
	"/api/docker/state":                    "实时监控容器性能 操作",
	"/api/executable/upload":               "上传可执行文件 操作",
	"/api/executable/download":             "下载可执行文件 操作",
	"/api/executable/delete":               "删除可执行文件 操作",
	"/api/executable/deploy":               "部署可执行文件操作",
	"/api/executable/task/delete":          "刪除已經執行的可执行文件任務  如果正在執行無法刪除 操作",
	"/api/executable/task/run":             "启动可执行文件任务操作",
	"/api/executable/task/kill":            "停止已經執行的可執行文件	操作",
	"/api/executable/task/restart":         "重啓已經執行的可执行文件進程 操作",
	"/api/monitor/rule/create":             "新增监控标准 操作",
	"/api/monitor/alarm/del":               "报警删除 操作",
	"/api/project/executable/create":       "新建 executable项目操作",
	"/api/project/executable/run":          " 部署executable项目操作",
	"/api/project/executable/download":     "下载executable项目操作",
	"/api/project/docker/create":           "新建docker容器项目	操作",
	"/api/project/docker/update":           "更新并重启容器项目操作",
	"/api/project/docker/update/image":     "更新项目的镜像，并重启操作",
	"/api/project/docker/update/duplicate": "更新副本数量操作",
	"/api/project/docker/run":              "部署docker项目操作",
	"/api/project/docker/container":        "查看项目下的docker容器 操作",
	"/api/project/docker/list":             "docker容器项目列表操作",
	"/api/gateway/run":                     "启动一个网关操作",
	"/api/gateway/list":                    "网关列表操作",
	"/api/gateway/delete":                  "删除一个网关操作",
	"/api/gateway/update/port":             "修改网关端口映射操作",
	"/api/cache/delete":                    "删除缓存操作",
	"/api/slave/hosts/update":              "修改 hosts 文件",
}

// Set 创建与修改任务
func (dao *DaoOperate) Set(clientIp, url string) error {
	note, ok := OperateUrlMap[url]
	if !ok {
		return nil
	}
	data := &entity.Operate{
		Date:     utils.NowTimeStr(),
		ClientIp: clientIp,
		Note:     note,
	}
	return Set(TABLE_Operate, utils.NowTimeStr(), data)
}

// GetALL 查看所有任务
func (dao *DaoOperate) GetALL() ([]*entity.Operate, []string) {
	var data []*entity.Operate
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Operate))
		c := b.Cursor()
		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			i++
			if i > 100 {
				break
			}
			keys = append(keys, string(k))
			hostInfo := &entity.Operate{}
			err := json.Unmarshal(v, &hostInfo)
			if err != nil {
				logger.Error(err)
				continue
			}
			data = append(data, hostInfo)
		}
		logger.Info(data)
		return nil
	})
	return data, keys
}

// GetALLPage 翻页查
func (dao *DaoOperate) GetALLPage(pg int) []*entity.Operate {
	data, _ := dao.GetALL()
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date > data[j].Date
	})
	rse := make([]*entity.Operate, 0, 10)
	fn := pg * 10
	ln := (pg + 1) * 10
	for n, v := range data {
		if n < fn {
			continue
		}
		if n > ln {
			break
		}
		rse = append(rse, v)
	}
	return rse
}

func (dao *DaoOperate) Del(key string) error {
	return Delete(TABLE_Operate, key)
}

// TODO  导出所有操作
