* Build: go build -o WebSocket_TCP
* Usage: ./WebSocket_TCP -c config.json

## JSON配置
- Mode是模式 有服务器模式（Server）和客户端模式（Client）
- IP是当前设备的IP，一般设置为nil自动获取即可，可手动指定
- 每一个规则的ID不能重复
- Port是监听的端口
- Address是连接地址，在客户端表示WebSocket服务器的地址，服务器端是TCP服务器地址
- ProxyProtocolVersion是Proxy Protocol版本，0是关闭，1是v1，2是v2，推荐查阅软件使用说明，建议写0
## 配置文件举例
```json
{
    "Mode": "Server",
	"IP": "nil", #设备IP，nil为自动获取
    "Rules": {
        "1": {
            "TLS": true or false,
            "Port": "监听端口",
            "Address": "转发目标",
            "ProxyProtocolVersion": 0
        }
    }
}
```