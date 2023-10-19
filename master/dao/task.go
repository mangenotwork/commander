package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/enum"
	"gitee.com/mangenotework/commander/common/utils"
	"sort"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

// TABLE_Task 表   task 任务数据持久化  key:任务id  value: 任务数据
const TABLE_Task = "task"

type DaoTask struct{}

func (dao *DaoTask) Set(key string, data *entity.Task) error {
	return Set(TABLE_Task, key, data)
}

func (dao *DaoTask) SetDefaultCreate(slave, key, note string) {
	_ = dao.Set(key, &entity.Task{
		ID:       key,
		IP:       slave,
		State:    enum.TaskStateRun.Value(),
		StateStr: enum.TaskStateRun.Str(),
		Create:   utils.NowTimeStr(),
		Note:     note,
	})
}

func (dao *DaoTask) SetCompleteRse(slave, key, rse string) {
	_ = dao.Set(slave, &entity.Task{
		ID:       key,
		IP:       slave,
		Result:   rse,
		State:    enum.TaskStateComplete.Value(),
		StateStr: enum.TaskStateComplete.Str(),
		Create:   utils.NowTimeStr(),
	})
}

func (dao *DaoTask) SetBreakOffRse(slave, key, rse string) {
	_ = dao.Set(slave, &entity.Task{
		ID:       key,
		IP:       slave,
		Result:   rse,
		State:    enum.TaskStateBreakOff.Value(),
		StateStr: enum.TaskStateBreakOff.Str(),
		Create:   utils.NowTimeStr(),
	})
}

func (dao *DaoTask) Get(key string) (*entity.Task, error) {
	var data *entity.Task
	err := Get(TABLE_Task, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoTask) GetALL() ([]*entity.Task, []string) {
	var data []*entity.Task
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Task))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//logger.Info("k = ", k)
			keys = append(keys, string(k))
			hostInfo := &entity.Task{}
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

func (dao *DaoTask) Del(key string) error {
	return Delete(TABLE_Task, key)
}

func (dao *DaoTask) GetALLPage(pg int) []*entity.Task {
	data, _ := dao.GetALL()
	sort.Slice(data, func(i, j int) bool {
		return data[i].Create > data[j].Create
	})
	rse := make([]*entity.Task, 0, 10)
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
