package server

import (
	"github.com/siddontang/ledisdb/client/go/ledis"
	"testing"
)

func TestBit(t *testing.T) {
	testBitGetSet(t)
	testBitMset(t)
	testBitCount(t)
	testBitOpt(t)
}

func testBitGetSet(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("test_cmd_bin_basic")

	//	get / set
	if v, err := ledis.Int(c.Do("bgetbit", key, 1024)); err != nil {
		t.Fatal(err)
	} else if v != 0 {
		t.Fatal(v)
	}

	if ori, err := ledis.Int(c.Do("bsetbit", key, 1024, 1)); err != nil {
		t.Fatal(err)
	} else if ori != 0 {
		t.Fatal(ori)
	}

	if v, err := ledis.Int(c.Do("bgetbit", key, 1024)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	//	fetch from revert pos
	c.Do("bsetbit", key, 1000, 1)

	if v, err := ledis.Int(c.Do("bgetbit", key, -1)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	if v, err := ledis.Int(c.Do("bgetbit", key, -25)); err != nil {
		t.Fatal(err)
	} else if v != 1 {
		t.Fatal(v)
	}

	//	delete
	if drop, err := ledis.Int(c.Do("bdelete", key)); err != nil {
		t.Fatal(err)
	} else if drop != 1 {
		t.Fatal(drop)
	}

	if drop, err := ledis.Int(c.Do("bdelete", key)); err != nil {
		t.Fatal(err)
	} else if drop != 0 {
		t.Fatal(drop)
	}
}

func testBitMset(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("test_cmd_bin_mset")

	if n, err := ledis.Int(
		c.Do("bmsetbit", key,
			500, 0,
			100, 1,
			200, 1,
			1000, 1,
			900, 0,
			500000, 1,
			600, 0,
			300, 1,
			100000, 1)); err != nil {
		t.Fatal(err)
	} else if n != 9 {
		t.Fatal(n)
	}

	fillPos := []int{100, 200, 300, 1000, 100000, 500000}
	for _, pos := range fillPos {
		v, err := ledis.Int(c.Do("bgetbit", key, pos))
		if err != nil || v != 1 {
			t.Fatal(pos)
		}
	}

	//	err
	if n, err := ledis.Int(
		c.Do("bmsetbit", key, 3, 0, 2, 1, 3, 0, 1, 1)); err == nil || n != 0 {
		t.Fatal(n) //	duplication on pos
	}
}

func testBitCount(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := []byte("test_cmd_bin_count")
	sum := 0
	for pos := 1; pos < 1000000; pos += 10001 {
		c.Do("bsetbit", key, pos, 1)
		sum++
	}

	if n, err := ledis.Int(c.Do("bcount", key)); err != nil {
		t.Fatal(err)
	} else if n != sum {
		t.Fatal(n)
	}
}

func testBitOpt(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	dstk := []byte("bin_op_res")
	kmiss := []byte("bin_op_miss")

	k0 := []byte("bin_op_0")
	k1 := []byte("bin_op_1")
	c.Do("bmsetbit", k0, 10, 1, 30, 1, 50, 1, 70, 1, 100, 1)
	c.Do("bmsetbit", k1, 20, 1, 40, 1, 60, 1, 80, 1, 100, 1)

	//	case - lack of args
	//	todo ...

	//	case - 'not' on inexisting key
	if blen, err := ledis.Int(
		c.Do("bopt", "not", dstk, kmiss)); err != nil {
		t.Fatal(err)
	} else if blen != 0 {
		t.Fatal(blen)
	}

	if v, _ := ledis.String(c.Do("bget", dstk)); v != "" {
		t.Fatal(v)
	}

	//	case - 'and', 'or', 'xor' with inexisting key
	opts := []string{"and", "or", "xor"}
	for _, op := range opts {
		if blen, err := ledis.Int(
			c.Do("bopt", op, dstk, kmiss, k0)); err != nil {
			t.Fatal(err)
		} else if blen != 0 {
			t.Fatal(blen)
		}
	}

	//	case - 'and'
	if blen, err := ledis.Int(
		c.Do("bopt", "and", dstk, k0, k1)); err != nil {
		t.Fatal(err)
	} else if blen != 101 {
		t.Fatal(blen)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 100)); v != 1 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 20)); v != 0 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 40)); v != 0 {
		t.Fatal(v)
	}

	//	case - 'or'
	if blen, err := ledis.Int(
		c.Do("bopt", "or", dstk, k0, k1)); err != nil {
		t.Fatal(err)
	} else if blen != 101 {
		t.Fatal(blen)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 100)); v != 1 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 20)); v != 1 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 40)); v != 1 {
		t.Fatal(v)
	}

	//	case - 'xor'
	if blen, err := ledis.Int(
		c.Do("bopt", "xor", dstk, k0, k1)); err != nil {
		t.Fatal(err)
	} else if blen != 101 {
		t.Fatal(blen)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 100)); v != 0 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 20)); v != 1 {
		t.Fatal(v)
	}

	if v, _ := ledis.Int(c.Do("bgetbit", dstk, 40)); v != 1 {
		t.Fatal(v)
	}

	return
}

func TestBitErrorParams(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	if _, err := c.Do("bget"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bdelete"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	// bsetbit
	if _, err := c.Do("bsetbit"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bsetbit", "test_bsetbit"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bsetbit", "test_bsetbit", "o", "v"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bsetbit", "test_bsetbit", "o", 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	// if _, err := c.Do("bsetbit", "test_bsetbit", -1, 1); err == nil  {
	// 	t.Fatal("invalid err of %v", err)
	// }

	if _, err := c.Do("bsetbit", "test_bsetbit", 1, "v"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bsetbit", "test_bsetbit", 1, 2); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bgetbit
	if _, err := c.Do("bgetbit", "test_bgetbit"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bgetbit", "test_bgetbit", "o"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	// if _, err := c.Do("bgetbit", "test_bgetbit", -1); err == nil  {
	// 	t.Fatal("invalid err of %v", err)
	// }

	//bmsetbit
	if _, err := c.Do("bmsetbit", "test_bmsetbit"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bmsetbit", "test_bmsetbit", 0, 1, 2); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bmsetbit", "test_bmsetbit", "o", "v"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bmsetbit", "test_bmsetbit", "o", 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	// if _, err := c.Do("bmsetbit", "test_bmsetbit", -1, 1); err == nil  {
	// 	t.Fatal("invalid err of %v", err)
	// }

	if _, err := c.Do("bmsetbit", "test_bmsetbit", 1, "v"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bmsetbit", "test_bmsetbit", 1, 2); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bmsetbit", "test_bmsetbit", 1, 0.1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bcount

	if _, err := c.Do("bcount"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bcount", "a", "b", "c"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bcount", 1, "a"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	// if _, err := c.Do("bcount", 1); err == nil  {
	// 	t.Fatal("invalid err of %v", err)
	// }

	//bopt
	if _, err := c.Do("bopt"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bopt", "and", 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bopt", "x", 1); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	if _, err := c.Do("bopt", ""); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bexpire
	if _, err := c.Do("bexpire", "test_bexpire"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bexpireat
	if _, err := c.Do("bexpireat", "test_bexpireat"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bttl
	if _, err := c.Do("bttl"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

	//bpersist
	if _, err := c.Do("bpersist"); err == nil {
		t.Fatal("invalid err of %v", err)
	}

}
