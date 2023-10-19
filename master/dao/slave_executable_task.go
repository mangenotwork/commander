package dao

import (
	"encoding/json"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

// SlaveTABLE_ExecutableTask 表  exceutable_task  slave執行任務數據持久化   key:任務ID  value:執行任務
const SlaveTABLE_ExecutableTask = "exceutable_task"

type DaoSlaveExecutableTask struct{}

func (dao *DaoSlaveExecutableTask) Set(key string, data *entity.ExecutableDeployTask) error {
	return Set(SlaveTABLE_ExecutableTask, key, data)
}

func (dao *DaoSlaveExecutableTask) Get(key string) (*entity.ExecutableDeployTask, error) {
	var data *entity.ExecutableDeployTask
	err := Get(SlaveTABLE_ExecutableTask, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoSlaveExecutableTask) GetALL() ([]*entity.ExecutableDeployTask, []string) {
	var data []*entity.ExecutableDeployTask
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_ExecutableTask))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.ExecutableDeployTask{}
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

func (dao *DaoSlaveExecutableTask) Delete(key string) error {
	return Delete(SlaveTABLE_ExecutableTask, key)
}
