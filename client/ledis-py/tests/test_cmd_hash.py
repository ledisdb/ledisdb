# coding: utf-8
# Test Cases for list commands

import unittest
import sys
import datetime, time
sys.path.append('..')

import ledis
from ledis._compat import b, iteritems, itervalues
from ledis import ResponseError


def current_time():
    return datetime.datetime.now()


class TestCmdHash(unittest.TestCase):
    def setUp(self):
        self.l = ledis.Ledis(port=6666)

    def tearDown(self):
        self.l.hmclear('myhash', 'a')
        

    def test_hdel(self):
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hdel('myhash', 'field1') == 1
        assert self.l.hdel('myhash', 'field1') == 0
        assert self.l.hdel('myhash', 'field1', 'field2') == 0

    def test_hexists(self):
        self.l.hset('myhash', 'field1', 'foo')
        self.l.hdel('myhash', 'field2')
        assert self.l.hexists('myhash', 'field1') == 1
        assert self.l.hexists('myhash', 'field2') == 0      

    def test_hget(self):
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hget('myhash', 'field1') == 'foo'
        self.assertIsNone(self.l.hget('myhash', 'field2'))

    def test_hgetall(self):
        h = {'field1': 'foo', 'field2': 'bar'}
        self.l.hmset('myhash', h)
        assert self.l.hgetall('myhash') == h

    def test_hincrby(self):
        assert self.l.hincrby('myhash', 'field1') == 1
        self.l.hclear('myhash')
        assert self.l.hincrby('myhash', 'field1', 1) == 1
        assert self.l.hincrby('myhash', 'field1', 5) == 6
        assert self.l.hincrby('myhash', 'field1', -10) == -4

    def test_hkeys(self):
        h = {'field1': 'foo', 'field2': 'bar'}
        self.l.hmset('myhash', h)
        assert self.l.hkeys('myhash') == ['field1', 'field2'] 

    def test_hlen(self):
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hlen('myhash') == 1
        self.l.hset('myhash', 'field2', 'bar')
        assert self.l.hlen('myhash') == 2


    def test_hmget(self):
        assert self.l.hmset('myhash', {'a': '1', 'b': '2', 'c': '3'})
        assert self.l.hmget('myhash', 'a', 'b', 'c') == ['1', '2', '3']


    def test_hmset(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        assert self.l.hmset('myhash', h)
        assert self.l.hgetall('myhash') == h

    def test_hset(self):
        self.l.hclear('myhash')
        assert int(self.l.hset('myhash', 'field1', 'foo')) == 1
        assert self.l.hset('myhash', 'field1', 'foo') == 0

    def test_hvals(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        self.l.hmset('myhash', h)
        local_vals = list(itervalues(h))
        remote_vals = self.l.hvals('myhash')
        assert sorted(local_vals) == sorted(remote_vals)


    def test_hclear(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        self.l.hmset('myhash', h)
        assert self.l.hclear('myhash') == 3
        assert self.l.hclear('myhash') == 0


    def test_hmclear(self):
        h = {'a': '1', 'b': '2', 'c': '3'}
        self.l.hmset('myhash1', h)
        self.l.hmset('myhash2', h)
        assert self.l.hmclear('myhash1', 'myhash2') == 2


    def test_hexpire(self):
        assert self.l.hexpire('myhash', 100) == 0
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hexpire('myhash', 100) == 1
        assert self.l.httl('myhash') <= 100

    def test_hexpireat_datetime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.hset('a', 'f', 'foo')
        assert self.l.hexpireat('a', expire_at)
        assert 0 < self.l.httl('a') <= 61

    def test_hexpireat_unixtime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.hset('a', 'f', 'foo')
        expire_at_seconds = int(time.mktime(expire_at.timetuple()))
        assert self.l.hexpireat('a', expire_at_seconds)
        assert 0 < self.l.httl('a') <= 61

    def test_zexpireat_no_key(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        assert not self.l.hexpireat('a', expire_at)

    def test_hexpireat(self):
        assert self.l.hexpireat('myhash', 1577808000) == 0
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hexpireat('myhash', 1577808000) == 1

    def test_httl(self):
        self.l.hset('myhash', 'field1', 'foo')
        assert self.l.hexpire('myhash', 100)
        assert self.l.httl('myhash') <= 100

    def test_hpersist(self):
        self.l.hset('myhash', 'field1', 'foo')
        self.l.hexpire('myhash', 100)
        assert self.l.httl('myhash') <= 100
        assert self.l.hpersist('myhash')
        assert self.l.httl('myhash') == -1

