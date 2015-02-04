package mdb

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Most mdb functions/methods can return errors. This example ignores errors
// for brevity. Real code should check all return values.
func Example() {
	// create a directory to hold the database
	path, _ := ioutil.TempDir("", "mdb_test")
	defer os.RemoveAll(path)

	// open the db
	env, _ := NewEnv()
	env.SetMapSize(1 << 20) // max file size
	env.Open(path, 0, 0664)
	defer env.Close()
	txn, _ := env.BeginTxn(nil, 0)
	dbi, _ := txn.DBIOpen(nil, 0)
	defer env.DBIClose(dbi)
	txn.Commit()

	// write some data
	txn, _ = env.BeginTxn(nil, 0)
	num_entries := 5
	for i := 0; i < num_entries; i++ {
		key := fmt.Sprintf("Key-%d", i)
		val := fmt.Sprintf("Val-%d", i)
		txn.Put(dbi, []byte(key), []byte(val), 0)
	}
	txn.Commit()

	// inspect the database
	stat, _ := env.Stat()
	fmt.Println(stat.Entries)

	// scan the database
	txn, _ = env.BeginTxn(nil, RDONLY)
	defer txn.Abort()
	cursor, _ := txn.CursorOpen(dbi)
	defer cursor.Close()
	for {
		bkey, bval, err := cursor.Get(nil, nil, NEXT)
		if err == NotFound {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %s\n", bkey, bval)
	}

	// random access
	bval, _ := txn.Get(dbi, []byte("Key-3"))
	fmt.Println(string(bval))

	// Output:
	// 5
	// Key-0: Val-0
	// Key-1: Val-1
	// Key-2: Val-2
	// Key-3: Val-3
	// Key-4: Val-4
	// Val-3
}
