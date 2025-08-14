### 使用方式
该程序监听UDP端口，并将收到的数据包返回给客户端。
配合客户端实现udpping的功能
```bash
# 服务端
go build -o udpserver
./udpserver -i 127.0.0.1 -p 23832 -v
# 客户端使用nc发送数据测试
nc -u 127.0.0.1 23832

```
支持的参数
```bash
Usage of udpserver:
  -d    Run as a daemon
  -i string
        IP address to listen on (default "0.0.0.0")
  -p string
        Port to listen on (default "23832")
  -v    Enable verbose logging
```