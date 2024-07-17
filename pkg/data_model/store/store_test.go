package store

import (
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
)

func InitStorage() *Store[uint64, uint64, uint8, any, string] {
	return &Store[uint64, uint64, uint8, any, string]{
		Schemes: StoreSchemes,
		EntryDriver: ,
	}

}

func TestSet(t *testing.T) {

}
