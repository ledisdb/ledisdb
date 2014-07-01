# coding: utf-8
# Test Cases for list commands

import unittest
import datetime, time
import sys
sys.path.append('..')

import ledis


def current_time():
    return datetime.datetime.now()


class TestCmdList(unittest.TestCase):
    def setUp(self):
        self.l = ledis.Ledis(port=6666)

    def tearDown(self):
        self.l.lmclear('mylist', 'mylist1', 'mylist2')

    def test_lindex(self):
        self.l.rpush('mylist', '1', '2', '3')
        assert self.l.lindex('mylist', 0) == '1'
        assert self.l.lindex('mylist', 1) == '2'
        assert self.l.lindex('mylist', 2) == '3'

    def test_llen(self):
        self.l.rpush('mylist', '1', '2', '3')
        assert self.l.llen('mylist') == 3

    def test_lpop(self):
        self.l.rpush('mylist', '1', '2', '3')
        assert self.l.lpop('mylist') == '1'
        assert self.l.lpop('mylist') == '2'
        assert self.l.lpop('mylist') == '3'
        assert self.l.lpop('mylist') is None

    def test_lpush(self):
        assert self.l.lpush('mylist', '1') == 1
        assert self.l.lpush('mylist', '2') == 2
        assert self.l.lpush('mylist', '3', '4', '5') == 5
        assert self.l.lrange('mylist', 0, -1) == ['5', '4', '3', '2', '1']

    def test_lrange(self):
        self.l.rpush('mylist', '1', '2', '3', '4', '5')
        assert self.l.lrange('mylist', 0, 2) == ['1', '2', '3']
        assert self.l.lrange('mylist', 2, 10) == ['3', '4', '5']
        assert self.l.lrange('mylist', 0, -1) == ['1', '2', '3', '4', '5']

    def test_rpush(self):
        assert self.l.rpush('mylist', '1') == 1
        assert self.l.rpush('mylist', '2') == 2
        assert self.l.rpush('mylist', '3', '4') == 4
        assert self.l.lrange('mylist', 0, -1) == ['1', '2', '3', '4']

    def test_rpop(self):
        self.l.rpush('mylist', '1', '2', '3')
        assert self.l.rpop('mylist') == '3'
        assert self.l.rpop('mylist') == '2'
        assert self.l.rpop('mylist') == '1'
        assert self.l.rpop('mylist') is None

    def test_lclear(self):
        self.l.rpush('mylist', '1', '2', '3')
        assert self.l.lclear('mylist') == 3
        assert self.l.lclear('mylist') == 0

    def test_lmclear(self):
        self.l.rpush('mylist1', '1', '2', '3')
        self.l.rpush('mylist2', '1', '2', '3')
        assert self.l.lmclear('mylist1', 'mylist2') == 2

    def test_lexpire(self):
        assert not self.l.lexpire('mylist', 100)
        self.l.rpush('mylist', '1')
        assert self.l.lexpire('mylist', 100)
        assert 0 < self.l.lttl('mylist') <= 100
        assert self.l.lpersist('mylist')
        assert self.l.lttl('mylist') == -1

    def test_lexpireat_datetime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.rpush('mylist', '1')
        assert self.l.lexpireat('mylist', expire_at)
        assert 0 < self.l.lttl('mylist') <= 61

    def test_lexpireat_unixtime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.rpush('mylist', '1')
        expire_at_seconds = int(time.mktime(expire_at.timetuple()))
        assert self.l.lexpireat('mylist', expire_at_seconds)
        assert self.l.lttl('mylist') <= 61
    
    def test_lexpireat_no_key(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        assert not self.l.lexpireat('mylist', expire_at)

    def test_lttl_and_lpersist(self):
        self.l.rpush('mylist', '1')
        self.l.lexpire('mylist', 100)
        assert 0 < self.l.lttl('mylist') <= 100
        assert self.l.lpersist('mylist')
        assert self.l.lttl('mylist') == -1

