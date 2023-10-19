package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

var PerformanceBucketName = "Performance_%s"

type DaoPerformance struct {
}

// SetPerformance 保存性能采集数据
func (*DaoPerformance) SetPerformance(slave string, value *entity.SlavePerformance) error {
	db := OpenPerformanceDB()
	defer func() {
		_ = db.Close()
	}()
	var err error
	return db.Update(func(tx *bolt.Tx) error {
		bkName := fmt.Sprintf(PerformanceBucketName, slave)
		b := tx.Bucket([]byte(bkName))
		//logger.Info("SetPerformance b = ", b)
		if b == nil {
			b, err = tx.CreateBucket([]byte(bkName))
			if err != nil {
				logger.Info("create the bucket [%s] failed! %v\n", bkName, err)
				return err
			}
		}
		if b != nil {
			timeStamp := utils.NowTimeStr()
			value.TimeStamp = timeStamp
			valueB, valueErr := utils.Any2JsonB(value)
			if valueErr != nil {
				logger.Info("put the data of new block into Dbfailed! %v\n", valueErr)
				return valueErr
			}
			return b.Put([]byte(timeStamp), valueB)
		}
		return nil
	})
}

func (dao *DaoPerformance) Del(slave string, key string) error {
	bkName := fmt.Sprintf(PerformanceBucketName, slave)
	logger.Info("DaoPerformance del bkName = ", bkName)
	logger.Info("DaoPerformance del key = ", key)
	return PerformanceDelete(bkName, key)
}

// GetPerformanceMinute 指定时间段获取
func (*DaoPerformance) GetPerformanceMinute(slave string) (map[string]*entity.SlavePerformance, error) {
	data := make(map[string]*entity.SlavePerformance)
	db := OpenPerformanceDB()
	defer func() {
		_ = db.Close()
	}()
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		bkName := fmt.Sprintf(PerformanceBucketName, slave)
		b := tx.Bucket([]byte(bkName))
		if b == nil {
			logger.Errorf("b is null")
			b, err = tx.CreateBucket([]byte(bkName))
			if err != nil {
				logger.Info("create the bucket [%s] failed! %v\n", bkName, err)
				return err
			}
		}

		if b != nil {
			c := b.Cursor()
			if c == nil {
				logger.Errorf("c is null")
			}
			// Our time range spans the 90's decade.
			nowTime := time.Now()
			//getTime := nowTime.AddDate(0, months, days)             //年，月，日   获取一个月前的时间
			minTime := utils.Unix2Date(nowTime.Unix() - 5*60) //获取的时间的格式
			maxTime := nowTime.Format("2006-01-02 15:04:05")  //获取的时间的格式
			min := []byte(minTime)
			max := []byte(maxTime)
			// Iterate over the 90's.
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				key := string(k)
				pref := &entity.SlavePerformance{}
				err := json.Unmarshal(v, &pref)
				if err != nil {
					logger.Error(err)
					continue
				}
				data[key] = pref
			}
		}
		return nil
	})
	return data, err
}

// GetPerformanceMinuteSection 指定时间段获取
func (*DaoPerformance) GetPerformanceMinuteSection(slave, minTime, maxTime string) (map[string]*entity.SlavePerformance, error) {
	data := make(map[string]*entity.SlavePerformance)
	db := OpenPerformanceDB()
	defer func() {
		_ = db.Close()
	}()

	if db == nil {
		logger.Errorf("db 为空")
	}

	err := db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		bkName := fmt.Sprintf(PerformanceBucketName, slave)

		b := tx.Bucket([]byte(bkName))
		if b == nil {
			var err error
			logger.Errorf("b is null")
			b, err = tx.CreateBucket([]byte(bkName))
			if err != nil {
				//logger.Infof("create the bucket [%s] failed! %v\n", bkName, err)
				return err
			}
		}

		if b != nil {
			c := b.Cursor()
			// Our time range spans the 90's decade.
			min := []byte(minTime)
			max := []byte(maxTime)
			// Iterate over the 90's.
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				key := string(k)
				pref := &entity.SlavePerformance{}
				err := json.Unmarshal(v, &pref)
				if err != nil {
					logger.Error(err)
					continue
				}
				data[key] = pref
			}
		}

		return nil
	})
	return data, err
}

// GetNowPerformance 获取最近一条
func (*DaoPerformance) GetNowPerformance(slave string) (*entity.SlavePerformance, error) {
	data := &entity.SlavePerformance{}
	db := OpenPerformanceDB()
	var err error
	defer func() {
		_ = db.Close()
	}()
	err = db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		bkName := fmt.Sprintf(PerformanceBucketName, slave)
		b := tx.Bucket([]byte(bkName))
		//logger.Info("SetPerformance b = ", b)
		if b == nil {
			b, err = tx.CreateBucket([]byte(bkName))
			if err != nil {
				logger.Info("create the bucket [%s] failed! %v\n", bkName, err)
				return err
			}
		}
		c := b.Cursor()
		// Our time range spans the 90's decade.
		nowTime := time.Now()
		getTime := nowTime.AddDate(0, 0, -1)             //年，月，日   获取一个月前的时间
		minTime := getTime.Format("2006-01-02 15:04:05") //获取的时间的格式
		maxTime := nowTime.Format("2006-01-02 15:04:05") //获取的时间的格式
		min := []byte(minTime)
		max := []byte(maxTime)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			pref := &entity.SlavePerformance{}
			err := json.Unmarshal(v, &pref)
			if err != nil {
				logger.Error(err)
				continue
			}
			data = pref
			return nil
		}
		return nil
	})
	return data, err
}
