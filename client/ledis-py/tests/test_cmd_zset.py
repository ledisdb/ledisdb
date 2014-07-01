# coding: utf-8
# Test Cases for list commands

import unittest
import sys
import datetime, time
sys.path.append('..')

import ledis
from ledis._compat import b, iteritems
from ledis import ResponseError


def current_time():
    return datetime.datetime.now()


class TestCmdZset(unittest.TestCase):
    def setUp(self):
        self.l = ledis.Ledis(port=6666)

    def tearDown(self):
        self.l.zclear('a')

    def test_zadd(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zrange('a', 0, -1) == ['a1', 'a2', 'a3']

    def test_zcard(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zcard('a') == 3

    def test_zcount(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zcount('a', '-inf', '+inf') == 3
        assert self.l.zcount('a', 1, 2) == 2
        assert self.l.zcount('a', 10, 20) == 0

    def test_zincrby(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zincrby('a', 'a2') == 3.0
        assert self.l.zincrby('a', 'a3', amount=5) == 8.0
        assert self.l.zscore('a', 'a2') == 3.0
        assert self.l.zscore('a', 'a3') == 8.0

    def test_zrange(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zrange('a', 0, 1) == ['a1', 'a2']
        assert self.l.zrange('a', 2, 3) == ['a3']

        #withscores
        assert self.l.zrange('a', 0, 1, withscores=True) == \
            [('a1', 1.0), ('a2', 2.0)]
        assert self.l.zrange('a', 2, 3, withscores=True) == \
            [('a3', 3.0)]

    def test_zrangebyscore(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zrangebyscore('a', 2, 4) == ['a2', 'a3', 'a4']

        # slicing with start/num
        assert self.l.zrangebyscore('a', 2, 4, start=1, num=2) == \
            ['a3', 'a4']

        # withscores 
        assert self.l.zrangebyscore('a', 2, 4, withscores=True) == \
            [('a2', 2.0), ('a3', 3.0), ('a4', 4.0)]

        # custom score function
        assert self.l.zrangebyscore('a', 2, 4, withscores=True,
                                    score_cast_func=int) == \
            [('a2', 2), ('a3', 3), ('a4', 4)]

    def test_rank(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zrank('a', 'a1') == 0
        assert self.l.zrank('a', 'a3') == 2
        assert self.l.zrank('a', 'a6') is None

    def test_zrem(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zrem('a', 'a2') == 1
        assert self.l.zrange('a', 0, -1) == ['a1', 'a3']
        assert self.l.zrem('a', 'b') == 0
        assert self.l.zrange('a', 0, -1) == ['a1', 'a3']

        # multiple keys
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zrem('a', 'a1', 'a2') == 2
        assert self.l.zrange('a', 0, -1) == ['a3']

    def test_zremrangebyrank(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zremrangebyrank('a', 1, 3) == 3
        assert self.l.zrange('a', 0, -1) == ['a1', 'a5']

    def test_zremrangebyscore(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zremrangebyscore('a', 2, 4) == 3
        assert self.l.zrange('a', 0, -1) == ['a1', 'a5']
        assert self.l.zremrangebyscore('a', 2, 4) == 0
        assert self.l.zrange('a', 0, -1) == ['a1', 'a5']
 
    def test_zrevrange(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zrevrange('a', 0, 1) == ['a3', 'a2']
        assert self.l.zrevrange('a', 1, 2) == ['a2', 'a1']

    def test_zrevrank(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zrevrank('a', 'a1') == 4
        assert self.l.zrevrank('a', 'a2') == 3
        assert self.l.zrevrank('a', 'a6') is None

    def test_zrevrangebyscore(self):
        self.l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert self.l.zrevrangebyscore('a', 4, 2) == ['a4', 'a3', 'a2']

        # slicing with start/num
        assert self.l.zrevrangebyscore('a', 4, 2, start=1, num=2) == \
                ['a3', 'a2']

        # withscores
        assert self.l.zrevrangebyscore('a', 4, 2, withscores=True) == \
                [('a4', 4.0), ('a3', 3.0), ('a2', 2.0)]

        # custom score function
        assert self.l.zrevrangebyscore('a', 4, 2, withscores=True,
                            score_cast_func=int) == \
            [('a4', 4), ('a3', 3), ('a2', 2)]

    def test_zscore(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zscore('a', 'a1') == 1.0
        assert self.l.zscore('a', 'a2') == 2.0
        assert self.l.zscore('a', 'a4') is None

    def test_zclear(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zclear('a') == 3
        assert self.l.zclear('a') == 0

    def test_zmclear(self):
        self.l.zadd('a', a1=1, a2=2, a3=3)
        self.l.zadd('b', b1=1, b2=2, b3=3)
        assert self.l.lmclear('a', 'b') == 2
        assert self.l.lmclear('c', 'd') == 2
 
    def test_zexpire(self):
        assert not self.l.zexpire('a', 100)
        self.l.zadd('a', a1=1, a2=2, a3=3)
        assert self.l.zexpire('a', 100)
        assert 0 < self.l.zttl('a') <= 100
        assert self.l.zpersist('a')
        assert self.l.zttl('a') == -1

    def test_zexpireat_datetime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.zadd('a', a1=1)
        assert self.l.zexpireat('a', expire_at)
        assert 0 < self.l.zttl('a') <= 61

    def test_zexpireat_unixtime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.zadd('a', a1=1)
        expire_at_seconds = int(time.mktime(expire_at.timetuple()))
        assert self.l.zexpireat('a', expire_at_seconds)
        assert 0 < self.l.zttl('a') <= 61

    def test_zexpireat_no_key(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        assert not self.l.zexpireat('a', expire_at)

    def test_zttl_and_zpersist(self):
        self.l.zadd('a', a1=1)
        self.l.zexpire('a', 100)
        assert 0 < self.l.zttl('a') <= 100
        assert self.l.zpersist('a')
        assert self.l.zttl('a') == -1

