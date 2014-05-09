## 测试环境搭建

+ Darwin karamatoMacBook-Air.local 13.1.0 Darwin Kernel Version 13.1.0: Thu Jan 16 19:40:37 PST 2014; root:xnu-2422.90.20~2/RELEASE_X86_64 x86_64
+ 2 CPU Intel Core i5 1.7GHz
+ 4GB
+ SSD 120G

使用redis-benchmark命令进行压力测试，测试语句

    redis-benchmark -n 10000 -t set,incr,get,lpush,lpop,lrange,mset -q

redis只用内存模式，关闭持久化
ssdb使用默认的配置，手动更改代码关闭了binlog的功能：

    void BinlogQueue::add_log(char type, char cmd, const leveldb::Slice &key){
        tran_seq ++;
        //Binlog log(tran_seq, type, cmd, key);
        //batch.Put(encode_seq_key(tran_seq), log.repr());
    }

ledisdb使用跟ssdb相同的leveldb配置

leveldb配置：
    
    compression       = false
    block_size        = 32KB
    write_buffer_size = 64MB
    cache_size        = 500MB


## redis

    SET: 42735.04 requests per second
    GET: 45871.56 requests per second
    INCR: 45248.87 requests per second
    LPUSH: 45045.04 requests per second
    LPOP: 43103.45 requests per second
    LPUSH (needed to benchmark LRANGE): 44843.05 requests per second
    LRANGE_100 (first 100 elements): 14727.54 requests per second
    LRANGE_300 (first 300 elements): 6915.63 requests per second
    LRANGE_500 (first 450 elements): 5042.86 requests per second
    LRANGE_600 (first 600 elements): 3960.40 requests per second
    MSET (10 keys): 33003.30 requests per second

## ssdb

    SET: 35971.22 requests per second
    GET: 47393.37 requests per second
    INCR: 36630.04 requests per second
    LPUSH: 37174.72 requests per second
    LPOP: 38167.94 requests per second
    LPUSH (needed to benchmark LRANGE): 37593.98 requests per second
    LRANGE_100 (first 100 elements): 905.55 requests per second
    LRANGE_300 (first 300 elements): 327.78 requests per second
    LRANGE_500 (first 450 elements): 222.36 requests per second
    LRANGE_600 (first 600 elements): 165.30 requests per second
    MSET (10 keys): 33112.59 requests per second

## ledisdb

    SET: 38759.69 requests per second
    GET: 40160.64 requests per second
    INCR: 36101.08 requests per second
    LPUSH: 33003.30 requests per second
    LPOP: 27624.31 requests per second
    LPUSH (needed to benchmark LRANGE): 32894.74 requests per second
    LRANGE_100 (first 100 elements): 7352.94 requests per second
    LRANGE_300 (first 300 elements): 2867.79 requests per second
    LRANGE_500 (first 450 elements): 1778.41 requests per second
    LRANGE_600 (first 600 elements): 1590.33 requests per second
    MSET (10 keys): 21881.84 requests per second

可以看到，ledisdb的性能比redis以及ssdb略差，（至于为啥ssdb lrange性能这样差我感到比较困惑），但还不至于不可用，我觉得可能如下原因：

+ go语言自身没有c，c++快
+ ledisdb使用的cgo调用的leveldb相关函数，这块开销其实蛮大的，后续是优化重点。
