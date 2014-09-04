# coding: utf-8
# Test Cases for hash commands

import unittest
import sys

sys.path.append('..')

import ledis
from ledis._compat import itervalues
from util import expire_at, expire_at_seconds


l = ledis.Ledis(port=6380)


class TestCmdHash(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()
        

    def test_hdel(self):
        l.hset('myhash', 'field1', 'foo')
        assert l.hdel('myhash', 'field1') == 1
        assert l.hdel('myhash', 'field1') == 0
        assert l.hdel('myhash', 'field1', 'field2') == 0

    def test_hexists(self):
        l.hset('myhash', 'field1', 'foo')
        l.hdel('myhash', 'field2')
        assert l.hexists('myhash', 'field1') == 1
        assert l.hexists('myhash', 'field2') == 0

    def test_hget(self):
        l.hset('myhash', 'field1', 'foo')
        assert l.hget('myhash', 'field1') == 'foo'
        assert (l.hget('myhash', 'field2')) is None

    def test_hgetall(self):
        h = {'field1': 'foo', 'field2': 'bar'}
        l.hmset('myhash', h)
        assert l.hgetall('myhash') == h

    def test_hincrby(self):
        assert l.hincrby('myhash', 'field1') == 1
        l.hclear('myhash')
        assert l.hincrby('myhash', 'field1', 1) == 1
        assert l.hincrby('myhash', 'field1', 5) == 6
        assert l.hincrby('myhash', 'field1', -10) == -4

    def test_hkeys(self):
        h = {'field1': 'foo', 'field2': 'bar'}
        l.hmset('myhash', h)
        assert l.hkeys('myhash') == ['field1', 'field2'] 

    def test_hlen(self):
        l.hset('myhash', 'field1', 'foo')
        assert l.hlen('myhash') == 1
        l.hset('myhash', 'field2', 'bar')
        assert l.hlen('myhash') == 2


    def test_hmget(self):
        assert l.hmset('myhash', {'a': '1', 'b': '2', 'c': '3'})
        assert l.hmget('myhash', 'a', 'b', 'c') == ['1', '2', '3']


    def test_hmset(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        assert l.hmset('myhash', h)
        assert l.hgetall('myhash') == h

    def test_hset(self):
        l.hclear('myhash')
        assert int(l.hset('myhash', 'field1', 'foo')) == 1
        assert l.hset('myhash', 'field1', 'foo') == 0

    def test_hvals(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        l.hmset('myhash', h)
        local_vals = list(itervalues(h))
        remote_vals = l.hvals('myhash')
        assert sorted(local_vals) == sorted(remote_vals)


    def test_hclear(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        l.hmset('myhash', h)
        assert l.hclear('myhash') == 3
        assert l.hclear('myhash') == 0


    def test_hmclear(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        l.hmset('myhash1', h)
        l.hmset('myhash2', h)
        assert l.hmclear('myhash1', 'myhash2') == 2


    def test_hexpire(self):
        assert l.hexpire('myhash', 100) == 0
        l.hset('myhash', 'field1', 'foo')
        assert l.hexpire('myhash', 100) == 1
        assert l.httl('myhash') <= 100

    def test_hexpireat_datetime(self):
        l.hset('a', 'f', 'foo')
        assert l.hexpireat('a', expire_at())
        assert 0 < l.httl('a') <= 61

    def test_hexpireat_unixtime(self):
        l.hset('a', 'f', 'foo')
        assert l.hexpireat('a', expire_at_seconds())
        assert 0 < l.httl('a') <= 61

    def test_hexpireat_no_key(self):
        assert not l.hexpireat('a', expire_at())

    def test_hexpireat(self):
        assert l.hexpireat('myhash', 1577808000) == 0
        l.hset('myhash', 'field1', 'foo')
        assert l.hexpireat('myhash', 1577808000) == 1

    def test_httl(self):
        l.hset('myhash', 'field1', 'foo')
        assert l.hexpire('myhash', 100)
        assert l.httl('myhash') <= 100

    def test_hpersist(self):
        l.hset('myhash', 'field1', 'foo')
        l.hexpire('myhash', 100)
        assert l.httl('myhash') <= 100
        assert l.hpersist('myhash')
        assert l.httl('myhash') == -1

