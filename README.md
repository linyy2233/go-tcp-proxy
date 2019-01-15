go-tcp-proxy
=============
golang tcp 代理，提供根据后端服务器状态关闭部分连接功能
## 安装方法
```bash
cd $GOPATH/src
mkdir github.com/linyy2233/ -p
cd github.com/linyy2233/
git clone https://github.com/linyy2233/go-tcp-proxy.git
cd go-tcp-proxy
go get ./...
./control build
```
编译成功之后，修改config.yaml文件相关信息，使用
```bash
./control start
```
即可启动

## 使用方法，以下22233为管理端口
```
1，下线一个后端服务器,多个用逗号分割，下线时会关闭正在服务的连接
  $ curl 127.0.0.1:22233/down -d "svr=140.82.7.232:443"
2，重新打开下线的服务器
  $ curl 127.0.0.1:22233/up -d "svr=140.82.7.232:443"
3，查询当前proxy的状态
  $ curl 127.0.0.1:22233/status
4，按照比例关闭后端部分服务器的连接，svr为要关闭部分连接的服务器，ratio为关闭的比例，比如关闭五分之一为5
  $ curl 127.0.0.1:22233/close -d "svr=140.82.7.232:443,111.13.100.92:443&ratio=5"
5，按照比例调度部分后端服务器的连接到其他后端服务器(使用的是关闭连接后等客户端重新连接，不是真正意义上的无感知调度哦)
  $ curl 127.0.0.1:22233/dispatch -d "from=140.82.7.232:443,111.13.100.92:443&to=111.13.100.91:443,111.13.100.93:443&ratio=5"
```

## FAQ
