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

type KvDriver[KeyPart datamodeltypes.KvPart] struct {
	Db            *pebble.DB
}

func (k *KvDriver[KeyPart])IsFirstPrefixOfSecond(a, b datamodeltypes.KvKey[KeyPart]) (bool, error) {
	if len(a.Key) > len(b.Key) {
		return false, nil
	}
	for index, component := range a.Key {
		res, err := k.CompareTwoKeyParts(component, b.Key[index])
		if err != nil {
			return false, err
		}
		if res != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (k *KvDriver[KeyPart])CompareTwoKeyParts(a, b KeyPart) (types.Rel, error) {
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

func (k *KvDriver[KeyPart])CompareKeys(a, b datamodeltypes.KvKey[KeyPart]) (types.Rel, error) {
	if len(a.Key) > len(b.Key) {
		return 1, nil
	} else if len(a.Key) < len(b.Key) {
		return -1, nil
	} else {
		for i, ele := range a.Key {
			res, err := k.CompareTwoKeyParts(ele, b.Key[i])
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

func (k *KvDriver[KeyPart])Close() error {
	err := k.Db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[KeyPart])Get(key []byte) ([]byte, error) {
	value, closer, err := k.Db.Get(key)
	defer closer.Close()
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (k *KvDriver[KeyPart])Set(key, value []byte) error {
	err := k.Db.Set(key, value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[KeyPart])Delete(key []byte) error {
	err := k.Db.Delete(key, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[KeyPart])Clear() error {
	iter, err := k.Db.NewIter(nil)
	if err != nil {
		return err
	}
	for iter.First(); iter.Valid(); iter.Next() {
		err := k.Db.Delete(iter.Key(), pebble.Sync)
		if err != nil {
			return err
		}
	}
	if err := iter.Close(); err != nil {
		return errors.New("failed to close the iterator")
	}
	return nil
}

func (k *KvDriver[KeyPart])ListAllValues() ([]struct {
	Key   []byte
	Value []byte
}, error,
) {
	var values []struct {
		Key   []byte
		Value []byte
	}
	iter, err := k.Db.NewIter(nil)
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
		value, closer, err := k.Db.Get(iter.Key())
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

func (k *KvDriver[KeyPart])Batch() (*pebble.Batch, error) {
	batch := k.Db.NewBatch()
	return batch, nil
}

func (k *KvDriver[KeyPart])GetMultipleValues(queryArray [][]byte)([][]byte, error){
	queryValues := make([][]byte, 0, len(queryArray))
	for _, key := range queryArray {
		value, err := k.Get(key)
		if err != nil {
			return nil, err
		}
		queryValues = append(queryValues, value)
	}
	return queryValues, nil
}