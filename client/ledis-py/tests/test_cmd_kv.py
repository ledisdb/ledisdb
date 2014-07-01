# coding: utf-8
# Test Cases for k/v commands

import unittest
import sys
import datetime, time
sys.path.append('..')

import ledis
from ledis._compat import b, iteritems


def current_time():
    return datetime.datetime.now()


class TestCmdKv(unittest.TestCase):
    def setUp(self):
        self.l = ledis.Ledis(port=6666)

    def tearDown(self):
        self.l.delete('a', 'b', 'c', 'non_exist_key')

    def test_decr(self):
        assert self.l.delete('a') == 1
        assert self.l.decr('a') == -1
        assert self.l['a'] == b('-1')
        assert self.l.decr('a') == -2
        assert self.l['a'] == b('-2')
        assert self.l.decr('a', amount=5) == -7
        assert self.l['a'] == b('-7')

        #FIXME: how to test exception?
        # self.l.set('b', '234293482390480948029348230948')
        # self.assertRaises(ResponseError, self.l.delete('b'))

    def test_decrby(self):
        assert self.l.delete('a') == 1
        assert self.l.decrby('a') == -1
        assert self.l['a'] == b('-1')
        assert self.l.decrby('a') == -2
        assert self.l['a'] == b('-2')
        assert self.l.decrby('a', amount=5) == -7
        assert self.l['a'] == b('-7')

    def test_del(self):
        assert self.l.delete('a') == 1
        assert self.l.delete('a', 'b', 'c') == 3

    def test_exists(self):
        self.l.delete('a', 'non_exist_key')
        self.l.set('a', 'hello')
        self.assertTrue(self.l.exists('a'))
        self.assertFalse(self.l.exists('non_exist_key'))

    def test_get(self):
        self.l.set('a', 'hello')
        assert self.l.get('a') == 'hello'
        self.l.set('b', '中文')
        assert self.l.get('b') == '中文'
        self.l.delete('non_exist_key')
        self.assertIsNone(self.l.get('non_exist_key'))

    def test_getset(self):
        self.l.set('a', 'hello')
        assert self.l.getset('a', 'world') == 'hello'
        assert self.l.get('a') == 'world'
        self.l.delete('non_exist_key')
        self.assertIsNone(self.l.getset('non_exist_key', 'non'))

    def test_incr(self):
        self.l.delete('non_exist_key')
        assert self.l.incr('non_exist_key') == 1
        self.l.set('a', 100)
        assert self.l.incr('a') == 101

    def test_incrby(self):
        self.l.delete('a')
        assert self.l.incrby('a', 100) == 100

        self.l.set('a', 100)
        assert self.l.incrby('a', 100) == 200
        assert self.l.incrby('a', amount=100) == 300

    def test_mget(self):
        self.l.set('a', 'hello')
        self.l.set('b', 'world')
        self.l.delete('non_exist_key')
        assert self.l.mget('a', 'b', 'non_exist_key') == ['hello', 'world', None]
        self.l.delete('a', 'b')
        assert self.l.mget(['a', 'b']) == [None, None]

    def test_mset(self):
        d = {'a': b('1'), 'b': b('2'), 'c': b('3')}
        assert self.l.mset(**d)
        for k, v in iteritems(d):
            assert self.l[k] == v

    def test_set(self):
        self.assertTrue(self.l.set('a', 100))

    def test_setnx(self):
        self.l.delete('a')
        assert self.l.setnx('a', '1')
        assert self.l['a'] == b('1')
        assert not self.l.setnx('a', '2')
        assert self.l['a'] == b('1')

    def test_ttl(self):
        assert self.l.set('a', 'hello')
        assert self.l.expire('a', 100)
        assert self.l.ttl('a') <= 100
        self.l.delete('a')
        assert self.l.ttl('a') == -1
        self.l.set('a', 'hello')
        assert self.l.ttl('a') == -1

    def test_persist(self):
        assert self.l.set('a', 'hello')
        assert self.l.expire('a', 100)
        assert self.l.ttl('a') <= 100
        assert self.l.persist('a')
        self.l.delete('non_exist_key')
        assert not self.l.persist('non_exist_key')

    def test_expire(self):
        assert not self.l.expire('a', 100)

        self.l.set('a', 'hello')
        self.assertTrue(self.l.expire('a', 100))
        self.l.delete('a')
        self.assertFalse(self.l.expire('a', 100))

    def test_expireat_datetime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.set('a', '1')
        assert self.l.expireat('a', expire_at)
        assert 0 < self.l.ttl('a') <= 61

    def test_expireat_unixtime(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        self.l.set('a', '1')
        expire_at_seconds = int(time.mktime(expire_at.timetuple()))
        assert self.l.expireat('a', expire_at_seconds)
        assert 0 < self.l.ttl('a') <= 61

    def test_expireat_no_key(self):
        expire_at = current_time() + datetime.timedelta(minutes=1)
        assert not self.l.expireat('a', expire_at)

    def test_expireat(self):
        self.l.set('a', 'hello')
        self.assertTrue(self.l.expireat('a', 1577808000)) # time is 2020.1.1
        self.l.delete('a')
        self.assertFalse(self.l.expireat('a', 1577808000))
        