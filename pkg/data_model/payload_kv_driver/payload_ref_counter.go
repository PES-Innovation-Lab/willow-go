package payloadDriver

import (
	"encoding/binary"
	"strings"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
)

type PayloadReferenceCounter struct {
	Store kv_driver.KvDriver
}

func (p *PayloadReferenceCounter)Increment(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	var buf []byte
	if err != nil && strings.Compare(err.Error(), "pebble: not found") != 0{
		return 0, err
	}else if err != nil && strings.Compare(err.Error(), "pebble: not found") == 0 {
		currCount = 1
	}else {
		currCount = binary.BigEndian.Uint64(currCountBytes) + 1
	}
	binary.BigEndian.AppendUint64(buf, currCount)
	p.Store.Set([]byte(payloadDigest), buf)
	return currCount, nil
}

func (p *PayloadReferenceCounter)Decrement(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	var buf []byte
	if err != nil {
		return 0, err
	}else {
		currCount = binary.BigEndian.Uint64(currCountBytes) - 1
	}
	binary.BigEndian.AppendUint64(buf, currCount)
	p.Store.Set([]byte(payloadDigest), buf)
	return currCount, nil
}

func (p *PayloadReferenceCounter)Count(payloadDigest types.PayloadDigest) (uint64, error) {
	currCountBytes, err := p.Store.Get([]byte(payloadDigest))
	var currCount uint64
	if err != nil {
		return 0, err
	}else {
		currCount = binary.BigEndian.Uint64(currCountBytes)
	}
	return currCount, nil
}

