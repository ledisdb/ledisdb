package mdb

import (
	crand "crypto/rand"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

// repeatedly put (overwrite) keys.
func BenchmarkTxnPut(b *testing.B) {
	initRandSource(b)
	env, path := setupBenchDB(b)
	defer teardownBenchDB(b, env, path)

	dbi := openBenchDBI(b, env)

	var ps [][]byte

	rc := newRandSourceCursor()
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	for i := 0; i < benchDBNumKeys; i++ {
		k := makeBenchDBKey(&rc)
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		ps = append(ps, k, v)
		bTxnMust(b, txn, err, "putting data")
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")

	txn, err = env.BeginTxn(nil, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := ps[rand.Intn(len(ps)/2)*2]
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		bTxnMust(b, txn, err, "putting data")
	}
	b.StopTimer()
	err = txn.Commit()
	bMust(b, err, "commiting transaction")
}

// repeatedly get random keys.
func BenchmarkTxnGetRDONLY(b *testing.B) {
	initRandSource(b)
	env, path := setupBenchDB(b)
	defer teardownBenchDB(b, env, path)

	dbi := openBenchDBI(b, env)

	var ps [][]byte

	rc := newRandSourceCursor()
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	for i := 0; i < benchDBNumKeys; i++ {
		k := makeBenchDBKey(&rc)
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		ps = append(ps, k, v)
		bTxnMust(b, txn, err, "putting data")
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")

	txn, err = env.BeginTxn(nil, RDONLY)
	bMust(b, err, "starting transaction")
	defer txn.Abort()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := txn.Get(dbi, ps[rand.Intn(len(ps))])
		if err == NotFound {
			continue
		}
		if err != nil {
			b.Fatalf("error getting data: %v", err)
		}
	}
	b.StopTimer()
}

// like BenchmarkTxnGetRDONLY, but txn.GetVal() is called instead.
func BenchmarkTxnGetValRDONLY(b *testing.B) {
	initRandSource(b)
	env, path := setupBenchDB(b)
	defer teardownBenchDB(b, env, path)

	dbi := openBenchDBI(b, env)

	var ps [][]byte

	rc := newRandSourceCursor()
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	for i := 0; i < benchDBNumKeys; i++ {
		k := makeBenchDBKey(&rc)
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		ps = append(ps, k, v)
		bTxnMust(b, txn, err, "putting data")
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")

	txn, err = env.BeginTxn(nil, RDONLY)
	bMust(b, err, "starting transaction")
	defer txn.Abort()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := txn.GetVal(dbi, ps[rand.Intn(len(ps))])
		if err == NotFound {
			continue
		}
		if err != nil {
			b.Fatalf("error getting data: %v", err)
		}
	}
	b.StopTimer()
}

// repeatedly scan all the values in a database.
func BenchmarkCursorScanRDONLY(b *testing.B) {
	initRandSource(b)
	env, path := setupBenchDB(b)
	defer teardownBenchDB(b, env, path)

	dbi := openBenchDBI(b, env)

	var ps [][]byte

	rc := newRandSourceCursor()
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	for i := 0; i < benchDBNumKeys; i++ {
		k := makeBenchDBKey(&rc)
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		ps = append(ps, k, v)
		bTxnMust(b, txn, err, "putting data")
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")

	txn, err = env.BeginTxn(nil, RDONLY)
	bMust(b, err, "starting transaction")
	defer txn.Abort()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			cur, err := txn.CursorOpen(dbi)
			bMust(b, err, "opening cursor")
			defer cur.Close()
			var count int64
			for {
				_, _, err := cur.Get(nil, nil, NEXT)
				if err == NotFound {
					return
				}
				if err != nil {
					b.Fatalf("error getting data: %v", err)
				}
				count++
			}
			if count != benchDBNumKeys {
				b.Fatalf("unexpected number of keys: %d", count)
			}
		}()
	}
	b.StopTimer()
}

// like BenchmarkCursoreScanRDONLY, but cursor.GetVal() is called instead.
func BenchmarkCursorScanValRDONLY(b *testing.B) {
	initRandSource(b)
	env, path := setupBenchDB(b)
	defer teardownBenchDB(b, env, path)

	dbi := openBenchDBI(b, env)

	var ps [][]byte

	rc := newRandSourceCursor()
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	for i := 0; i < benchDBNumKeys; i++ {
		k := makeBenchDBKey(&rc)
		v := makeBenchDBVal(&rc)
		err := txn.Put(dbi, k, v, 0)
		ps = append(ps, k, v)
		bTxnMust(b, txn, err, "putting data")
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")

	txn, err = env.BeginTxn(nil, RDONLY)
	bMust(b, err, "starting transaction")
	defer txn.Abort()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			cur, err := txn.CursorOpen(dbi)
			bMust(b, err, "opening cursor")
			defer cur.Close()
			var count int64
			for {
				_, _, err := cur.GetVal(nil, nil, NEXT)
				if err == NotFound {
					return
				}
				if err != nil {
					b.Fatalf("error getting data: %v", err)
				}
				count++
			}
			if count != benchDBNumKeys {
				b.Fatalf("unexpected number of keys: %d", count)
			}
		}()
	}
	b.StopTimer()
}

