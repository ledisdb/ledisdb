package ledis

import (
	"runtime/debug"
	"testing"
)

func testBuildValues(values ...string) [][]byte {
	v := make([][]byte, 0, len(values))
	for _, value := range values {
		v = append(v, []byte(value))
	}
	return v
}

func checkSortRes(t *testing.T, res [][]byte, values ...string) {
	if len(res) != len(values) {
		debug.PrintStack()
		t.Fatalf("invalid xsort res len, %d = %d", len(res), len(values))
	}

	for i := 0; i < len(res); i++ {
		if string(res[i]) != values[i] {
			debug.PrintStack()
			t.Fatalf("invalid xsort res at %d, %s != %s", i, res[i], values[i])
		}
	}
}

func checkTestSort(t *testing.T, db *DB, items []string, offset int, size int, alpha bool,
	desc bool, sortBy []byte, sortGet [][]byte, checkValues []string) {

	vv := testBuildValues(items...)

	res, err := db.xsort(vv, offset, size, alpha, desc, sortBy, sortGet)
	if err != nil {
		t.Fatal(err)
	}

	checkSortRes(t, res, checkValues...)
}

func TestSort(t *testing.T) {
	db := getTestDB()

	db.FlushAll()

	// Prepare data
	db.MSet(
		KVPair{[]byte("weight_1"), []byte("30")},
		KVPair{[]byte("weight_2"), []byte("20")},
		KVPair{[]byte("weight_3"), []byte("10")},
		KVPair{[]byte("weight_a"), []byte("60")},
		KVPair{[]byte("weight_b"), []byte("50")},
		KVPair{[]byte("weight_c"), []byte("40")})

	db.HSet([]byte("hash_weight_1"), []byte("index"), []byte("30"))
	db.HSet([]byte("hash_weight_2"), []byte("index"), []byte("20"))
	db.HSet([]byte("hash_weight_3"), []byte("index"), []byte("10"))
	db.HSet([]byte("hash_weight_a"), []byte("index"), []byte("60"))
	db.HSet([]byte("hash_weight_b"), []byte("index"), []byte("50"))
	db.HSet([]byte("hash_weight_c"), []byte("index"), []byte("40"))

	db.MSet(
		KVPair{[]byte("object_1"), []byte("30")},
		KVPair{[]byte("object_2"), []byte("20")},
		KVPair{[]byte("object_3"), []byte("10")},
		KVPair{[]byte("number_1"), []byte("10")},
		KVPair{[]byte("number_2"), []byte("20")},
		KVPair{[]byte("number_3"), []byte("30")},
		KVPair{[]byte("object_a"), []byte("60")},
		KVPair{[]byte("object_b"), []byte("50")},
		KVPair{[]byte("object_c"), []byte("40")},
		KVPair{[]byte("number_a"), []byte("40")},
		KVPair{[]byte("number_b"), []byte("50")},
		KVPair{[]byte("number_c"), []byte("60")})

	db.HSet([]byte("hash_object_1"), []byte("index"), []byte("30"))
	db.HSet([]byte("hash_object_2"), []byte("index"), []byte("20"))
	db.HSet([]byte("hash_object_3"), []byte("index"), []byte("10"))
	db.HSet([]byte("hash_number_1"), []byte("index"), []byte("10"))
	db.HSet([]byte("hash_number_2"), []byte("index"), []byte("20"))
	db.HSet([]byte("hash_number_3"), []byte("index"), []byte("30"))

	db.HSet([]byte("hash_object_a"), []byte("index"), []byte("60"))
	db.HSet([]byte("hash_object_b"), []byte("index"), []byte("50"))
	db.HSet([]byte("hash_object_c"), []byte("index"), []byte("40"))
	db.HSet([]byte("hash_number_a"), []byte("index"), []byte("40"))
	db.HSet([]byte("hash_number_b"), []byte("index"), []byte("50"))
	db.HSet([]byte("hash_number_c"), []byte("index"), []byte("60"))

	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, nil, []string{"1", "2", "3"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, 1, false, false, nil, nil, []string{"1"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, 2, false, false, nil, nil, []string{"1", "2"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, 3, false, false, nil, nil, []string{"1", "2", "3"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, 4, false, false, nil, nil, []string{"1", "2", "3"})

	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, true, false, nil, nil, []string{"1", "2", "3"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, true, nil, nil, []string{"3", "2", "1"})

	if _, err := db.xsort(testBuildValues("c", "b", "a"), 0, -1, false, false, nil, nil); err == nil {
		t.Fatal("must nil")
	}

	checkTestSort(t, db, []string{"c", "b", "a"}, 0, -1, true, false, nil, nil, []string{"a", "b", "c"})
	checkTestSort(t, db, []string{"c", "b", "a"}, 0, -1, true, true, nil, nil, []string{"c", "b", "a"})

	checkTestSort(t, db, []string{"1", "2", "3"}, 0, 1, false, false, nil, nil, []string{"1"})
	checkTestSort(t, db, []string{"1", "2", "3"}, 0, 1, false, true, nil, nil, []string{"3"})

	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, []byte("abc"), nil, []string{"3", "2", "1"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, []byte("weight_*"), nil, []string{"3", "2", "1"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, true, []byte("weight_*"), nil, []string{"1", "2", "3"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, []byte("hash_weight_*->index"), nil, []string{"3", "2", "1"})

	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, [][]byte{[]byte("object_*")}, []string{"30", "20", "10"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, [][]byte{[]byte("number_*")}, []string{"10", "20", "30"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, [][]byte{[]byte("#"), []byte("number_*")},
		[]string{"1", "10", "2", "20", "3", "30"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, [][]byte{[]byte("object_*"), []byte("number_*")},
		[]string{"30", "10", "20", "20", "10", "30"})
	checkTestSort(t, db, []string{"3", "2", "1"}, 0, -1, false, false, nil, [][]byte{[]byte("object_*_abc")}, []string{"", "", ""})
}
