package dao

import (
	"encoding/json"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

// TABLE_Gateway 表  gateway  网关表   key:项目  value:网关信息
const TABLE_Gateway = "gateway"

type DaoGateway struct{}

func (dao *DaoGateway) Set(key string, data *entity.GatewayBase) error {
	return Set(TABLE_Gateway, key, data)
}

func (dao *DaoGateway) Get(key string) (*entity.GatewayBase, error) {
	var data *entity.GatewayBase
	err := Get(TABLE_Gateway, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoGateway) GetALL() ([]*entity.GatewayBase, []string) {
	var data []*entity.GatewayBase
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Gateway))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.GatewayBase{}
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

func (dao *DaoGateway) Delete(key string) error {
	return Delete(TABLE_Gateway, key)
}
