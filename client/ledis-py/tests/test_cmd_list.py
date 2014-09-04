# coding: utf-8
# Test Cases for list commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from util import expire_at, expire_at_seconds


l = ledis.Ledis(port=6380)


class TestCmdList(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_lindex(self):
        l.rpush('mylist', '1', '2', '3')
        assert l.lindex('mylist', 0) == b('1')
        assert l.lindex('mylist', 1) == b('2')
        assert l.lindex('mylist', 2) == b('3')

    def test_llen(self):
        l.rpush('mylist', '1', '2', '3')
        assert l.llen('mylist') == 3

    def test_lpop(self):
        l.rpush('mylist', '1', '2', '3')
        assert l.lpop('mylist') == b('1')
        assert l.lpop('mylist') == b('2')
        assert l.lpop('mylist') == b('3')
        assert l.lpop('mylist') is None

    def test_lpush(self):
        assert l.lpush('mylist', '1') == 1
        assert l.lpush('mylist', '2') == 2
        assert l.lpush('mylist', '3', '4', '5') == 5
        assert l.lrange('mylist', 0, -1) == ['5', '4', '3', '2', '1']

    def test_lrange(self):
        l.rpush('mylist', '1', '2', '3', '4', '5')
        assert l.lrange('mylist', 0, 2) == ['1', '2', '3']
        assert l.lrange('mylist', 2, 10) == ['3', '4', '5']
        assert l.lrange('mylist', 0, -1) == ['1', '2', '3', '4', '5']

    def test_rpush(self):
        assert l.rpush('mylist', '1') == 1
        assert l.rpush('mylist', '2') == 2
        assert l.rpush('mylist', '3', '4') == 4
        assert l.lrange('mylist', 0, -1) == ['1', '2', '3', '4']

    def test_rpop(self):
        l.rpush('mylist', '1', '2', '3')
        assert l.rpop('mylist') == b('3')
        assert l.rpop('mylist') == b('2')
        assert l.rpop('mylist') == b('1')
        assert l.rpop('mylist') is None

    def test_lclear(self):
        l.rpush('mylist', '1', '2', '3')
        assert l.lclear('mylist') == 3
        assert l.lclear('mylist') == 0

    def test_lmclear(self):
        l.rpush('mylist1', '1', '2', '3')
        l.rpush('mylist2', '1', '2', '3')
        assert l.lmclear('mylist1', 'mylist2') == 2

    def test_lexpire(self):
        assert not l.lexpire('mylist', 100)
        l.rpush('mylist', '1')
        assert l.lexpire('mylist', 100)
        assert 0 < l.lttl('mylist') <= 100
        assert l.lpersist('mylist')
        assert l.lttl('mylist') == -1

    def test_lexpireat_datetime(self):
        l.rpush('mylist', '1')
        assert l.lexpireat('mylist', expire_at())
        assert 0 < l.lttl('mylist') <= 61

    def test_lexpireat_unixtime(self):
        l.rpush('mylist', '1')
        assert l.lexpireat('mylist', expire_at_seconds())
        assert l.lttl('mylist') <= 61

    def test_lexpireat_no_key(self):
        assert not l.lexpireat('mylist', expire_at())

    def test_lttl_and_lpersist(self):
        l.rpush('mylist', '1')
        l.lexpire('mylist', 100)
        assert 0 < l.lttl('mylist') <= 100
        assert l.lpersist('mylist')
        assert l.lttl('mylist') == -1

