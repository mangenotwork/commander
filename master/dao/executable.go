package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

const TABLE_Executable = "executable"

// DaoExecutable 表  executable  可执行文件   key:可执行文件名  value:可执行文件名信息
type DaoExecutable struct{}

func (dao *DaoExecutable) Set(key string, data *entity.ExecutableFile) error {
	return Set(TABLE_Executable, key, data)
}

func (dao *DaoExecutable) Get(key []byte) (*entity.ExecutableFile, error) {
	var data *entity.ExecutableFile
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Executable))
		if b != nil {
			bt := b.Get(key)
			err := json.Unmarshal(bt, &data)
			return err
		}
		return nil
	})

	return data, err
}

func (dao *DaoExecutable) GetALL() ([]*entity.ExecutableFile, []string) {
	var data []*entity.ExecutableFile
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Executable))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.ExecutableFile{}
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

func (dao *DaoExecutable) Delete(key string) error {
	return Delete(TABLE_Executable, key)
}
