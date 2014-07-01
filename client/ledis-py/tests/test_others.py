# coding: utf-8
# Test Cases for other commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from ledis import ResponseError

l = ledis.Ledis(port=6380)

class TestOtherCommands(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        pass

    # server information
    def test_echo(self):
        assert l.echo('foo bar') == b('foo bar')

    def test_ping(self):
        assert l.ping()

    def test_select(self):
        assert l.select('1')
        assert l.select('15')
        self.assertRaises(ResponseError, lambda: l.select('16'))