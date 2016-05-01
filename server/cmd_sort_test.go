package server

import (
	"fmt"
	"testing"

	"github.com/siddontang/goredis"
)

func checkTestSortRes(ay interface{}, checks []string) error {
	values, ok := ay.([]interface{})
	if !ok {
		return fmt.Errorf("invalid res type %T, must [][]byte", ay)
	}

	if len(values) != len(checks) {
		return fmt.Errorf("invalid res number %d != %d", len(values), len(checks))
	}

	for i := range values {
		if string(values[i].([]byte)) != checks[i] {
			return fmt.Errorf("invalid res at %d, %s != %s", i, values[i], checks[i])
		}
	}
	return nil
}

func TestSort(t *testing.T) {
	c := getTestConn()
	defer c.Close()

	key := "my_sort_key"
	storeKey := "my_sort_store_key"

	if _, err := c.Do("LPUSH", key, 3, 2, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Do("MSET", "weight_1", 3, "weight_2", 2, "weight_3", 1); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Do("MSET", "object_1", 10, "object_2", 20, "object_3", 30); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"1", "2", "3"}); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key, "DESC"); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"3", "2", "1"}); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key, "LIMIT", 0, 1); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"1"}); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key, "BY", "weight_*"); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"3", "2", "1"}); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key, "GET", "object_*"); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"10", "20", "30"}); err != nil {
		t.Fatal(err)
	}

	if ay, err := c.Do("XLSORT", key, "GET", "object_*", "GET", "#"); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"10", "1", "20", "2", "30", "3"}); err != nil {
		t.Fatal(err)
	}

	if n, err := goredis.Int(c.Do("XLSORT", key, "STORE", storeKey)); err != nil {
		t.Fatal(err)
	} else if n != 3 {
		t.Fatalf("invalid return store sort number, %d != 3", n)
	} else if ay, err := c.Do("LRANGE", storeKey, 0, -1); err != nil {
		t.Fatal(err)
	} else if err = checkTestSortRes(ay, []string{"1", "2", "3"}); err != nil {
		t.Fatal(err)
	}
}
