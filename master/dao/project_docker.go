package dao

import (
	"encoding/json"
	"fmt"

	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"

	"github.com/boltdb/bolt"
)

const TABLE_Project_Docker = "project_docker"

type DaoProjectDocker struct{}

func (dao *DaoProjectDocker) Set(key string, data *entity.ProjectDocker) error {
	return Set(TABLE_Project_Docker, key, data)
}

func (dao *DaoProjectDocker) Get(key string) (*entity.ProjectDocker, error) {
	var data *entity.ProjectDocker
	err := Get(TABLE_Project_Docker, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoProjectDocker) GetALL() ([]*entity.ProjectDocker, []string) {
	var data []*entity.ProjectDocker
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Project_Docker))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.ProjectDocker{}
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

func (dao *DaoProjectDocker) Delete(key string) error {
	return Delete(TABLE_Project_Docker, key)
}

var ProjectDockerContainerBucketName = "ProjectDockerContainer_%s" // %s 是项目

func (dao *DaoProjectDocker) SetProjectDockerContainer(project, containerId string, value *entity.DockerContainerDeploy) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	var err error
	return db.Update(func(tx *bolt.Tx) error {
		bkName := fmt.Sprintf(ProjectDockerContainerBucketName, project)
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
				logger.Error("put the data of new block into Dbfailed! %v\n", valueErr)
				return valueErr
			}
			return b.Put([]byte(containerId), valueB)
		}
		return nil
	})
}

func (dao *DaoProjectDocker) GetProjectDockerContainer(project string) ([]*entity.DockerContainerDeploy, []string, error) {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	var data []*entity.DockerContainerDeploy
	var keys []string
	bkName := fmt.Sprintf(ProjectDockerContainerBucketName, project)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkName))
		if b == nil {
			return fmt.Errorf("bkName is null.")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.DockerContainerDeploy{}
			err := json.Unmarshal(v, &hostInfo)
			if err != nil {
				logger.Error(err)
				continue
			}
			data = append(data, hostInfo)
		}
		return nil
	})
	return data, keys, err
}

func (dao *DaoProjectDocker) DelProjectDockerContainer(project, containerId string) error {
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	bkName := fmt.Sprintf(ProjectDockerContainerBucketName, project)
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bkName))
		if err := b.Delete([]byte(containerId)); err != nil {
			logger.Info("你要删除的key不存在")
			return err
		}
		return nil
	})
}
