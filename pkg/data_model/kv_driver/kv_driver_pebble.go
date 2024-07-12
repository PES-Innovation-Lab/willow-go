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

// import (
// 	"errors"
// 	"log"
// 	"reflect"
// 	"strings"

// 	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
// 	"github.com/PES-Innovation-Lab/willow-go/types"
// 	"github.com/PES-Innovation-Lab/willow-go/utils"
// 	"github.com/cockroachdb/pebble"
// )

func isFirstPrefixOfSecond[T datamodeltypes.KvPart](a, b datamodeltypes.KvKey[T]) (bool, error) {
	if len(a.Key) > len(b.Key) {
		return false, nil
	}
	for index, component := range a.Key {
		res, err := CompareTwoKeyParts(component, b.Key[index])
		if err != nil {
			return false, err
		}
		if res != 0 {
			return false, nil
		}
	}
	return true, nil
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
	return 0, errors.New("the type of KV part is not matching with allowed types")
}

func CompareKeys[T datamodeltypes.KvPart](a, b datamodeltypes.KvKey[T]) (types.Rel, error) {
	if len(a.Key) > len(b.Key) {
		return 1, nil
	} else if len(a.Key) < len(b.Key) {
		return -1, nil
	} else {
		for i, ele := range a.Key {
			res, err := CompareTwoKeyParts(ele, b.Key[i])
			if err != nil {
				return 0, err
			}
			if res != 0 {
				return res, nil
			}
		}
	}
	return 0, nil
}

// func Close(Db *pebble.DB) error {
// 	err := Db.Close()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func Get(Db *pebble.DB, key []byte) ([]byte, error) {
	value, closer, err := Db.Get(key)
	defer closer.Close()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// func Set(Db *pebble.DB, key, value []byte) error {
// 	err := Db.Set(key, value, pebble.Sync)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func Delete(Db *pebble.DB, key []byte) error {
// 	err := Db.Delete(key, pebble.Sync)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func Clear(Db *pebble.DB) error {
	iter, err := Db.NewIter(nil)
	if err != nil {
		return err
	}
	for iter.First(); iter.Valid(); iter.Next() {
		err := Db.Delete(iter.Key(), pebble.Sync)
		if err != nil {
			return err
		}
	}
	if err := iter.Close(); err != nil {
		return errors.New("failed to close the iterator")
	}
	return nil
}

func ListAllValues(Db *pebble.DB) ([]struct {
	Key   []byte
	Value []byte
}, error,
) {
	var values []struct {
		Key   []byte
		Value []byte
	}
	iter, err := Db.NewIter(nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := iter.Close(); err != nil {
			log.Fatal("error in closing the iter")
		}
	}()
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value, closer, err := Db.Get(iter.Key())
		if err != nil {
			return nil, err
		}
		closer.Close()
		values = append(values, struct {
			Key   []byte
			Value []byte
		}{Key: key, Value: value})
	}
	return values, nil
}

func Batch(Db *pebble.DB) (*pebble.Batch, error) {
	batch := Db.NewBatch()
	return batch, nil
}

func CreateEntryDriver[T datamodeltypes.KvPart](Db *pebble.DB) (datamodeltypes.KvDriver, error) {
	entryDriver := datamodeltypes.KvDriver{
		Db:            Db,
		Get:           Get,
		Set:           Set,
		Clear:         Clear,
		ListAllValues: ListAllValues,
		Batch:         Batch,
	}
	return entryDriver, nil
}
