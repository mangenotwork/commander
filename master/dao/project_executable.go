package dao

import (
	"encoding/json"
	"fmt"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

const TABLE_Project_Executable = "project_executable"

type DaoProjectExecutable struct{}

func (dao *DaoProjectExecutable) Set(key string, data *entity.ProjectExecutable) error {
	return Set(TABLE_Project_Executable, key, data)
}

func (dao *DaoProjectExecutable) Get(key string) (*entity.ProjectExecutable, error) {
	var data *entity.ProjectExecutable
	err := Get(TABLE_Project_Executable, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoProjectExecutable) GetALL() ([]*entity.ProjectExecutable, []string) {
	var data []*entity.ProjectExecutable
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Project_Executable))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.ProjectExecutable{}
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

func (dao *DaoProjectExecutable) Delete(key string) error {
	return Delete(TABLE_Project_Executable, key)
}

var ProjectExecutableProcessBucketName = "ProjectExecutableProcess_%s" // %s 是项目

// SetProjectExecutableProcess 部署时候分配的 TaskId
func (dao *DaoProjectExecutable) SetProjectExecutableProcess(project, TaskId string, value *entity.ExecutableProcess) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	var err error
	return db.Update(func(tx *bolt.Tx) error {
		bkName := fmt.Sprintf(ProjectExecutableProcessBucketName, project)
		b := tx.Bucket([]byte(bkName))
		logger.Info("SetPerformance b = ", b)
		if b == nil {
			b, err = tx.CreateBucket([]byte(bkName))
			if err != nil {
				logger.Error("create the bucket [%s] failed! %v\n", bkName, err)
				return err
			}
		}
		if b != nil {
			valueB, valueErr := utils.Any2JsonB(value)
			if valueErr != nil {
				logger.Error("put the data of new block into Dbfailed! %v\n", err)
				return err
			}
			return b.Put([]byte(TaskId), valueB)
		}
		return nil
	})
}

func (dao *DaoProjectExecutable) GetProjectExecutableProcess(project string) ([]*entity.ExecutableProcess, []string, error) {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	var data []*entity.ExecutableProcess
	var keys []string
	bkName := fmt.Sprintf(ProjectExecutableProcessBucketName, project)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkName))
		if b == nil {
			return fmt.Errorf("bkName is null.")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.ExecutableProcess{}
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
	return data, keys, err
}
