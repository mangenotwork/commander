1.[解决] DockerPs err =  error during connect: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.24/images/json": dial unix /var/run/docker.sock: socket: too many open files
[方案]: 增加日志的传输时间
---

2. [解决] 任务列表记录不完整
   
---

3. [解决] 报警数据按日期排序 最近时间， 并且增加分页
---

4. [解决] 报警内容显示错误  cpu 超标, 超过目标:%!d(float32=60), 当前为:%!d(float32=91.022446)
---

5. [解决] docker 管理页面如果是 /docker/0 则默认显示一个在线主机
---

6. [解决] 镜像列表点击部署这个容器点不动
 
---

7. [解决] docker 容器没有 restart=always   
---

8. [解决]  slave 配置的地址路径，没有则自动生成
```bigquery
# 可执行文件保存路径
exeStoreHousePath: "/home/exeStoreHousePath_Linux/"

# 可执行文件日志保存路径
exeStoreHouseLogs: "/home/exeStoreHouseLogs/"

# 项目的可执行文件保存路径
projectExeStoreHousePath: "/home/projectExeStoreHousePath_Linux/"

# 数据持久化保存路径
dbPath:
  data: "/home/slave_linux_db/data.db"

```
---

9.  [解决] 偶尔 主页, Slave(服务器)数量 与实际不匹配
---

10. [解决]  crash  
```
2022-10-10 15:45:14 |Info  |f=master/udp_server/host.go:32 | 保存采集性能成功
    panic: runtime error: invalid memory address or nil pointer dereference
    [signal SIGSEGV: segmentation violation code=0x1 addr=0x10 pc=0xa45d0a]

goroutine 10990 [running]:
gitee.com/mangenotework/commander/master/udp_server.Hello(0xc0002ac080)
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/udp_server/host.go:39 +0x22a
gitee.com/mangenotework/commander/master/udp_server.Handle(0xc0002ac020, 0xc000322000)
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/udp_server/run.go:101 +0xba
created by gitee.com/mangenotework/commander/master/udp_server.RunUDPServer.func1
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/udp_server/run.go:55 +0x66b
exit status 2
```
---

11. [解决] 前端bug
```
ReferenceError: data is not defined
    at xr.DockerRun (192.168.0.190:996:29)
    at click (eval at ac (vue@2?pgid=:1:1), <anonymous>:3:15288)
    at fn (vue@2?pgid=:11:21646)
    at HTMLButtonElement.r (vue@2?pgid=:11:10246)
    at HTMLButtonElement.i._wrapper (vue@2?pgid=:11:59009)
```

---

12. [解决]  crash,删除缓存后没有建立表，导致取数据空指针 
```bigquery
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/pkg/mod/github.com/boltdb/bolt@v1.3.1/bucket.go:91 (0x7af991) 
        (*Bucket).Cursor: b.tx.stats.CursorCount++                                                                   
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/dao/project_executable.go:66 (0x7af995)                                                                                               
        (*DaoProjectExecutable).GetALL.func2: c := b.Cursor()                                                        
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/pkg/mod/github.com/boltdb/bolt@v1.3.1/db.go:629 (0x796c32)    
        (*DB).View: err = fn(t)                                                                                      
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/dao/project_executable.go:64 (0x7a9685)                                                                                               
        (*DaoProjectExecutable).GetALL: _=db.View(func(tx *bolt.Tx) error {                                          
/media/mange/c73f23f4-81bc-4d93-961c-bb6643e59ea6/MyGo/src/gitee.com/mangenotework/commander/master/handler/project_executable.go:111 (0x9d4e4e)                                                                                          
        ProjectExecutableList: data, _ := new(dao.DaoProjectExecutable).GetALL()  
```

--- 
13. [解决] slave 页面 有时候数据显示不到

---
14. [解决] docker 容器部署失败:
```bigquery
 ContainerCreate err = Error response from daemon: maximum retry count cannot be used with restart policy 'always'
```

---
15. [解决] docker 容器管理停止和删除 显示的通知是乱码
```bigquery
��D�;J�@�����!�;g��&�����7�F샕m*-,��]��BR������i�:�[Y!!p���"TЃ�4�!@��@��xX�ݡ����b��7�l̆��>e&3�%[9�Q3�2-hg�z2y������ܟ�i~�.A��KZ�䚱�B�z�5}���>��O߾~׷�������w����
```

