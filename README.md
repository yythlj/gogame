## 简介
基于开源大厅游戏框架pitaya，重写消息的Dispatch逻辑，使其成为可控go线程
go线程之间通过事务管道，发送函数到对应携程执行，实现多线程下的无锁协作
ref：https://github.com/topfreegames/pitaya
docs：https://blog.csdn.net/weixin_44627989/article/details/130072534
## Required
- golang
- websocket

## Run
```
docker-compose -f docker-compose.yml up -d etcd nats mongo
go run main.go
```
基于pitaya2.3.0，对pitaya框架进行了fix源码以暴露hook
下载pitaya-2.3.0源码，放到external，并改名文件夹为pitaya-2.3.0-fix
然后将里面的fixservice.go放到 pitaya-2.3.0-fix/service
将 fixapp.go放到 pitaya-2.3.0-fix/

open browser => http://localhost:3851/web/

如果要使用可靠rpc还需要redis(reliablerpc)
docker-compose -f docker-compose.yml up -d etcd nats mongo redis