func setupBenchDB(b *testing.B) (*Env, string) {
	env, err := NewEnv()
	bMust(b, err, "creating env")
	err = env.SetMaxDBs(26)
	bMust(b, err, "setting max dbs")
	err = env.SetMapSize(1 << 30) // 1GB
	bMust(b, err, "sizing env")
	path, err := ioutil.TempDir("", "mdb_test-bench-")
	bMust(b, err, "creating temp directory")
	err = env.Open(path, 0, 0644)
	if err != nil {
		teardownBenchDB(b, env, path)
	}
	bMust(b, err, "opening database")
	return env, path
}

func openBenchDBI(b *testing.B, env *Env) DBI {
	txn, err := env.BeginTxn(nil, 0)
	bMust(b, err, "starting transaction")
	name := "benchmark"
	dbi, err := txn.DBIOpen(&name, CREATE)
	if err != nil {
		txn.Abort()
		b.Fatalf("error opening dbi: %v", err)
	}
	err = txn.Commit()
	bMust(b, err, "commiting transaction")
	return dbi
}

func teardownBenchDB(b *testing.B, env *Env, path string) {
	env.Close()
	os.RemoveAll(path)
}

func randBytes(n int) []byte {
	p := make([]byte, n)
	crand.Read(p)
	return p
}

func bMust(b *testing.B, err error, action string) {
	if err != nil {
		b.Fatalf("error %s: %v", action, err)
	}
}

func bTxnMust(b *testing.B, txn *Txn, err error, action string) {
	if err != nil {
		txn.Abort()
		b.Fatalf("error %s: %v", action, err)
	}
}

const randSourceSize = 500 << 20 // size of the 'entropy pool' for random byte generation.
const benchDBNumKeys = 100000    // number of keys to store in benchmark databases
const benchDBMaxKeyLen = 30      // maximum length for database keys (size is limited by MDB)
const benchDBMaxValLen = 2000    // maximum lengh for database values

func makeBenchDBKey(c *randSourceCursor) []byte {
	return c.NBytes(rand.Intn(benchDBMaxKeyLen) + 1)
}

func makeBenchDBVal(c *randSourceCursor) []byte {
	return c.NBytes(rand.Intn(benchDBMaxValLen) + 1)
}

// holds a bunch of random bytes so repeated generation af 'random' slices is
// cheap.  acts as a ring which can be read from (although doesn't implement io.Reader).
var randSource [randSourceSize]byte

func initRandSource(b *testing.B) {
	if randSource[0] == 0 && randSource[1] == 0 && randSource[2] == 0 && randSource[3] == 0 {
		b.Logf("initializing random source data")
		n, err := crand.Read(randSource[:])
		bMust(b, err, "initializing random source")
		if n < len(randSource) {
			b.Fatalf("unable to read enough random source data %d", n)
		}
	}
}

// acts as a simple byte slice generator.
type randSourceCursor int

func newRandSourceCursor() randSourceCursor {
	i := rand.Intn(randSourceSize)
	return randSourceCursor(i)
}

func (c *randSourceCursor) NBytes(n int) []byte {
	i := int(*c)
	if n >= randSourceSize {
		panic("rand size too big")
	}
	*c = (*c + randSourceCursor(n)) % randSourceSize
	_n := i + n - randSourceSize
	if _n > 0 {
		p := make([]byte, n)
		m := copy(p, randSource[i:])
		copy(p[m:], randSource[:])
		return p
	}
	return randSource[i : i+n]
}
