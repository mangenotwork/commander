package dao

import (
	"encoding/json"
	"fmt"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"github.com/boltdb/bolt"
)

var TableAll = []string{TABLE_Slave, TABLE_Task, TABLE_Executable, TABLE_Project_Executable, TABLE_Project_Docker,
	TABLE_Gateway, TABLE_User, TABLE_MonitorRule, SlaveTABLE_ExecutableTask, TABLE_MonitorAlarm, TABLE_Operate,
	TABLE_HttpsProxy, TABLE_Socket5Proxy, TABLE_TCPForward}

func OpenDB() *bolt.DB {
	db, err := bolt.Open(conf.MasterConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	return db
}

// CreateTable 创建表
func CreateTable(table string) error {
	db, err := bolt.Open(conf.MasterConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer func() {
		_ = db.Close()
	}()
	err = db.Update(func(tx *bolt.Tx) error {
		//判断要创建的表是否存在
		b := tx.Bucket([]byte(table))
		if b == nil {
			_, err = tx.CreateBucket([]byte(table))
			if err != nil {
				logger.Error(err)
			}
		}
		return nil
	})
	return err
}

// DBInit 初始化数据
func DBInit() {
	db, err := bolt.Open(conf.MasterConf.DBPath.Data, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	defer func() {
		_ = db.Close()
	}()
	for _, table := range TableAll {
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

// HasTable 表是否存在
func HasTable(table string) bool {
	var rse = false
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b != nil {
			rse = true
		}
		return nil
	})
	return rse
}

// OpenPerformanceDB 性能表
func OpenPerformanceDB() *bolt.DB {
	db, err := bolt.Open(conf.MasterConf.DBPath.Performance, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	return db
}

// CreatePerformanceTable  创建性能采集表
func CreatePerformanceTable(table string) {
	db, err := bolt.Open(conf.MasterConf.DBPath.Performance, 0600, nil)
	if err != nil {
		logger.Panic(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {

		//判断要创建的表是否存在
		b := tx.Bucket([]byte(table))
		if b == nil {
			//创建叫"MyBucket"的表
			_, err := tx.CreateBucket([]byte(table))
			if err != nil {
				//也可以在这里对表做插入操作
				logger.Panic(err)
			}
		}
		return nil
	})
	//更新数据库失败
	if err != nil {
		logger.Panic(err)
	}
}

// ClearTable  创建表
func ClearTable(table string) error {
	db, err := bolt.Open(conf.MasterConf.DBPath.Data, 0600, nil)
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
	if db == nil {
		logger.Error("OpenDB is fail.")
		return fmt.Errorf("OpenDB is fail.")
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			logger.Error("未获取到表")
			return fmt.Errorf("未获取到表")
		}
		if err := b.Delete([]byte(key)); err != nil {
			return err
		}
		return nil
	})
}

func PerformanceDelete(table, key string) error {
	db := OpenPerformanceDB()
	defer func() {
		_ = db.Close()
	}()
	if db == nil {
		logger.Error("OpenDB is fail.")
		return fmt.Errorf("OpenDB is fail.")
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			logger.Error("未获取到表")
			return fmt.Errorf("未获取到表")
		}
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
