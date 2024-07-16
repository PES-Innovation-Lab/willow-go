package kv_driver

import (
	"log"
	"testing"

	"github.com/cockroachdb/pebble"
)

func TestSetGet(t *testing.T) {
	tc := []struct {
		key      string
		value    string
		expected string
	}{
		{
			key:      "hello",
			value:    "value",
			expected: "value",
		},
		{
			key:      "123124",
			value:    "4242121",
			expected: "4242121",
		},
	}
	db, err := pebble.Open("demo", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	k := KvDriver[[]byte]{Db: db}
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, cases := range tc {
		err := k.Set([]byte(cases.key), []byte(cases.value))
		if err != nil {
			log.Fatal(err)
		}
		value, err := k.Get([]byte(cases.key))
		if err != nil {
			log.Fatal(err)
		}
		if string(value) != cases.expected {
			log.Printf("expected: %s\ngot:%s", cases.expected, string(value))
		}
	}
}

func TestListAllClear(t *testing.T) {
	tc := []struct {
		keyValues []struct {
			key   string
			value string
		}
		expected1 []struct {
			key   string
			value string
		}
		expected2 []struct {
			key   string
			value string
		}
	}{
		{
			[]struct {
				key   string
				value string
			}{
				{key: "123124", value: "4242121"},
				{key: "hello", value: "world"},
			},
			[]struct {
				key   string
				value string
			}{
				{key: "123124", value: "4242121"},
				{key: "hello", value: "world"},
			},
			[]struct {
				key   string
				value string
			}{},
		},
		{
			[]struct {
				key   string
				value string
			}{
				{key: "112easd", value: "4fdaasd"},
				{key: "123124", value: "4242121"},
				{key: "Vibhav", value: "Vinay"},
				{key: "hello", value: "world"},
			},
			[]struct {
				key   string
				value string
			}{
				{key: "112easd", value: "4fdaasd"},
				{key: "123124", value: "4242121"},
				{key: "Vibhav", value: "Vinay"},
				{key: "hello", value: "world"},
			},
			[]struct {
				key   string
				value string
			}{},
		},
	}
	db, err := pebble.Open("demo", &pebble.Options{})
	k := &KvDriver[[]byte]{Db: db}
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, cases := range tc {
		for _, pairs := range cases.keyValues {
			err := db.Set([]byte(pairs.key), []byte(pairs.value), pebble.Sync)
			t.Logf("Inserted: %s, %s", pairs.key, string([]byte(pairs.value)))
			if err != nil {
				t.Fatal(err)
			}
		}
		values, err := k.ListAllValues()
		if err != nil {
			t.Fatal(err)
		}
		for index, ele := range cases.expected1 {
			if ele.key != string(values[index].Key) && ele.value != string(values[index].Value) {
				t.Fatalf("index: %d\nexpected: (%s, %s)\n got: (%s, %s)\n%v", index, ele.key, ele.value, string(values[index].Key), string(values[index].Value), values)
			}
		}
		k.Clear()
		values1, _ := k.ListAllValues()

		if len(values1) > 0 {
			t.Fatalf("%v", values1)
		}
	}
}
