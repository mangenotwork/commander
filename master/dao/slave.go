package dao

import (
	"encoding/json"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

// TABLE_Slave 表   slave  salve数据的持久化   key:slave的ip  value:slave的基础信息
const TABLE_Slave = "slave"

type DaoSlave struct{}

func (dao *DaoSlave) Set(key string, data *entity.HostInfo) error {
	return Set(TABLE_Slave, key, data)
}

func (dao *DaoSlave) Get(key string) (*entity.HostInfo, error) {

	var data *entity.HostInfo
	err := Get(TABLE_Slave, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoSlave) GetALL() ([]*entity.HostInfo, []string) {
	var data []*entity.HostInfo
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Slave))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.HostInfo{}
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
