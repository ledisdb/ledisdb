# coding: utf-8
# Test Cases for zset commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from util import expire_at, expire_at_seconds

l = ledis.Ledis(port=6380)


class TestCmdZset(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_zadd(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zrange('a', 0, -1) == [b('a1'), b('a2'), b('a3')]

    def test_zcard(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zcard('a') == 3

    def test_zcount(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zcount('a', '-inf', '+inf') == 3
        assert l.zcount('a', 1, 2) == 2
        assert l.zcount('a', 10, 20) == 0

    def test_zincrby(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zincrby('a', 'a2') == 3
        assert l.zincrby('a', 'a3', amount=5) == 8
        assert l.zscore('a', 'a2') == 3
        assert l.zscore('a', 'a3') == 8

    def test_zrange(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zrange('a', 0, 1) == [b('a1'), b('a2')]
        assert l.zrange('a', 2, 3) == [b('a3')]

        #withscores
        assert l.zrange('a', 0, 1, withscores=True) == \
            [(b('a1'), 1), (b('a2'), 2)]
        assert l.zrange('a', 2, 3, withscores=True) == \
            [(b('a3'), 3)]

    def test_zrangebyscore(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zrangebyscore('a', 2, 4) == [b('a2'), b('a3'), b('a4')]

        # slicing with start/num
        assert l.zrangebyscore('a', 2, 4, start=1, num=2) == \
            [b('a3'), b('a4')]

        # withscores
        assert l.zrangebyscore('a', 2, 4, withscores=True) == \
            [('a2', 2), ('a3', 3), ('a4', 4)]

    def test_zrank(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zrank('a', 'a1') == 0
        assert l.zrank('a', 'a3') == 2
        assert l.zrank('a', 'a6') is None

    def test_zrem(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zrem('a', 'a2') == 1
        assert l.zrange('a', 0, -1) == [b('a1'), b('a3')]
        assert l.zrem('a', 'b') == 0
        assert l.zrange('a', 0, -1) == [b('a1'), b('a3')]

        # multiple keys
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zrem('a', 'a1', 'a2') == 2
        assert l.zrange('a', 0, -1) == [b('a3')]

    def test_zremrangebyrank(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zremrangebyrank('a', 1, 3) == 3
        assert l.zrange('a', 0, -1) == [b('a1'), b('a5')]

    def test_zremrangebyscore(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zremrangebyscore('a', 2, 4) == 3
        assert l.zrange('a', 0, -1) == [b('a1'), b('a5')]
        assert l.zremrangebyscore('a', 2, 4) == 0
        assert l.zrange('a', 0, -1) == [b('a1'), b('a5')]

    def test_zrevrange(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zrevrange('a', 0, 1) == [b('a3'), b('a2')]
        assert l.zrevrange('a', 1, 2) == [b('a2'), b('a1')]

    def test_zrevrank(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zrevrank('a', 'a1') == 4
        assert l.zrevrank('a', 'a2') == 3
        assert l.zrevrank('a', 'a6') is None

    def test_zrevrangebyscore(self):
        l.zadd('a', a1=1, a2=2, a3=3, a4=4, a5=5)
        assert l.zrevrangebyscore('a', 4, 2) == [b('a4'), b('a3'), b('a2')]

        # slicing with start/num
        assert l.zrevrangebyscore('a', 4, 2, start=1, num=2) == \
                [b('a3'), b('a2')]

        # withscores
        assert l.zrevrangebyscore('a', 4, 2, withscores=True) == \
                [(b('a4'), 4), (b('a3'), 3), (b('a2'), 2)]

    def test_zscore(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zscore('a', 'a1') == 1
        assert l.zscore('a', 'a2') == 2
        assert l.zscore('a', 'a4') is None

    def test_zclear(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zclear('a') == 3
        assert l.zclear('a') == 0

    def test_zmclear(self):
        l.zadd('a', a1=1, a2=2, a3=3)
        l.zadd('b', b1=1, b2=2, b3=3)
        assert l.lmclear('a', 'b') == 2
        assert l.lmclear('c', 'd') == 2
 
    def test_zexpire(self):
        assert not l.zexpire('a', 100)
        l.zadd('a', a1=1, a2=2, a3=3)
        assert l.zexpire('a', 100)
        assert 0 < l.zttl('a') <= 100
        assert l.zpersist('a')
        assert l.zttl('a') == -1

    def test_zexpireat_datetime(self):
        l.zadd('a', a1=1)
        assert l.zexpireat('a', expire_at())
        assert 0 < l.zttl('a') <= 61

    def test_zexpireat_unixtime(self):
        l.zadd('a', a1=1)
        assert l.zexpireat('a', expire_at_seconds())
        assert 0 < l.zttl('a') <= 61

    def test_zexpireat_no_key(self):
        assert not l.zexpireat('a', expire_at())

    def test_zttl_and_zpersist(self):
        l.zadd('a', a1=1)
        l.zexpire('a', 100)
        assert 0 < l.zttl('a') <= 100
        assert l.zpersist('a')
        assert l.zttl('a') == -1
