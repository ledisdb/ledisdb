# coding: utf-8
# Test Cases for bit commands

import unittest
import sys
sys.path.append('..')

import ledis
from ledis._compat import b
from util import expire_at, expire_at_seconds

l = ledis.Ledis(port=6380)


simple_script = "return {KEYS[1], KEYS[2], ARGV[1], ARGV[2]}"


class TestCmdScript(unittest.TestCase):
    def setUp(self):
        pass

    def tearDown(self):
        l.flushdb()

    def test_eval(self):
        assert l.eval(simple_script, ["key1", "key2"], "first", "second") == ["key1", "key2", "first", "second"]

    def test_evalsha(self):
        sha1 = l.scriptload(simple_script)
        assert len(sha1) == 40

        assert l.evalsha(sha1, ["key1", "key2"], "first", "second") == ["key1", "key2", "first", "second"]

    def test_scriptload(self):
        sha1 = l.scriptload(simple_script)
        assert len(sha1) == 40

    def test_scriptexists(self):
        sha1 = l.scriptload(simple_script)
        assert l.scriptexists(sha1) == [1L]

    def test_scriptflush(self):
        sha1 = l.scriptload(simple_script)
        assert l.scriptexists(sha1) == [1L]
        assert l.scriptflush() == 'OK'

        assert l.scriptexists(sha1) == [0L]







    