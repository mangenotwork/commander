package dao

import (
	"encoding/json"
	"gitee.com/mangenotework/commander/common/entity"
	"gitee.com/mangenotework/commander/common/logger"
	"github.com/boltdb/bolt"
)

// 网络代理 网络转发的数据持久化

// TABLE_HttpsProxy ==========  表  https_proxy  http/s代理表   key:名称  value:https_proxy数据
const TABLE_HttpsProxy = "https_proxy"

type DaoHttpsProxy struct{}

func (dao *DaoHttpsProxy) Set(key string, data *entity.HttpsProxy) error {
	return Set(TABLE_HttpsProxy, key, data)
}

func (dao *DaoHttpsProxy) Get(key string) (*entity.HttpsProxy, error) {
	var data *entity.HttpsProxy
	err := Get(TABLE_HttpsProxy, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoHttpsProxy) GetALL() ([]*entity.HttpsProxy, []string) {
	var data []*entity.HttpsProxy
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_HttpsProxy))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.HttpsProxy{}
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

func (dao *DaoHttpsProxy) Delete(key string) error {
	return Delete(TABLE_HttpsProxy, key)
}

// TABLE_Socket5Proxy ==========  表  socket5_proxy  socket5代理表   key:名称  value:socket5_proxy数据
const TABLE_Socket5Proxy = "socket5_proxy"

type DaoSocket5Proxy struct{}

func (dao *DaoSocket5Proxy) Set(key string, data *entity.Socket5Proxy) error {
	return Set(TABLE_Socket5Proxy, key, data)
}

func (dao *DaoSocket5Proxy) Get(key string) (*entity.Socket5Proxy, error) {
	var data *entity.Socket5Proxy
	err := Get(TABLE_Socket5Proxy, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoSocket5Proxy) GetALL() ([]*entity.Socket5Proxy, []string) {
	var data []*entity.Socket5Proxy
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_Socket5Proxy))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.Socket5Proxy{}
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

func (dao *DaoSocket5Proxy) Delete(key string) error {
	return Delete(TABLE_Socket5Proxy, key)
}

// TABLE_TCPForward ==========  表  tcp_forward  tcp转发   key:名称  value:tcp_forward数据
const TABLE_TCPForward = "tcp_forward"

type DaoTCPForward struct{}

func (dao *DaoTCPForward) Set(key string, data *entity.TCPForward) error {
	return Set(TABLE_TCPForward, key, data)
}

func (dao *DaoTCPForward) Get(key string) (*entity.TCPForward, error) {
	var data *entity.TCPForward
	err := Get(TABLE_TCPForward, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (dao *DaoTCPForward) GetALL() ([]*entity.TCPForward, []string) {
	var data []*entity.TCPForward
	var keys []string
	db := OpenDB()
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TABLE_TCPForward))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			keys = append(keys, string(k))
			hostInfo := &entity.TCPForward{}
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

func (dao *DaoTCPForward) Delete(key string) error {
	return Delete(TABLE_TCPForward, key)
}
