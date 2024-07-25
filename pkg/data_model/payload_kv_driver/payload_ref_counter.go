package payloadDriver

import (
	"encoding/binary"
	"strings"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

/*
Struct payload reference counter!
Contains the Database inside which payload reference count is persisted
Stores payloadDigest: count as key value
*/
type PayloadReferenceCounter[T constraints.Unsigned] struct {
	Store kv_driver.KvDriver[T]
}

/*
Function takes in payloadDigest as a field
Checks if the value already exists, if it doesnt exist,
we assume it's a new entry and set the count to 1 while
updating the same in the database
If it does exist, then it increments the value and updates in database
*/
func (p *PayloadReferenceCounter[T]) Increment(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	buf := make([]byte, 8)
	if err != nil && strings.Compare(err.Error(), "pebble: not found") != 0 {
		return 0, err
	} else if err != nil && strings.Compare(err.Error(), "pebble: not found") == 0 {
		currCount = 1
	} else {
		currCount = binary.BigEndian.Uint64(currCountBytes) + 1
	}
	binary.BigEndian.PutUint64(buf, currCount)
	p.Store.Set([]byte(payloadDigest), buf)
	return currCount, nil
}

/*
Function takes in payloadDigest as a field
Checks if the value already exists, if it doesnt exist,
it returns an error as you cannot decrement a value which does not exist!
If it does exist, then it decrements the value and updates in database
*/
func (p *PayloadReferenceCounter[T]) Decrement(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	buf := make([]byte, 8)
	if err != nil {
		return 0, err
	} else {
		currCount = binary.BigEndian.Uint64(currCountBytes) - 1
	}
	binary.BigEndian.PutUint64(buf, currCount)
	p.Store.Set([]byte(payloadDigest), buf)
	return currCount, nil
}

/*
Function checks if payloadDigest count key value pair exists, if it does not exist then it returns an error,
if it does exist it returns the!
*/
func (p *PayloadReferenceCounter[T]) Count(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	if err != nil {
		return 0, err
	} else {
		currCount = binary.BigEndian.Uint64(currCountBytes)
	}
	return currCount, nil
}
