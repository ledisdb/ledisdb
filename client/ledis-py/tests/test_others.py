# coding: utf-8
# Test Cases for other commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from ledis import ResponseError

l = ledis.Ledis(port=6380)
dbs = ["leveldb", "rocksdb", "goleveldb", "hyperleveldb", "lmdb", "boltdb"]

class TestOtherCommands(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    # server information
    def test_echo(self):
        assert l.echo('foo bar') == b('foo bar')

    def test_ping(self):
        assert l.ping()

    def test_select(self):
        assert l.select('1')
        assert l.select('15')
        self.assertRaises(ResponseError, lambda: l.select('16'))


    def test_info(self):
        info1 = l.info() 
        assert info1.get("db_name") in dbs
        info2 = l.info(section="server")
        assert info2.get("os") in ["linux", "darwin"]

    def test_flushdb(self):
        l.set("a", 1)
        assert l.flushdb() == "OK"
        assert l.get("a") is None

    def test_flushall(self):
        l.select(1)
        l.set("a", 1)
        assert l.get("a") == b("1")

        l.select(10)
        l.set("a", 1)
        assert l.get("a") == b("1")

        assert l.flushall() == "OK"

        assert l.get("a") is None
        l.select(1)
        assert l.get("a") is None


    # test *scan  commands

    def check_keys(self, scan_type):
        d = {
            "xscan": l.xscan,
            "sxscan": l.sxscan,
            "lxscan": l.lxscan,
            "hxscan": l.hxscan,
            "zxscan": l.zxscan,
            "bxscan": l.bxscan
        }

        key, keys = d[scan_type]()
        assert key == ""
        assert set(keys) == set([b("a"), b("b"), b("c")])

        _, keys = d[scan_type](match="a")
        assert set(keys) == set([b("a")])

        _, keys = d[scan_type](key="a")
        assert set(keys) == set([b("b"), b("c")])


    def test_xscan(self):
        d = {"a":1, "b":2, "c": 3}
        l.mset(d)
        self.check_keys("xscan") 


    def test_lxscan(self):
        l.rpush("a", 1)
        l.rpush("b", 1)
        l.rpush("c", 1)
        self.check_keys("lxscan")


    def test_hxscan(self):
        l.hset("a", "hello", "world")
        l.hset("b", "hello", "world")
        l.hset("c", "hello", "world")
        self.check_keys("hxscan")

    def test_sxscan(self):
        l.sadd("a", 1)
        l.sadd("b", 2)
        l.sadd("c", 3)
        self.check_keys("sxscan")

    def test_zxscan(self):
        l.zadd("a", 1, "a")
        l.zadd("b", 1, "a")
        l.zadd("c", 1, "a")
        self.check_keys("zxscan")

    def test_bxscan(self):
        l.bsetbit("a", 1, 1)
        l.bsetbit("b", 1, 1)
        l.bsetbit("c", 1, 1)
        self.check_keys("bxscan")

