package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"github.com/boltdb/bolt"
)

var SlaveTableAll = []string{SlaveTABLE_ExecutableTask, TABLE_Gateway, SlaveTABLE_GatewayIps,
	TABLE_HttpsProxy, TABLE_Socket5Proxy, TABLE_TCPForward}

// DBInit slave 的數據持久化
func DBInit() {
	db, err := bolt.Open(conf.SlaveConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	defer db.Close()
	for _, table := range SlaveTableAll {
		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(table))
			if b == nil {
				_, err = tx.CreateBucket([]byte(table))
				if err != nil {
					logger.Panic(err)
				}
			}
			return nil
		})
		if err != nil {
			logger.Panic(err)
		}
	}
}

func OpenDB() *bolt.DB {
	db, err := bolt.Open(conf.SlaveConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	return db
}

func ClearTable(table string) error {
	db, err := bolt.Open(conf.SlaveConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(table))
	})
}

func Delete(table, key string) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if err := b.Delete([]byte(key)); err != nil {
			return err
		}
		return nil
	})
}

func Set(table, key string, data interface{}) error {
	value, err := utils.Any2JsonB(data)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b != nil {
			err = b.Put([]byte(key), value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func Get(table, key string, data interface{}) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b != nil {
			bt := b.Get([]byte(key))
			err := json.Unmarshal(bt, data)
			return err
		}
		return nil
	})
}
