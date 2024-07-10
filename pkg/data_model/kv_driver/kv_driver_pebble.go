package kv_driver

import (
	"errors"
	"log"
	"reflect"
	"strings"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"github.com/cockroachdb/pebble"
)

func isFirstPrefixOfSecond[T datamodeltypes.KvPart](a, b datamodeltypes.KvKey[T]) bool {
	if len(a.Key) > len(b.Key) {
		return false
	}
    for 
}

func CompareTwoKeyParts[T datamodeltypes.KvPart](a, b T) (types.Rel, error) {
	typeA := reflect.TypeOf(a)
	typeB := reflect.TypeOf(b)
	valueA := reflect.ValueOf(a)
	valueB := reflect.ValueOf(b)
	if typeA == reflect.TypeOf([]byte(nil)) {
		if typeB == reflect.TypeOf([]byte(nil)) {
			utils.OrderBytes(valueA.Bytes(), valueB.Bytes())
		} else {
			return -1, nil
		}
	} else if typeA.Kind() == reflect.String {
		if typeB == reflect.TypeOf([]byte(nil)) {
			return 1, nil
		} else if typeB.Kind() == reflect.String {
			return types.Rel(strings.Compare(valueA.String(), valueB.String())), nil
		} else {
			return -1, nil
		}
	} else {
		if typeB.Kind() == reflect.String || typeB == reflect.TypeOf([]byte(nil)) {
			return 1, nil
		} else {
			if valueA.Float() < valueB.Float() {
				return -1, nil
			} else if valueA.Float() > valueB.Float() {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	}
	return 0, errors.New("The type of KV part is not matching with allowed types!")
}

func CompareKeys[T datamodeltypes.KvPart](a, b datamodeltypes.KvKey[T]) (bool, error) {
	if len(a.Key) > len(b.Key) {
		return false, nil
	} else {
		for i, ele := range a.Key {
			res, err := CompareTwoKeyParts(ele, b.Key[i])
			if err != nil {
				log.Fatal(err)
			}
			if res != 0 {
				return false, nil
			}
			return true, nil
		}
	}
	return false, errors.New("Oops something went wrong!")
}

func Close(Db *pebble.DB) error {
	err := Db.Close()
	if err != nil {
		return err
	}
	return nil
}

func Get(Db *pebble.DB, key []byte) (datamodeltypes.KvValue, error) {
	value, closer, err := Db.Get(key)
	if err != nil {
		return nil, err
	}
	closer.Close()
	return value, nil
}

func Set(Db *pebble.DB, key, value []byte) error {
	err := Db.Set(key, value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func Delete(Db *pebble.DB, key []byte) error {
	err := Db.Delete(key, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func List[T datamodeltypes.KvPart](
	Db *pebble.DB, selector datamodeltypes.ListSelector[T],
	opts datamodeltypes.ListOpts,
) ([]datamodeltypes.EntryIterator[T], error) {
	var reverse bool
	var limit uint
	var batchSize uint
	var prefix datamodeltypes.KvKey[T]
	var start datamodeltypes.KvKey[T]
	var end datamodeltypes.KvKey[T]

	if !reflect.DeepEqual(opts, datamodeltypes.ListOpts{}) {
		reverse = opts.Reverse
		limit = opts.Limit
		batchSize = opts.BatchSize
	} else {
		reverse = false
		limit = 0
		batchSize = 0
	}

	if limit == 0 {
		return nil, nil
	}
	if reflect.DeepEqual(selector.Prefix, datamodeltypes.KvKey[T]{}) {
		prefix = datamodeltypes.KvKey[T]{}
	} else {
		prefix = selector.Prefix
	}

	start = selector.Start
	end = selector.End

	if reflect.DeepEqual(prefix, datamodeltypes.KvKey[T]{}) {
	}
}
