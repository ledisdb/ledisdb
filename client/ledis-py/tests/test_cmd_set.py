# coding: utf-8
# Test set commands

import unittest
import sys
sys.path.append('..')

import pytest

import ledis
from ledis._compat import b
from ledis import ResponseError
from util import expire_at, expire_at_seconds

l = ledis.Ledis(port=6380)


class TestCmdSet(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_sadd(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.smembers('a') == members

    def test_scard(self):
        l.sadd('a', '1', '2', '3')
        assert l.scard('a') == 3

    def test_sdiff(self):
        l.sadd('a', '1', '2', '3')
        assert l.sdiff('a', 'b') == set([b('1'), b('2'), b('3')])
        l.sadd('b', '2', '3')
        assert l.sdiff('a', 'b') == set([b('1')])

    def test_sdiffstore(self):
        l.sadd('a', '1', '2', '3')
        assert l.sdiffstore('c', 'a', 'b') == 3
        assert l.smembers('c') == set([b('1'), b('2'), b('3')])
        l.sadd('b', '2', '3')
        print l.smembers('c')
        print "before"
        assert l.sdiffstore('c', 'a', 'b') == 1
        print l.smembers('c')
        assert l.smembers('c') == set([b('1')])

    def test_sinter(self):
        l.sadd('a', '1', '2', '3')
        assert l.sinter('a', 'b') == set()
        l.sadd('b', '2', '3')
        assert l.sinter('a', 'b') == set([b('2'), b('3')])

    def test_sinterstore(self):
        l.sadd('a', '1', '2', '3')
        assert l.sinterstore('c', 'a', 'b') == 0
        assert l.smembers('c') == set()
        l.sadd('b', '2', '3')
        assert l.sinterstore('c', 'a', 'b') == 2
        assert l.smembers('c') == set([b('2'), b('3')])

    def test_sismember(self):
        l.sadd('a', '1', '2', '3')
        assert l.sismember('a', '1')
        assert l.sismember('a', '2')
        assert l.sismember('a', '3')
        assert not l.sismember('a', '4')

    def test_smembers(self):
        l.sadd('a', '1', '2', '3')
        assert l.smembers('a') == set([b('1'), b('2'), b('3')])

    def test_srem(self):
        l.sadd('a', '1', '2', '3', '4')
        assert l.srem('a', '5') == 0
        assert l.srem('a', '2', '4') == 2
        assert l.smembers('a') == set([b('1'), b('3')])

    def test_sunion(self):
        l.sadd('a', '1', '2')
        l.sadd('b', '2', '3')
        assert l.sunion('a', 'b') == set([b('1'), b('2'), b('3')])

    def test_sunionstore(self):
        l.sadd('a', '1', '2')
        l.sadd('b', '2', '3')
        assert l.sunionstore('c', 'a', 'b') == 3
        assert l.smembers('c') == set([b('1'), b('2'), b('3')])

    def test_sclear(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.sclear('a') == 3
        assert l.sclear('a') == 0

    def test_smclear(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        l.sadd('b', *members)
        assert l.smclear('a', 'b') == 2

    def test_sexpire(self):
        members = set([b('1'), b('2'), b('3')])
        assert l.sexpire('a', 100) == 0
        l.sadd('a', *members)
        assert l.sexpire('a', 100) == 1
        assert l.sttl('a') <= 100

    def test_sexpireat_datetime(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.sexpireat('a', expire_at())
        assert 0 < l.sttl('a') <= 61

    def test_sexpireat_unixtime(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.sexpireat('a', expire_at_seconds())
        assert 0 < l.sttl('a') <= 61

    def test_sexpireat_no_key(self):
        assert not l.sexpireat('a', expire_at())

    def test_sexpireat(self):
        assert l.sexpireat('a', 1577808000) == 0
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.sexpireat('a', 1577808000) == 1

    def test_sttl(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        assert l.sexpire('a', 100)
        assert l.sttl('a') <= 100

    def test_spersist(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        l.sexpire('a', 100)
        assert l.sttl('a') <= 100
        assert l.spersist('a')
        assert l.sttl('a') == -1

    def test_invalid_params(self):
        with pytest.raises(ResponseError) as excinfo:
            l.sadd("a")
        assert excinfo.value.message == "invalid command param"

    def test_invalid_value(self):
        members = set([b('1'), b('2'), b('3')])
        l.sadd('a', *members)
        self.assertRaises(ResponseError, lambda: l.sexpire('a', 'a'))

