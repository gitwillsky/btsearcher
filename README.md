# btsearcher_engine
磁力搜索引擎的go实现。
Bittorrent search engine implement by Google go language.

### 包说明 (go package)

1. dht: BitTorrent DHT 网络爬虫的实现。涉及DHT网络原理，KRPC通信协议(Bittorrent DHT search spider, use [DHT theory， KPRC protocol](http://www.bittorrent.org/beps/bep_0005.html))。
2. filter: 种子信息过滤，过滤规则在conf/文件夹下定义(Torrent file filter)。
3. models：Mysql数据库表结构定义(DataBase structure)。
4. torrent：种子文件解析(Convert magnet link to torrent file)。

### 使用 (How to use)

1. `go get github.com/gitwillsky/btsearcher`
2. 需要编辑conf/下的配置文件**app.conf**，重点是DHT的UDP端口以及Mysql数据库链接信息(Need modify **app.conf** configure file at `conf/`, most important is DHT's UDP port and Mysql database options)。
3. `go run btengine.go`


### 关于前端展示示例 (UI example)

前端展示示例在 https://github.com/gitwillsky/btsearcher_web
