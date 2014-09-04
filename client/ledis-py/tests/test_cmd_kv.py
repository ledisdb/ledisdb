# coding: utf-8
# Test Cases for k/v commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b, iteritems
from util import expire_at, expire_at_seconds


l = ledis.Ledis(port=6380)


class TestCmdKv(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_decr(self):
        assert l.delete('a') == 1
        assert l.decr('a') == -1
        assert l['a'] == b('-1')
        assert l.decr('a') == -2
        assert l['a'] == b('-2')
        assert l.decr('a', amount=5) == -7
        assert l['a'] == b('-7')

    def test_decrby(self):
        assert l.delete('a') == 1
        assert l.decrby('a') == -1
        assert l['a'] == b('-1')
        assert l.decrby('a') == -2
        assert l['a'] == b('-2')
        assert l.decrby('a', amount=5) == -7
        assert l['a'] == b('-7')

    def test_del(self):
        assert l.delete('a') == 1
        assert l.delete('a', 'b', 'c') == 3

    def test_exists(self):
        l.delete('a', 'non_exist_key')
        l.set('a', 'hello')
        assert (l.exists('a'))
        assert not (l.exists('non_exist_key'))

    def test_get(self):
        l.set('a', 'hello')
        assert l.get('a') == 'hello'
        l.set('b', '中文')
        assert l.get('b') == '中文'
        l.delete('non_exist_key')
        assert (l.get('non_exist_key')) is None

    def test_getset(self):
        l.set('a', 'hello')
        assert l.getset('a', 'world') == 'hello'
        assert l.get('a') == 'world'
        l.delete('non_exist_key')
        assert (l.getset('non_exist_key', 'non')) is None

    def test_incr(self):
        l.delete('non_exist_key')
        assert l.incr('non_exist_key') == 1
        l.set('a', 100)
        assert l.incr('a') == 101

    def test_incrby(self):
        l.delete('a')
        assert l.incrby('a', 100) == 100

        l.set('a', 100)
        assert l.incrby('a', 100) == 200
        assert l.incrby('a', amount=100) == 300

    def test_mget(self):
        l.set('a', 'hello')
        l.set('b', 'world')
        l.delete('non_exist_key')
        assert l.mget('a', 'b', 'non_exist_key') == ['hello', 'world', None]
        l.delete('a', 'b')
        assert l.mget(['a', 'b']) == [None, None]

    def test_mset(self):
        d = {'a': b('1'), 'b': b('2'), 'c': b('3')}
        assert l.mset(**d)
        for k, v in iteritems(d):
            assert l[k] == v

    def test_set(self):
        assert (l.set('a', 100))

    def test_setnx(self):
        l.delete('a')
        assert l.setnx('a', '1')
        assert l['a'] == b('1')
        assert not l.setnx('a', '2')
        assert l['a'] == b('1')

    def test_ttl(self):
        assert l.set('a', 'hello')
        assert l.expire('a', 100)
        assert l.ttl('a') <= 100
        l.delete('a')
        assert l.ttl('a') == -1
        l.set('a', 'hello')
        assert l.ttl('a') == -1

    def test_persist(self):
        assert l.set('a', 'hello')
        assert l.expire('a', 100)
        assert l.ttl('a') <= 100
        assert l.persist('a')
        l.delete('non_exist_key')
        assert not l.persist('non_exist_key')

    def test_expire(self):
        assert not l.expire('a', 100)

        l.set('a', 'hello')
        assert (l.expire('a', 100))
        l.delete('a')
        assert not (l.expire('a', 100))

    def test_expireat_datetime(self):
        l.set('a', '1')
        assert l.expireat('a', expire_at())
        assert 0 < l.ttl('a') <= 61

    def test_expireat_unixtime(self):
        l.set('a', '1')
        assert l.expireat('a', expire_at_seconds())
        assert 0 < l.ttl('a') <= 61

    def test_expireat_no_key(self):
        assert not l.expireat('a', expire_at())

    def test_expireat(self):
        l.set('a', 'hello')
        assert (l.expireat('a', 1577808000)) # time is 2020.1.1
        l.delete('a')
        assert not(l.expireat('a', 1577808000))
        