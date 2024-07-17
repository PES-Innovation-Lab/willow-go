package kv_driver

import (
	"errors"
	"log"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/cockroachdb/pebble"
)

type KvDriver struct {
	Db            *pebble.DB
}

func (k *KvDriver)IsFirstPrefixOfSecond(a, b []byte) (bool, error) {
	if len(a) > len(b) {
		return false, nil
	}
	for i, component := range a {
		res, err := k.CompareTwoKeyParts(component, b[i])
		if err != nil {
			return false, err
		}
		if res != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (k *KvDriver)CompareTwoKeyParts(a, b byte) (types.Rel, error) {
	if a > b{
		return 1, nil
	} else if a < b {
		return -1, nil
	} else {
		return 0, nil
	}
}

func (k *KvDriver)CompareKeys(a, b []byte) (types.Rel, error) {
	if len(a) > len(b) {
		return 1, nil
	} else if len(a) < len(b) {
		return -1, nil
	} else {
		for i, ele := range a {
			res, err := k.CompareTwoKeyParts(ele, b[i])
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

func (k *KvDriver)Close() error {
	err := k.Db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver)Get(key []byte) ([]byte, error) {
	value, closer, err := k.Db.Get(key)
	if err != nil {
		return []byte{}, err
	}
	closer.Close()
	return value, nil
}

func (k *KvDriver)Set(key, value []byte) error {
	err := k.Db.Set(key, value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver)Delete(key []byte) error {
	err := k.Db.Delete(key, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver)Clear() error {
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

func (k *KvDriver)ListAllValues() ([]struct {
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

func (k *KvDriver)Batch() (*pebble.Batch, error) {
	batch := k.Db.NewBatch()
	return batch, nil
}

func (k *KvDriver)GetMultipleValues(queryArray [][]byte)([][]byte, error){
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