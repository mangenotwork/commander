package dao

import (
	"encoding/json"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

// SlaveTABLE_ExecutableTask 表  exceutable_task  slave執行任務數據持久化   key:任務ID  value:執行任務
const SlaveTABLE_ExecutableTask = "exceutable_task"

type DaoSlaveExecutableTask struct{}

func (dao *DaoSlaveExecutableTask) Set(key []byte, data *entity.ExecutableDeployTask) error {
	value, err := utils.Any2JsonB(data)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_ExecutableTask))
		if b != nil {
			err = b.Put(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *DaoSlaveExecutableTask) Get(key []byte) (*entity.ExecutableDeployTask, error) {
	var data *entity.ExecutableDeployTask
	db := OpenDB()
	defer db.Close()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_ExecutableTask))
		if b != nil {
			bt := b.Get(key)
			err := json.Unmarshal(bt, &data)
			return err
		}
		return nil
	})

	return data, err
}

func (dao *DaoSlaveExecutableTask) GetALL() ([]*entity.ExecutableDeployTask, []string) {
	var data []*entity.ExecutableDeployTask
	var keys []string
	db := OpenDB()
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
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

func (dao *DaoSlaveExecutableTask) Delete(key []byte) error {
	db := OpenDB()
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_ExecutableTask)) //打开视图
		// Delete  删除
		if err := b.Delete(key); err != nil {
			logger.Error("你要删除的key不存在")
			return err
		}
		return nil
	})
}
