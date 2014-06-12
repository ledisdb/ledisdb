# ledisdb

ledisdb是一个用go实现的类似redis的高性能nosql数据库，底层基于leveldb实现。提供了kv，list，hash以及zset几种数据结构的支持。

最开始源于[ssdb](https://github.com/ideawu/ssdb)，在使用了一段时间之后，因为兴趣的原因，决定用go实现一个。

## 编译

+ 创建一个工作目录，并check ledisdb源码

        mkdir $WORKSPACE
        cd $WORKSPACE
        git clone git@github.com:siddontang/ledisdb.git src/github.com/siddontang/ledisdb

        cd src/github.com/siddontang/ledisdb

+ 安装leveldb以及snappy，如果你已经安装，忽略
    
    我提供了一个简单的脚本进行leveldb的安装，你可以直接在shell中输入：

        sh build_leveldb.sh

    默认该脚本会将leveldb以及snappy安装到/usr/local/leveldb以及/usr/local/snappy目录

+ 在dev.sh里面设置LEVELDB_DIR以及SNAPPY_DIR为实际的安装路径，默认为/usr/local/leveldb以及/usr/local/snappy

+ 运行bootstrap.sh构建ledisdb go的依赖库

        . ./bootstap.sh 或者 source ./bootstrap.sh

+ 运行dev.sh

        . ./dev.sh 或者 source ./dev.sh

+ 编译安装ledisdb

        go install ./...

## 运行

    ./ledis-server -config=/etc/ledis.json

    //another shell
    redis-cli -p 6380
    
    redis 127.0.0.1:6380> set a 1
    OK
    redis 127.0.0.1:6380> get a
    "1"

## 嵌入库
    
    import "github.com/siddontang/ledisdb/ledis"
    l, _ := ledis.OpenWithConfig(cfg)
    db, _ := l.Select(0)

    db.Set(key, value)

    db.Get(key)

## Benchmark

可以通过查看benchmark.md获取最新的性能测试结果

## Replication

通过配置或者运行时输入slaveof开启slave的replication功能

    redis-cli -p 6381 

    redis 127.0.0.1:6381> slaveof 127.0.0.1:6380
    OK
    
## Todo

+ Admin

## 联系我

Gmail: siddontang@gmail.com