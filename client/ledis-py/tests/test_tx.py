import unittest
import sys
sys.path.append("..")

import ledis


class TestTx(unittest.TestCase):
    def setUp(self):
        self.l = ledis.Ledis(port=6380)

    def tearDown(self):
        self.l.delete("a")

    def test_commit(self):
        tx = self.l.tx()
        self.l.set("a", "no-tx")
        assert self.l.get("a") == "no-tx"
        tx.begin()
        tx.set("a", "tx")
        assert self.l.get("a") == "no-tx"
        assert tx.get("a") == "tx"

        tx.commit()
        assert self.l.get("a") == "tx"

    def test_rollback(self):
        tx = self.l.tx()
        self.l.set("a", "no-tx")
        assert self.l.get("a") == "no-tx"

        tx.begin()
        tx.set("a", "tx")
        assert tx.get("a") == "tx"
        assert self.l.get("a") == "no-tx"

        tx.rollback()
        assert self.l.get("a") == "no-tx"