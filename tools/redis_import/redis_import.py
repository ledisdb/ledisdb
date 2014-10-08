#!/usr/bin/env python
# coding: utf-8

# refer: https://github.com/ideawu/ssdb/blob/master/tools/redis-import.php

# Notice: for zset, float score will be converted to integer.

import sys
import os
from collections import OrderedDict as od

import redis
import ledis

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


def set_ttl(redis_client, ledis_client, key, k_type):
    k_types = {
        "string": ledis_client.expire,
        "list": ledis_client.lexpire,
        "hash": ledis_client.hexpire,
        "set": ledis_client.sexpire,
        "zset": ledis_client.zexpire
    }
    timeout = redis_client.ttl(key)
    if timeout > 0:
        k_types[k_type](key, timeout)


def copy_key(redis_client, ledis_client, key, convert=False):
    global entries
    k_type = redis_client.type(key)

    if k_type == "string":
        value = redis_client.get(key)
        ledis_client.set(key, value)
        set_ttl(redis_client, ledis_client, key, k_type)
        entries += 1

    elif k_type == "list":
        _list = redis_client.lrange(key, 0, -1)
        for value in _list:
            ledis_client.rpush(key, value)
        set_ttl(redis_client, ledis_client, key, k_type)
        entries += 1

    elif k_type == "hash":
        mapping = od(redis_client.hgetall(key))
        ledis_client.hmset(key, mapping)
        set_ttl(redis_client, ledis_client, key, k_type)
        entries += 1

    elif k_type == "zset":
        # dangerous to do this?
        out = redis_client.zrange(key, 0, -1, withscores=True)
        pieces = od()
        for i in od(out).iteritems():
            pieces[i[0]] = int(i[1])
        ledis_client.zadd(key, **pieces)
        set_ttl(redis_client, ledis_client, key, k_type)
        entries += 1

    elif k_type == "set":
        mbs = list(redis_client.smembers(key))

        if mbs is not None:
            ledis_client.sadd(key, *mbs)
            set_ttl(redis_client, ledis_client, key, k_type)
            entries += 1

    else:
        print "KEY %s of TYPE %s is not supported by LedisDB." % (key, k_type)


def copy_keys(redis_client, ledis_client, keys, convert=False):
    for key in keys:
        copy_key(redis_client, ledis_client, key, convert=convert)


def scan(redis_client, count=1000):
    keys = []
    total = redis_client.dbsize()
    if total > 1000:
        print "It may take a while, be patient please."

    first = True
    cursor = 0
    while cursor != 0 or first:
        cursor, data = redis_client.scan(cursor, count=count)
        keys.extend(data)
        first = False
    assert len(keys) == total
    return keys, total


def copy(redis_client, ledis_client, count=1000, convert=False):
    if scan_available(redis_client):
        print "\nTransfer begin ...\n"
        keys, total = scan(redis_client, count=count)
        copy_keys(redis_client, ledis_client, keys, convert=convert)

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


def get_prompt(choice):
    yes = set(['yes', 'ye', 'y', ''])
    no = set(['no', 'n'])

    if choice in yes:
        return True
    elif choice in no:
        return False
    else:
        sys.stdout.write("Please respond with 'yes' or 'no'")


def main():
    if len(sys.argv) < 6:
        usage()
        sys.exit()
    convert = False
    if len(sys.argv) >= 6:
        (redis_host, redis_port, redis_db, ledis_host, ledis_port) = sys.argv[1:6]
        if int(redis_db) >= 16:
            print redis_db
            sys.exit("LedisDB only support 16 databases([0-15]")

    choice = raw_input("[y/N]").lower()
    if not get_prompt(choice):
        sys.exit("No proceed")

    redis_c = redis.Redis(host=redis_host, port=int(redis_port), db=int(redis_db))
    ledis_c = ledis.Ledis(host=ledis_host, port=int(ledis_port), db=int(redis_db))
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

    copy(redis_c, ledis_c, convert=convert)
    print "Done\n"


if __name__ == "__main__":
    main()
