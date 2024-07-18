package PinaGolada

import (
	"fmt"
	"log"
	"sync"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/store"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/cockroachdb/pebble"
).

func InitStorage(nameSpaceId types.NamespaceId) *store.Store[uint64, uint64, uint8, []byte, string] {

	payloadRefDb, err := pebble.Open(fmt.Sprintf("willow/%s/payloadrefcounter", string(nameSpaceId)), &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}

	payloadRefKVstore := kv_driver.KvDriver{Db: payloadRefDb}
	PayloadReferenceCounter := payloadDriver.PayloadReferenceCounter{
		Store: payloadRefKVstore,
	}

	entryDb, err := pebble.Open(fmt.Sprintf("willow/%s/entries", string(nameSpaceId)), &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	entryKvStore := kv_driver.KvDriver{Db: entryDb}

	PayloadLock := &sync.Mutex{}
	TestPayloadDriver := payloadDriver.MakePayloadDriver(fmt.Sprintf("willow/%s/payload", string(nameSpaceId)), TestPayloadScheme, PayloadLock)

	entryDriver := entrydriver.EntryDriver[uint64, uint64, uint8]{
		PayloadReferenceCounter: PayloadReferenceCounter,
		Opts: struct {
			KVDriver          kv_driver.KvDriver
			NamespaceScheme   datamodeltypes.NamespaceScheme
			SubspaceScheme    datamodeltypes.SubspaceScheme
			PayloadScheme     datamodeltypes.PayloadScheme
			PathParams        types.PathParams[uint8]
			FingerprintScheme datamodeltypes.FingerprintScheme[uint64, uint64]
		}{
			KVDriver:          entryKvStore,
			NamespaceScheme:   TestNameSpaceScheme,
			SubspaceScheme:    TestSubspaceScheme,
			PayloadScheme:     TestPayloadScheme,
			PathParams:        TestPathParams,
			FingerprintScheme: TestFingerprintScheme,
		},
	}
	TestPrefixDriver := kv_driver.PrefixDriver[uint8]{}

	return &store.Store[uint64, uint64, uint8, []byte, string]{
		Schemes:            StoreSchemes,
		EntryDriver:        entryDriver,
		PayloadDriver:      TestPayloadDriver,
		NameSpaceId:        nameSpaceId,
		IngestionMutexLock: sync.Mutex{},
		PrefixDriver:       TestPrefixDriver,
	}
}



