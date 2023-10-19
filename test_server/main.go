package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var name = RandStr2(10)
var reqTimes int32

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", sayBye)
	mux.HandleFunc("/favicon.ico", favicon)

	server := &http.Server{
		Addr:         ":12300",
		WriteTimeout: time.Second * 3,            //设置3秒的写超时
		Handler:      mux,
	}
	log.Println("Starting httpserver :12300")
	log.Fatal(server.ListenAndServe())
}


func sayBye(w http.ResponseWriter, r *http.Request) {
	id := GetUrlArg(r, "id")
	log.Println("sayBye...")
	atomic.AddInt32(&reqTimes, 1)
	w.Write([]byte(fmt.Sprintf("v2  %s ; times = %d ; id= %s", name, reqTimes, id)))
}

func favicon(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func RandStr2(n int) string {
	result := make([]byte, n/2)
	rand.Read(result)
	return hex.EncodeToString(result)
}

// GetUrlArg 获取URL的GET参数
func GetUrlArg(r *http.Request, name string) string {
	var arg string
	values := r.URL.Query()
	arg=values.Get(name)
	return arg
}