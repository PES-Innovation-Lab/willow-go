package kv_driver

import (
	"errors"
	"log"

	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"github.com/cockroachdb/pebble"
	"golang.org/x/exp/constraints"
)

type KvDriver[T constraints.Unsigned] struct {
	Db *pebble.DB
}

func (k *KvDriver[T]) IsFirstPrefixOfSecond(a, b []byte) (bool, error) {
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

func (k *KvDriver[T]) CompareTwoKeyParts(a, b byte) (types.Rel, error) {
	if a > b {
		return 1, nil
	} else if a < b {
		return -1, nil
	} else {
		return 0, nil
	}
}

func (k *KvDriver[T]) CompareKeys(a, b []byte) (types.Rel, error) {
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

func (k *KvDriver[T]) Close() error {
	err := k.Db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[T]) Get(key []byte) ([]byte, error) {
	value, closer, err := k.Db.Get(key)
	if err != nil {
		return []byte{}, err
	}
	closer.Close()
	return value, nil
}

func (k *KvDriver[T]) Set(key, value []byte) error {
	err := k.Db.Set(key, value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[T]) Delete(key []byte) error {
	err := k.Db.Delete(key, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (k *KvDriver[T]) Clear() error {
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

func (k *KvDriver[T]) ListAllValues() ([]struct {
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

func (k *KvDriver[T]) Batch() (*pebble.Batch, error) {
	batch := k.Db.NewBatch()
	return batch, nil
}

func (k *KvDriver[T]) GetMultipleValues(queryArray [][]byte) ([][]byte, error) {
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

func (k *KvDriver[T]) ListValues(aoi types.AreaOfInterest, params types.PathParams[T], nameSpaceId types.NamespaceId) ([]types.Entry, error) {
	//To calculate the payload length so that it does not exceed MaxPayloadLength defined bu aoi
	var payloadLength uint64
	//Variable to store entries in the aoi
	var values []types.Entry

	//Creatign iter for DB
	iter, err := k.Db.NewIter(nil)
	if err != nil {
		return nil, err
	}

	//Closing function of iter
	defer func() {
		if err := iter.Close(); err != nil {
			log.Fatal("error in closing the iter")
		}
	}()

	//iterating from last to first, we are extracting newest x entries(mentioned in the aoi) from the DB and iterating from bottom up
	//iterating bottom up ensures we encounter the newest entries first
	for iter.Last(); iter.Valid(); iter.Prev() {
		//Gets the key in the DB
		encodedKey := iter.Key()
		//decodes the key
		timestamp, subspace, path, err := DecodeKey(encodedKey, params)

		//return error if any
		if err != nil {
			return nil, err
		}

		//Checks if the timestamp is lesser than the mentioned end
		//Checks if the subspace matches
		//Checks if the path is a suffix of the path in the aoi
		if (timestamp > aoi.Area.Times.End) || utils.OrderSubspace(subspace, aoi.Area.Subspace_id) != 0 || len(path) < len(aoi.Area.Path) {
			continue
		}
		//Checking for prefix
		if utils.OrderPath(path[:len(aoi.Area.Path)], aoi.Area.Path) != 0 {
			continue
		}

		//If the Max_count is defined in the aoi, then we check if the values are exceeding the limit
		//Undefined Max_count or Max_size are 0 values
		//If the values exceed the limit, we break the loop
		//We also check if we go past the lower time constraint, if we do we break the loop
		//Since the keys are ordered wrt to timestamp, this is possible
		if (aoi.Max_count > 0 && len(values) >= int(aoi.Max_count)) || timestamp < aoi.Area.Times.Start {
			break
		}

		//Gets the entry valyeues from the DB
		encodedValue, closer, err := k.Db.Get(encodedKey)
		if err != nil {
			return nil, err
		}

		//Decodes the values
		value := DecodeValues(encodedValue)
		//Adds the payload length to the variable to check Max_size
		payloadLength += value.PayloadLength

		//If the Max_size is defined in the aoi, then we check if the values are exceeding the limit
		if aoi.Max_size > 0 && payloadLength >= aoi.Max_size {
			break
		}

		//Appends the values to the entries
		values = append(values, types.Entry{
			Timestamp:      timestamp,
			Subspace_id:    subspace,
			Path:           path,
			Payload_digest: value.PayloadDigest,
			Payload_length: value.PayloadLength,
			Namespace_id:   nameSpaceId,
		})

		//Closes the value
		closer.Close()
	}
	return values, nil
}
