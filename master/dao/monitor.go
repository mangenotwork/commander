package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
	"sort"
)

// TABLE_MonitorRule 监控规则表  key:slave ip  value: 监控规则
const TABLE_MonitorRule = "monitor"

type DaoMonitorRule struct{}

func (dao *DaoMonitorRule) Set(key string, data *entity.MonitorRule) error {
	return Set(TABLE_MonitorRule, key, data)
}

func (dao *DaoMonitorRule) Get(key string) (*entity.MonitorRule, error) {
	var data *entity.MonitorRule
	err := Get(TABLE_MonitorRule, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoMonitorRule) GetALL() ([]*entity.MonitorRule, []string) {
	var data []*entity.MonitorRule
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_MonitorRule))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.MonitorRule{}
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

const TABLE_MonitorAlarm = "monitor_alarm" // key:id（随机）

func (dao *DaoMonitorRule) SetAlarm(key string, data *entity.Alarm) error {
	return Set(TABLE_MonitorAlarm, key, data)
}

func (dao *DaoMonitorRule) GetAlarm(key []byte) *entity.Alarm {
	var buf []byte
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_MonitorAlarm))
		if b != nil {
			buf = b.Get(key)
		}
		return nil
	})
	data := &entity.Alarm{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		logger.Error(err)
	}
	return data
}

func (dao *DaoMonitorRule) GetAlarmALLPage(pg int) []*entity.Alarm {
	data, _ := dao.GetAlarmALL()
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date > data[j].Date
	})
	rse := make([]*entity.Alarm, 0, 10)
	fn := pg * 10
	ln := (pg + 1) * 10
	for n, v := range data {
		if n < fn {
			continue
		}
		if n > ln {
			break
		}
		rse = append(rse, v)
	}
	return rse
}

func (dao *DaoMonitorRule) GetAlarmALL() ([]*entity.Alarm, []string) {
	var data []*entity.Alarm
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_MonitorAlarm))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.Alarm{}
			err := json.Unmarshal(v, &hostInfo)
			if err != nil {
				logger.Error(err)
				continue
			}
			data = append(data, hostInfo)
		}
		return nil
	})
	return data, keys
}

func (dao *DaoMonitorRule) DeleteAlarm(key string) error {
	return Delete(TABLE_MonitorAlarm, key)
}
