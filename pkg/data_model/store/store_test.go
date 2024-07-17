package store

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/cockroachdb/pebble"
)

func InitStorage(nameSpaceId types.NamespaceId) *Store[uint64, uint64, uint8, []byte, string] {

	payloadRefDb, err := pebble.Open("willow/payloadrefcounter", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}

	payloadRefKVstore := kv_driver.KvDriver{Db: payloadRefDb}
	PayloadReferenceCounter := payloadDriver.PayloadReferenceCounter{
		Store: payloadRefKVstore,
	}

	entryDb, err := pebble.Open("willow/entries", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	entryKvStore := kv_driver.KvDriver{Db: entryDb}

	TestPayloadDriver := payloadDriver.MakePayloadDriver("willow/payload", TestPayloadScheme)

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

	return &Store[uint64, uint64, uint8, []byte, string]{
		Schemes:            StoreSchemes,
		EntryDriver:        entryDriver,
		PayloadDriver:      TestPayloadDriver,
		NameSpaceId:        nameSpaceId,
		IngestionMutexLock: sync.Mutex{},
		PrefixDriver:       TestPrefixDriver,
	}
}

var TestStore *Store[uint64, uint64, uint8, []byte, string] = InitStorage([]byte("Test"))

func TestSet(t *testing.T) {
	tc := []struct {
		input    datamodeltypes.EntryInput
		authOpts []byte
	}{
		{
			input: datamodeltypes.EntryInput{
				Subspace:  []byte("Samarth"),
				Payload:   []byte("Samarth is a 5th sem student at PES University, now interning at PIL"),
				Timestamp: uint64(time.Now().UnixMicro()) - 200,
				Path:      types.Path{[]byte("intro"), []byte("to"), []byte("samarth")},
			},
			authOpts: []byte("Samarth"),
		},
		{
			input: datamodeltypes.EntryInput{
				Subspace:  []byte("Samarth"),
				Payload:   []byte("Samarth is a 5th sem gandu"),
				Timestamp: uint64(time.Now().UnixMicro()),
				Path:      types.Path{[]byte("intro"), []byte("to")},
			},
			authOpts: []byte("Samarth"),
		},
		{
			input: datamodeltypes.EntryInput{
				Subspace:  []byte("Samar"),
				Payload:   []byte("Samarth is a 5th sem student at PES University, now interning at PIL"),
				Timestamp: uint64(time.Now().UnixMicro()) - 200,
				Path:      types.Path{[]byte("intro"), []byte("to"), []byte("samarth")},
			},
			authOpts: []byte("Samar"),
		},
		{
			input: datamodeltypes.EntryInput{
				Subspace:  []byte("Manas"),
				Payload:   []byte("Manas is a crazy gigachad with big muscles and a small bike"),
				Timestamp: uint64(time.Now().UnixMicro()),
				Path:      types.Path{[]byte("intro"), []byte("to")},
			},
			authOpts: []byte("Manas"),
		},
	}
	TestStore.Storage = TestStore.EntryDriver.MakeStorage([]byte("Test"))
	for _, cases := range tc {
		// fmt.Println(utils.OrderBytes(first, second))
		returnedValue := TestStore.Set(cases.input, cases.authOpts)
		fmt.Println("Pruned Entries: ", returnedValue)
		fmt.Println("============================")
		entry := TestStore.Storage.Get(cases.input.Subspace, cases.input.Path)
		fmt.Println("============================")
		fmt.Println("Entry")
		fmt.Printf("Subspace: %s Path: %v Timestamp: %v\n", entry.Subspace, entry.Path, entry.Time)
		fmt.Println("============================")
		encodedKey, err := kv_driver.EncodeKey(entry.Time, entry.Subspace, TestStore.Schemes.PathParams, entry.Path)
		if err != nil {
			log.Fatal(err)
		}
		encodedValue, err := TestStore.EntryDriver.Opts.KVDriver.Get(encodedKey)
		if err != nil {
			log.Fatal(err)
		}
		payloadLength, payloadDigest, authDigest := kv_driver.DecodeValues(encodedValue)
		fmt.Println("============================")
		fmt.Println("Values from db")
		fmt.Printf("PayloadLength: %v PayloadDigest: %v AuthDigest: %v\n", payloadLength, payloadDigest, authDigest)
		fmt.Println("============================")

		payload, err := TestStore.PayloadDriver.Get(payloadDigest)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("============================")
		fmt.Println("Payload")
		fmt.Printf("%s\n", payload.Bytes())
		fmt.Println("============================")
	}
}
