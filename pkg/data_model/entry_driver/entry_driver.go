package entrydriver

import (
	"errors"
	"log"
	"reflect"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	kdtree "github.com/rishitc/go-kd-tree"
	"golang.org/x/exp/constraints"
)

// All the necesarry functions and options requires along with the EntryDriver struct!!!
type EntryDriver[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned] struct {
	PayloadReferenceCounter payloadDriver.PayloadReferenceCounter
	Storage                 datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]
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
func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) MakeStorage(nameSpaceId types.NamespaceId, dbValues []kdnode.Key) datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K] {
	storage := datamodeltypes.KDTreeStorage[PreFingerPrint, FingerPrint, K]{
		KDTree: kdtree.NewKDTreeWithValues[kdnode.Key](3, dbValues),
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

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Get(Subspace types.SubspaceId, Path types.Path) (datamodeltypes.ExtendedEntry, error) {

	entryExists, err := e.Storage.Get(Subspace, Path)
	if err != nil {
		return datamodeltypes.ExtendedEntry{}, err
	}

	if reflect.DeepEqual(entryExists, types.Position3d{}) {
		return datamodeltypes.ExtendedEntry{}, errors.New("entry does not exist")
	}
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{
		Time:     entryExists.Time,
		Subspace: entryExists.Subspace,
		Path:     entryExists.Path,
	}, e.Opts.PathParams)

	if err != nil {
		return datamodeltypes.ExtendedEntry{}, err
	}
	entryBytes, err := e.Opts.KVDriver.Get(encodedKey)
	if err != nil {
		return datamodeltypes.ExtendedEntry{}, err
	}
	value := kv_driver.DecodeValues(entryBytes)

	return datamodeltypes.ExtendedEntry{
		Entry: types.Entry{
			Timestamp:      entryExists.Time,
			Path:           Path,
			Subspace_id:    Subspace,
			Payload_digest: value.PayloadDigest,
			Payload_length: value.PayloadLength,
			Namespace_id:   e.Storage.Opts.Namespace,
		},
		AuthDigest: value.AuthDigest,
	}, nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Insert(extendedEntry datamodeltypes.ExtendedEntry) error {
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{Time: extendedEntry.Entry.Timestamp, Subspace: extendedEntry.Entry.Subspace_id, Path: extendedEntry.Entry.Path}, e.Opts.PathParams)
	if err != nil {
		return err
	}
	encodedValue := kv_driver.EncodeValues(struct {
		PayloadLength uint64
		PayloadDigest types.PayloadDigest
		AuthDigest    types.PayloadDigest
	}{
		PayloadLength: extendedEntry.Entry.Payload_length,
		PayloadDigest: extendedEntry.Entry.Payload_digest,
		AuthDigest:    extendedEntry.AuthDigest,
	})
	err = e.Opts.KVDriver.Set(encodedKey, encodedValue)
	if err != nil {
		return err
	}
	err = e.Storage.Insert(extendedEntry.Entry.Subspace_id, extendedEntry.Entry.Path, extendedEntry.Entry.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Delete(entry types.Entry) error {
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{Time: entry.Timestamp, Subspace: entry.Subspace_id, Path: entry.Path}, e.Opts.PathParams)
	if err != nil {
		return err
	}

	err = e.Opts.KVDriver.Delete(encodedKey)
	if err != nil {
		return err
	}
	if !(e.Storage.Remove(types.Position3d{Time: entry.Timestamp, Subspace: entry.Subspace_id, Path: entry.Path})) {
		return errors.New("entry does not exist in the KD Tree")
	}
	return nil
}

func (e *EntryDriver[PreFingerPrint, FingerPrint, K]) Query(range3d types.Range3d) ([]datamodeltypes.ExtendedEntry, error) {
	entryNodes := e.Storage.Query(range3d)
	Entries := make([]datamodeltypes.ExtendedEntry, len(entryNodes))
	for _, node := range entryNodes {
		encodedKey, err := kv_driver.EncodeKey(types.Position3d{Time: node.Timestamp, Subspace: node.Subspace, Path: node.Path}, e.Opts.PathParams)
		if err != nil {
			log.Fatalln(err, "can't Encode key")

		}
		encodedValue, err := e.Opts.KVDriver.Get(encodedKey)
		if err != nil {
			return nil, err
		}

		decodedValue := kv_driver.DecodeValues(encodedValue)
		entry := types.Entry{
			Timestamp:      node.Timestamp,
			Path:           node.Path,
			Subspace_id:    node.Subspace,
			Payload_digest: decodedValue.PayloadDigest,
			Payload_length: decodedValue.PayloadLength,
			Namespace_id:   e.Storage.Opts.Namespace,
		}
		Entries = append(Entries, datamodeltypes.ExtendedEntry{
			Entry:      entry,
			AuthDigest: decodedValue.AuthDigest,
		})
	}
	return Entries, nil
}
