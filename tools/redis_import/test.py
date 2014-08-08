#coding: utf-8

import random, string

import redis

from redis_import import copy


def random_word(words, length):
    return ''.join(random.choice(words) for i in range(length))


def get_words():
    word_file = "/usr/share/dict/words"
    words = open(word_file).read().splitlines()
    return words[:10]


def get_mapping(words, length=1000):
    d = {}
    for word in words:
        d[word] = random.randint(1, length)
    return d


def random_set(client, words, length=1000):
    d = get_mapping(words, length)
    client.mset(d)


def random_hset(client, words, length=1000):
    d = get_mapping(words, length)
    client.hmset("hashName", d)


def random_lpush(client, words, length=1000):
    client.lpush("listName", *words)


def random_zadd(client, words, length=1000):
    d = get_mapping(words, length)
    client.zadd("myset", **d)


def test():
    words = get_words()
    rds = redis.Redis()
    print "Flush all redis data before insert new."
    rds.flushall()

    random_set(rds, words)
    print "random_set done"
    random_hset(rds, words)
    print "random_hset done"
    random_lpush(rds, words)
    print "random_lpush done"
    random_zadd(rds, words)

    lds = redis.Redis(port=6380)
    copy(rds, lds, 0)

    # for all keys
    keys = rds.scan(0, count=rds.dbsize())
    for key in keys:
        if rds.type(key) == "string" and not lds.exists(key):
            print key
            print "String data not consistent"

    # for list
    l1 = rds.lrange("listName", 0, -1)
    l2 = lds.lrange("listName", 0, -1)
    assert l1 == l2
   
    #for hash

    for key in keys:
        if rds.type(key) == "hash" and not lds.hexists("hashName", key):
            print "List data not consistent"

    # for zset
    z1 = rds.zrange("myset", 0, -1, withscores=True)
    z2 = lds.zrange("myset", 0, -1, withscores=True)
    assert z1 == z2
    

if __name__ == "__main__":
    test()
    print "Test passed."
