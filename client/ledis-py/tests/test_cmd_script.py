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

    

    