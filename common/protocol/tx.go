package protocol

import (
	"fmt"
	"sync"
)

// Tx 输出 chan
type Tx struct {
	Data chan interface{}
	Err  chan error
	Stop chan int
}

var TXMap = make(map[string]*Tx)
var TXMapMutex sync.Mutex

func Set(request string) {
	TXMapMutex.Lock()
	defer TXMapMutex.Unlock()
	TXMap[request] = &Tx{
		Data: make(chan interface{}),
		Err:  make(chan error),
		Stop: make(chan int),
	}
}

func Get(request string) (*Tx, error) {
	TXMapMutex.Lock()
	defer TXMapMutex.Unlock()
	tx, ok := TXMap[request]
	if ok {
		return tx, nil
	}
	return nil, fmt.Errorf("请求ID已经过期")
}

func Close(request string) {
	TXMapMutex.Lock()
	defer TXMapMutex.Unlock()
	tx := TXMap[request]
	tx, ok := TXMap[request]
	if !ok {
		return
	}
	close(tx.Data)
	close(tx.Err)
	close(tx.Stop)
	delete(TXMap, request)
}
