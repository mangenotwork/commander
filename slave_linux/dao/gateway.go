package dao

import (
	"encoding/json"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

// TABLE_Gateway 表  gateway  网关表   key:项目  value:网关信息
const TABLE_Gateway = "gateway"

type DaoGateway struct{}

func (dao *DaoGateway) Set(key []byte, data *entity.GatewayBase) error {
	value, err := utils.Any2JsonB(data)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Gateway))
		if b != nil {
			err = b.Put(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *DaoGateway) Get(key []byte) (*entity.GatewayBase, error) {
	var data *entity.GatewayBase
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Gateway))
		if b != nil {
			bt := b.Get(key)
			err := json.Unmarshal(bt, &data)
			return err
		}
		return nil
	})
	return data, err
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

func (dao *DaoGateway) Delete(key []byte) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Gateway))
		if err := b.Delete(key); err != nil {
			logger.Error("你要删除的key不存在")
			return err
		}
		return nil
	})
}
