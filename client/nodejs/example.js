var ledis = require("ledis"),
        client = ledis.createClient();

client.on("error", function (err) {
        console.log("Error " + err);
});

client.set("string key", "string val", ledis.print);
client.get("string key", ledis.print);
client.hset("hash key", "hashtest 1", "some value", ledis.print);
client.hset(["hash key", "hashtest 2", "some other value"], ledis.print);
client.hkeys("hash key", function (err, replies) {
    console.log(replies.length + " replies:");
    replies.forEach(function (reply, i) {
        console.log("    " + i + ": " + reply);
    });
});

//ledis special commands
client.lpush("list key", "1", "2", "3", ledis.print);
client.lrange("list key", "0", "2", ledis.print);
client.lclear("list key", ledis.print);
client.lrange("list key", "0", "2", ledis.print);

client.zadd("zset key", 100, "m", ledis.print);
client.zexpire("zset key", 40, ledis.print);
client.zttl("zset key", ledis.print);

client.bsetbit("bit key 1", 1, 1, ledis.print);
client.bsetbit("bit key 2", 1, 1, ledis.print); 
client.bopt("and", "bit key 3", "bit key 1", "bit key 2", ledis.print);
client.bget("bit key 3", function(err, result){
    if (result=="\x02"){
        console.log("Reply: \\x02")
    }
});

//test zunionstore & zinterstore 
client.zadd("zset1", 1, "one");
client.zadd("zset1", 2, "two");

client.zadd("zset2", 1, "one");
client.zadd("zset2", 2, "two");
client.zadd("zset2", 3, "three");

client.zunionstore("out", 2, "zset1", "zset2", "weights", 2, 3, ledis.print);
client.zrange("out", 0, -1, "withscores", ledis.print);

client.zinterstore("out", 2, "zset1", "zset2", "weights", 2, 3, ledis.print);
client.zrange("out", 0, -1, "withscores", ledis.print);


//example of set commands
client.sadd("a", 1, 2, 3);
client.sadd("b", 3, 4, 5);
client.sismember("a", 1, ledis.print);
client.smembers("a", ledis.print);
client.sdiff("a", "b", ledis.print);
client.sdiffstore("c", "a", "b", ledis.print);
client.sinter("a", "b", ledis.print);
client.sinterstore("c", "a", "b", ledis.print);
client.sunion("a", "b", ledis.print);
client.sunionstore("c", "a", "b", ledis.print);
client.srem("a", 1, ledis.print);
client.sclear("c", ledis.print);
client.smclear("a", "b", ledis.print);
client.sexpire("a", 100, ledis.print);
client.sexpireat("a", 1577808000, ledis.print);
client.sttl("a", ledis.print);
client.spersist("a", ledis.print);


client.quit()
