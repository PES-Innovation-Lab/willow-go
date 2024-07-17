package store

import (
	"log"
	"testing"

	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/cockroachdb/pebble"
)

func InitStorage() *Store[uint64, uint64, uint8, any, string] {

	db, err := pebble.Open("demo", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	Kvstore := kv_driver.KvDriver{Db: db}
	PayloadReferenceCounter := payloadDriver.PayloadReferenceCounter{
		Store: Kvstore,
	}

	entryDriver := entrydriver.EntryDriver[uint64, uint64, uint64]{
		PayloadReferenceCounter: PayloadReferenceCounter,
		Opts:                    struct{},
	}

	return &Store[uint64, uint64, uint8, any, string]{
		Schemes: StoreSchemes,
	}
}

func TestSet(t *testing.T) {

}
