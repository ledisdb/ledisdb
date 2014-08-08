#!/usr/bin/env python
# coding: utf-8

# refer: https://github.com/ideawu/ssdb/blob/master/tools/redis-import.php

# Notice: for zset, float score will be converted to integer.

import sys
import os
from collections import OrderedDict as od

import redis

total = 0
entries = 0


def scan_available(redis_client):
    """"Scan Command is available since redis-server 2.8.0"""

    if "scan" in dir(redis_client):
        info = redis_client.info()
        server_version = info["redis_version"]
        version_list = server_version.split(".")
        if len(version_list) > 2:
            n = int(version_list[0]) * 10 + int(version_list[1])
            if n >= 28:
                return True
    return False


def copy_key(redis_client, ledis_client, key):
    global entries
    k_type = redis_client.type(key)
    if k_type == "string":
        value = redis_client.get(key)
        ledis_client.set(key, value)
        entries += 1

    elif k_type == "list":
        _list = redis_client.lrange(key, 0, -1)
        for value in _list:
            ledis_client.rpush(key, value)
        entries += 1

    elif k_type == "hash":
        mapping = od(redis_client.hgetall(key))
        ledis_client.hmset(key, mapping)
        entries += 1

    elif k_type == "zset":
        out = redis_client.zrange(key, 0, -1, withscores=True)
        pieces = od()
        for i in od(out).iteritems():
            pieces[i[0]] = int(i[1])
        ledis_client.zadd(key, **pieces)
        entries += 1

    else:
        print "%s is not supported by LedisDB." % k_type


def copy_keys(redis_client, ledis_client, keys):
    for key in keys:
        copy_key(redis_client, ledis_client, key)


def copy(redis_client, ledis_client, redis_db):
    global total
    if scan_available(redis_client):
        total = redis_client.dbsize()
        # scan return a
        keys = redis_client.scan(cursor=0, count=total)[1] 
        copy_keys(redis_client, ledis_client, keys)

    else:
        msg = """We do not support Redis version less than 2.8.0.
            Please check both your redis server version and redis-py
            version.
              """
        print msg
        sys.exit()
    print "%d keys, %d entries copied" % (total, entries)


def usage():
    usage = """
        Usage:
        python %s redis_host redis_port redis_db ledis_host ledis_port
        """
    print usage % os.path.basename(sys.argv[0])


def main():
    if len(sys.argv) != 6:
        usage()
        sys.exit()

    (redis_host, redis_port, redis_db, ledis_host, ledis_port) = sys.argv[1:]

    redis_c = redis.Redis(host=redis_host, port=int(redis_port), db=int(redis_db))
    ledis_c = redis.Redis(host=ledis_host, port=int(ledis_port), db=int(redis_db))
    try:
        redis_c.ping()
    except redis.ConnectionError:
        print "Could not connect to Redis Server"
        sys.exit()

    try:
        ledis_c.ping()
    except redis.ConnectionError:
        print "Could not connect to LedisDB Server"
        sys.exit()

    copy(redis_c, ledis_c, redis_db)
    print "done\n"  


if __name__ == "__main__":
    main()
