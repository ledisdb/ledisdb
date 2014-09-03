# coding: utf-8
# Test Cases for bit commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from util import expire_at, expire_at_seconds

l = ledis.Ledis(port=6380)


class TestCmdScript(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testEval(self):
        assert l.eval("return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}", ["key1", "key2"], "first", "second") == ["key1", "key2", "first", "second"]

    