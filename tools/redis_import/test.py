#coding: utf-8

import random, string

import redis
import ledis

from redis_import import copy, scan, set_ttl

rds = redis.Redis()
lds = ledis.Ledis(port=6380)


def random_word(words, length):
    return ''.join(random.choice(words) for i in range(length))


def get_words():
    word_file = "/usr/share/dict/words"
    words = open(word_file).read().splitlines()
    return words[:1000]


def get_mapping(words, length=1000):
    d = {}
    for word in words:
        d[word] = random.randint(1, length)
    return d


def random_string(client, words, length=1000):
    d = get_mapping(words, length)
    client.mset(d)


def random_hash(client, words, length=1000):
    d = get_mapping(words, length)
    client.hmset("hashName", d)


def random_list(client, words, length=1000):
    client.lpush("listName", *words)


def random_zset(client, words, length=1000):
    d = get_mapping(words, length)
    client.zadd("zsetName", **d)


def random_set(client, words, length=1000):
    client.sadd("setName", *words)

def test():
    words = get_words()
    print "Flush all redis data before insert new."
    rds.flushall()

    random_string(rds, words)
    print "random_string done"

    random_hash(rds, words)
    print "random_hash done"
    
    random_list(rds, words)
    print "random_list done"

    random_zset(rds, words)
    print "random_zset done"

    random_set(rds, words)
    print "random_set done"

    copy(rds, lds, convert=True)

    # for all keys
    keys = scan(rds, 1000)
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
        if rds.type(key) == "hash":
            assert rds.hgetall(key) == lds.hgetall(key)
            assert sorted(rds.hkeys(key)) == sorted(lds.hkeys(key))
            assert sorted(rds.hvals(key)) == sorted(lds.hvals(key))

    # for zset
    z1 = rds.zrange("zsetName", 0, -1, withscores=True)
    z2 = lds.zrange("zsetName", 0, -1, withscores=True)
    assert z1 == z2


def ledis_ttl(ledis_client, key, k_type):
    ttls = {
        "string": lds.ttl,
        "list": lds.lttl,
        "hash": lds.httl,
        "zset": lds.zttl,
        "set": lds.sttl,
    }
    return ttls[k_type](key)


def test_ttl():
    keys, total = scan(rds, 1000)
    for key in keys:
        k_type = rds.type(key)
        rds.expire(key, (60 * 60 * 24))
        set_ttl(rds, lds, key, k_type)
        if rds.ttl(key):
            assert ledis_ttl(lds, key, k_type) > 0

if __name__ == "__main__":
    rds.flushdb()
    lds.flushdb()

    test()
    test_ttl()
    print "Test passed."
