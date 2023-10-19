package dao

import (
	"encoding/json"
	"fmt"

	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

// SlaveTABLE_GatewayIps 定时持久化 网关的 ips
const SlaveTABLE_GatewayIps = "gateway_ips"

type DaoSlaveGatewayIps struct{}

func (dao *DaoSlaveGatewayIps) Set(key []byte, data []string) error {
	value, err := utils.Any2JsonB(data)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_GatewayIps))
		if b != nil {
			err = b.Put(key, value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *DaoSlaveGatewayIps) Get(key []byte) ([]string, error) {
	var data []string
	db := OpenDB()
	defer db.Close()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_GatewayIps))
		if b != nil {
			bt := b.Get(key)
			err := json.Unmarshal(bt, &data)
			return err
		}
		return nil
	})

	return data, err
}

func (dao *DaoSlaveGatewayIps) Delete(key []byte) error {
	db := OpenDB()
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(SlaveTABLE_GatewayIps)) //打开视图
		// Delete  删除
		if err := b.Delete(key); err != nil {
			fmt.Println("你要删除的key不存在")
			return err
		}
		return nil
	})
}
