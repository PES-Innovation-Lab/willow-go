package entrydriver

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

// All the necesarry functions and options requires along with the EntryDriver struct!!!
type EntryDriver[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	PayloadReferenceCounter payloadDriver.PayloadReferenceCounter
	Storage            datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]
	// GetPayloadLength        func(digest types.PayloadDigest) uint64 why do we need this again????
	Opts struct {
		KVDriver          kv_driver.KvDriver
		NamespaceScheme   datamodeltypes.NamespaceScheme
		SubspaceScheme    datamodeltypes.SubspaceScheme
		PayloadScheme     datamodeltypes.PayloadScheme
		PathParams        types.PathParams[K]
		FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
	}
}

/*
Instantiates a new KD tree and then returns the KD tree
Thos function will be used when we want to instantiate a new KD tree at the start of the application
*/
func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) MakeStorage(nameSpaceId types.NamespaceId, dbValues []Kdtree.KDNodeKey) datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K] {
	storage := datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]{
		KDTree: Kdtree.NewKDTreeWithValues[Kdtree.KDNodeKey](3, dbValues),
		Opts: struct {
			Namespace         types.NamespaceId
			SubspaceScheme    datamodeltypes.SubspaceScheme
			PayloadScheme     datamodeltypes.PayloadScheme
			PathParams        types.PathParams[K]
			FingerprintScheme datamodeltypes.FingerprintScheme[PreFingerPrint, FingerPrint]
		}{
			Namespace:         nameSpaceId,
			SubspaceScheme:    e.Opts.SubspaceScheme,
			PayloadScheme:     e.Opts.PayloadScheme,
			PathParams:        e.Opts.PathParams,
			FingerprintScheme: e.Opts.FingerprintScheme,
		},
	}
	return storage
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Get(Subspace types.SubspaceId, Path types.Path) (struct {
	Entry         types.Entry
	AuthTokenHash types.PayloadDigest
}, error) {
	entryExists := e.Storage.Get(Subspace, Path)
	if reflect.DeepEqual(entryExists, types.Position3d{}) {
		return struct{Entry types.Entry; AuthTokenHash types.PayloadDigest}{}, errors.New("entry does not exist")
	}
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{
		Time:     entryExists.Time,
		Subspace: entryExists.Subspace,
		Path: entryExists.Path,
	}, e.Opts.PathParams)

	if err != nil {
		return struct{Entry types.Entry; AuthTokenHash types.PayloadDigest}{}, err
	}
	entryBytes, err := e.Opts.KVDriver.Get(encodedKey)
	fmt.Println("Got entry from pebble")
	if err != nil {
		return struct{Entry types.Entry; AuthTokenHash types.PayloadDigest}{}, err
	}
	value := kv_driver.DecodeValues(entryBytes)
	entry := types.Entry{
		Timestamp: entryExists.Time,
		Path: 	Path,
		Subspace_id: 	Subspace,
		Payload_digest: value.PayloadDigest,
		Payload_length: value.PayloadLength,
		Namespace_id: e.Storage.Opts.Namespace,
	}
	
	return struct{Entry types.Entry; AuthTokenHash types.PayloadDigest}{
		Entry: entry, 
		AuthTokenHash: value.AuthDigest}, nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Insert(entry types.Entry, authDigest types.PayloadDigest) error {
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{Time: entry.Timestamp, Subspace: entry.Subspace_id, Path: entry.Path}, e.Opts.PathParams)
	if err != nil {
		return err
	}
	encodedValue := kv_driver.EncodeValues(struct{PayloadLength uint64; PayloadDigest types.PayloadDigest; AuthDigest types.PayloadDigest}{
		PayloadLength: entry.Payload_length,
		PayloadDigest: entry.Payload_digest,
		AuthDigest: authDigest,
	})
	err = e.Opts.KVDriver.Set(encodedKey, encodedValue)
	if err != nil {
		return err
	}
	err = e.Storage.Insert(entry.Subspace_id, entry.Path, entry.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K])Delete(entry types.Entry) error {
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{Time: entry.Timestamp, Subspace: entry.Subspace_id, Path: entry.Path}, e.Opts.PathParams)
	if err != nil {
		return err
	}

	err = e.Opts.KVDriver.Delete(encodedKey)
	if err != nil {
		return err
	}
	if !(e.Storage.Remove(types.Position3d{Time: entry.Timestamp, Subspace: entry.Subspace_id, Path: entry.Path})){
		return errors.New("entry does not exist in the KD Tree")
	}
	return nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Query(range3d types.Range3d) ([]types.Entry, error){
	entryNodes := e.Storage.Query(range3d)
	Entries := make([]types.Entry,len(entryNodes))
	for _,node := range entryNodes {
		encodedKey,err := kv_driver.EncodeKey(types.Position3d{Time: node.Timestamp, Subspace: node.Subspace, Path: node.Path},e.Opts.PathParams)
		if err!=nil {
			log.Fatalln(err, "can't Encode key")
		}

		encodedValue,err:= e.Opts.KVDriver.Get(encodedKey)
		if err!=nil {
			return []types.Entry{},err
		}

		decodedValue:= kv_driver.DecodeValues(encodedValue)
		entry:=types.Entry{
			Timestamp: node.Timestamp,
			Path: 	node.Path,
			Subspace_id: 	node.Subspace,
			Payload_digest: decodedValue.PayloadDigest,
			Payload_length: decodedValue.PayloadLength,
			Namespace_id: e.Storage.Opts.Namespace,
		}
		Entries = append(Entries,entry)

	}
	return Entries,nil
}