# MGin框架之qmgo插件，用于代替内部的mgo.v2连接MongoDB

## 说明

- 使github.com/qiniu/qmgo包代替mgo.v2，支持高版本的mongo，支持各种新特性
- Ping带有超时时长，避免出现panic
- 自带连接池，无需归还
- 主要操作保持mgo.v2的主要函数，出入参一致
- 需要注意: ObjectId类型变成官方包的primitive.ObjectID

## 引入
```go
import "github.com/maczh/mgqmgo"
```

## 配置文件，存放于配置中心,使用原mongo的配置文件
```yaml
go:
  data:
    mongodb:
      uri: mongodb://user:pwd@ip:port/db
      db: db
    mongo_pool:
      min: 2       #最小连接数
      max: 20      #最大连接数
      idle: 300    #空闲超时时间，单位为秒
      timeout: 60  #socket连接超过，单位为秒
```

## 多库连接的配置文件范例
```yaml
go:
  data:
    mongodb:
      multidb: true
      dbNames: tag1,tag2
      tag1:
          uri: mongodb://user1:pwd1@ip1:port1/db1
          db: db1
      tag2:
        uri: mongodb://user2:pwd2@ip2:port2/db2
        db: db2
    mongo_pool:
      min: 2       #最小连接数
      max: 20      #最大连接数
      idle: 300    #空闲超时时间，单位为秒
      timeout: 60  #socket连接超过，单位为秒
```


## 在应用主配置文件中的配置
```yaml
go:
  application:
    name: myapp         #应用名称,用于自动注册微服务时的服务名
    port: 8080          #端口号
    ip: xxx.xxx.xxx.xxx  #微服务注册时登记的本地IP，不配可自动获取，如需指定外网IP或Docker之外的IP时配置
  discovery: nacos                      #微服务的服务发现与注册中心类型 nacos,consul,默认是 nacos
  config:                               #统一配置服务器相关
    server: http://192.168.1.5:8848/    #配置服务器地址
    server_type: nacos                  #配置服务器类型 nacos,consul,springconfig
    env: test                           #配置环境 一般常用test/prod/dev等，跟相应配置文件匹配
    type: .yml                          #文件格式，目前仅支持yaml
    mid: "-"                            #配置文件中间名
    used: qmgo                          #当前应用启用的配置,qmgo代表使用mgqmgo插件，mongodb代表使用mgo插件
    prefix:                             #配置文件名前缀定义
      mysql: mysql                      #mysql对应的配置文件名前缀，如当前配置中对应的配置文件名为 mysql-go-test.yml
      mongodb: mongodb
      qmgo: mongodb
      redis: redis
      rabbitmq: rabbitmq
      nacos: nacos
```

## 初始化
在main.go中，在执行完	mgconfig.InitConfig(configFile) 之后导入
```go
func main(){
	...
	mgin.Init(configFile)
    defer mgin.MGin.SaveExit()
	// 初始化TDengine连接
    mgin.MGin.Use("qmgo", mgqmgo.Mongo.Init, mgqmgo.Mongo.Close, mgqmgo.Mongo.Check)
}
```

## 获取qmgo连接
```go
    conn,err := mgqmgo.Mongo.GetConnection()
    if err != nil {
    	logs.Error("Mongo connection error: {}", err.Error())
    }
```

## 多库支持时获取指定库qmgo连接
```go
    td,err := mgqmgo.Mongo.GetConnection("test1")
    if err != nil {
    	logs.Error("Mongo connection error: {}", err.Error())
    }
```

## 更新日志
- v1.0.0 初次提交
