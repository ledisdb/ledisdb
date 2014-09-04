# coding: utf-8
# Test Cases for bit commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from util import expire_at, expire_at_seconds

l = ledis.Ledis(port=6380)


class TestCmdBit(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_bget(self):
        "bget is the same as get in K/V commands"
        l.bmsetbit('a', 0, 1, 1, 1, 2, 1, 3, 1, 4, 1, 5, 1, 6, 1, 7, 0)
        assert l.bget('a') == b('\x7f')

    def test_bdelete(self):
        l.bsetbit('a', 0, 1)
        assert l.bdelete('a')
        assert not l.bdelete('non_exists_key')

    def test_get_set_bit(self):
        assert not l.bgetbit('a', 5)
        assert not l.bsetbit('a', 5, True)
        assert l.bgetbit('a', 5)

        assert not l.bsetbit('a', 4, False)
        assert not l.bgetbit('a', 4)

        assert not l.bsetbit('a', 4, True)
        assert l.bgetbit('a', 4)

        assert l.bsetbit('a', 5, True)
        assert l.bgetbit('a', 5)

    def test_bmsetbit(self):
        assert l.bmsetbit('a', 0, 1, 2, 1, 3, 1) == 3

    def test_bcount(self):
        l.bsetbit('a', 5, 1)
        assert l.bcount('a') == 1
        l.bsetbit('a', 6, 1)
        assert l.bcount('a') == 2
        l.bsetbit('a', 5, 0)
        assert l.bcount('a') == 1
        l.bmsetbit('a', 10, 1, 20, 1, 30, 1, 40, 1)
        assert l.bcount('a') == 5
        assert l.bcount('a', 0, 10) == 2
        assert l.bcount('a', 20, 30) == 2
        assert l.bcount('a', 10, 10) == 1

    def test_bopt_not_empty_string(self):
        l.bopt('not', 'r', 'a')
        assert l.bget('r') is None

    def test_bopt(self):
        l.bmsetbit('a1', 10, 1, 30, 1, 50, 1, 70, 1, 90, 1)
        l.bmsetbit('a2', 20, 1, 40, 1, 60, 1, 80, 1, 100, 1)
        assert l.bopt('and', 'res1', 'a1', 'a2') == 101
        assert l.bcount('res1') == 0

        assert l.bopt('or', 'res2', 'a1', 'a2') == 101
        assert l.bcount('res2') == 10

        assert l.bopt('xor', 'res3', 'a1', 'a2') == 101
        assert l.bcount('res3') == 10

        assert l.bopt('not', 'res4', 'a1') == 91
        assert l.bcount('res4') == 86

    def test_bexpire(self):
        assert not l.bexpire('a', 100)
        l.bsetbit('a', 1, True)
        assert l.bexpire('a', 100)
        assert 0 < l.bttl('a') <= 100
        assert l.bpersist('a')
        assert l.bttl('a') == -1

    def test_bexpireat_datetime(self):
        l.bsetbit('a', 1, True)
        assert l.bexpireat('a', expire_at())
        assert 0 < l.bttl('a') <= 61

    def test_bexpireat_unixtime(self):
        l.bsetbit('a', 1, True)
        assert l.bexpireat('a', expire_at_seconds())
        assert 0 < l.bttl('a') <= 61

    def test_bexpireat_no_key(self):
        assert not l.bexpireat('a', expire_at())

    def test_bttl_and_bpersist(self):
        l.bsetbit('a', 1, True)
        l.bexpire('a', 100)
        assert 0 < l.bttl('a') <= 100
        assert l.bpersist('a')
        assert l.bttl('a') == -1