---
16. [解决] 任务列表 与 操作记录没有翻页，也没有按时间来排序

---
17. [解决] executable 改为压缩的目录，按规则: 含有 可执行文件，配置文件，其他目录与文件，
需要设置一个执行命令， 执行环境变量

---
18. [已加入需求]monitor采集的性能文件，逐渐增大 需要采用方案解决这个问题

---
19. [解决] 操作记录 删除错误

---
20. [解决] 可執行文件運行日誌 改为前端轮询下发请求数据 用ws通讯

---
21. [解决] 可执行文件执行任务，应该实时轮询查看状态，如果有状态变更应发送通知

---
22. [解决] docker项目需要关联网关信息，能直达网关页

---
23. [解决] docker项目更新镜像没有启动并更替到新的镜像

---
24. [解决] docker项目 更改副本数量，容器有启动，但是项目上的副本数量没有变化

---
25. 网关的删除与关闭区分开

---
26. [解决] 错误
```bigquery
2022-10-19 15:32:22 |Info  |f=common/protocol/udp.go:219 | Command = 268ce0d3010f8d22b38997126679d1f6CMD_DockerRm: docker rm
2022/10/19 15:32:22 docker pull err  =  gzip: invalid header
2022-10-19 15:32:22 |Info  |f=slave_linux/handler/docker.go:266 | containerId = 
2022-10-19 15:32:22 |Info  |f=slave_linux/handler/docker.go:267 | 删除一个容器 ......
2022-10-19 15:32:22 |Error |f=slave_linux/handler/docker.go:276 | 删除容器失败 err = Error: No such container: 
```

---
27. 关闭或删除的网关又启动了

---
28. [解决] docker项目 容器列表 查看详情点不动

---
29. [已加入需求]需要有docker项目的删除

---
30. [已加入需求]操作记录 应该翻页与只保留最近的操作

---
31. [解决] 异常错误日志, 怀疑跟网关有关系
```bigquery
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused
连接192.168.0.192:20113失败:dial tcp 192.168.0.192:20113: connect: connection refused

```

系统代理问题

---
32. [解决] slave crash
```bigquery
unexpected end of JSON input
panic: interface conversion: interface {} is nil, not string

goroutine 1013 [running]:
gitee.com/mangenotework/commander/slve_linux/gateway.GetIps({0xc000492540, 0x9}, {0xc000492536, 0x5})
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/slave_linux/gateway/main.go:414 +0x459
gitee.com/mangenotework/commander/slve_linux/gateway.RunGateway.func1()
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/slave_linux/gateway/main.go:93 +0x3c
created by gitee.com/mangenotework/commander/slve_linux/gateway.RunGateway
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/slave_linux/gateway/main.go:91 +0x28a
exit status 2

```

---
33. [解决] slave 离线概率比较高
```bigquery
|f=master/handler/host.go:23 | 192.168.3.228  离线  udpC = nil 
```

---
34. 死机

```shell
2022-11-29 16:55:38 |Error |f=master/udp_server/host.go:36 | 获取监控标准失败: unexpected end of JSON input
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x10 pc=0xa33b58]

goroutine 446 [running]:
gitee.com/mangenotework/commander/master/udp_server.Hello(0xc000067600)
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/master/udp_server/host.go:40 +0x1d8
gitee.com/mangenotework/commander/master/udp_server.Handle(0xc0000675e0, 0xc0005724c0)
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/master/udp_server/run.go:92 +0xaf
created by gitee.com/mangenotework/commander/master/udp_server.RunUDPServer.func1
        /media/ManGe/fe3b15b2-5de5-42d6-98ce-ed93b8af992b/mygo/src/gitee.com/mangenotework/commander/master/udp_server/run.go:49 +0x558

```

---
35. 需要有master 与 slave 之间的网络延迟， 需要能看到 slave当前的网络延迟

---

36. 在弱网下，docker点删除没反映

---
37. 在弱网下，创建项目，docker起不来

---
38. 项目 看不到docker 是否在正常运行

---
39. docker 管理 删除镜像 通知是空白的

---


---


---


---


---