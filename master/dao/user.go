package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"github.com/boltdb/bolt"
)

const TABLE_User = "user"

type DaoUser struct{}

func (dao *DaoUser) Set(data *entity.User) error {
	value, err := utils.Any2JsonB(data)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_User))
		if b != nil {
			err = b.Put([]byte("user"), value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *DaoUser) Get() (*entity.User, error) {
	var data *entity.User
	err := Get(TABLE_User, "user", &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetWhiteIP 白名单添加ip
func (dao *DaoUser) SetWhiteIP(ip string) error {
	white := dao.GetWhite()
	white = append(white, ip)
	value, err := utils.Any2JsonB(white)
	if err != nil {
		return err
	}
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_User))
		if b != nil {
			err = b.Put([]byte("white"), value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// GetWhite 获取白名单
func (dao *DaoUser) GetWhite() []string {
	var buf []byte
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_User))
		if b != nil {
			buf = b.Get([]byte("white"))
		}
		return nil
	})
	data := make([]string, 0)
	err := json.Unmarshal(buf, &data)
	if err != nil {
		logger.Error(err)
	}
	return data
}
