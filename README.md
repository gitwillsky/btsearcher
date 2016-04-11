# btsearcher_engine
磁力搜索引擎的go实现。

### 包说明

1. dht: BitTorrent DHT 网络爬虫的实现。涉及DHT网络原理，KRPC通信协议。
2. filter: 种子信息过滤，过滤规则在conf/文件夹下定义。
3. models：Mysql数据库表结构定义。
4. torrent：种子文件解析。

### 使用

1. `go get github.com/gitwillsky/btsearcher_engine`
2. 需要编辑conf/下的配置文件**app.conf**，重点是DHT的UDP端口以及Mysql数据库链接信息。
3. `go run btengine.go`


### 关于前端展示示例

前端展示示例在 https://github.com/gitwillsky/btsearcher_web
