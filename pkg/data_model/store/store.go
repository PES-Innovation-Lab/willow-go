package store

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type Store[PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts []byte, AuthorisationToken string] struct {
	Schemes            datamodeltypes.StoreSchemes[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
	EntryDriver        entrydriver.EntryDriver[PreFingerPrint, FingerPrint, K]
	PayloadDriver      payloadDriver.PayloadDriver
	NameSpaceId        types.NamespaceId
	IngestionMutexLock sync.Mutex
	PrefixDriver       kv_driver.PrefixDriver[K]
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) Set(
	input datamodeltypes.EntryInput,
	authorisation AuthorisationOpts,
) []types.Entry {
	timestamp := input.Timestamp
	if timestamp == 0 {
		timestamp = uint64(time.Now().UnixMicro())
	}
	digest, _, length := s.PayloadDriver.Set(input.Payload)

	entry := types.Entry{
		Subspace_id:    input.Subspace,
		Payload_digest: digest,
		Path:           input.Path,
		Payload_length: length,
		Timestamp:      timestamp,
		Namespace_id:   s.NameSpaceId,
	}
	authToken, err := s.Schemes.AuthorisationScheme.Authorise(entry, authorisation)
	if err != nil {
		log.Fatal(err)
	}
	prunedEntries, err := s.IngestEntry(entry, authToken)
	if err != nil {
		log.Fatal(err)
	}
	count, err := s.EntryDriver.PayloadReferenceCounter.Count(digest)
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		s.PayloadDriver.Erase(digest)
	}
	return prunedEntries
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) IngestEntry(
	entry types.Entry,
	authorisation AuthorisationToken,
) ([]types.Entry, error) {
	s.IngestionMutexLock.Lock() // Locked so that no parallel entry insertions can happen

	// Check if the namespace id of the entry and the current namespace match!
	if !(s.Schemes.NamespaceScheme.IsEqual(s.NameSpaceId, entry.Namespace_id)) {
		s.IngestionMutexLock.Unlock()
		log.Fatal("failed to ingest entry\nnamespace does not match store namespace")
	}

	// Check if the authorisation token is valid
	if !(s.Schemes.AuthorisationScheme.IsAuthoriseWrite(entry, authorisation)) {
		s.IngestionMutexLock.Unlock()
		log.Fatal("failed to ingest entry\nauthorisation failed")
	}

	// Get all the prefixes of the entry path to be inserted, iterate through them
	// and check if a newer prefix exists, if it does, then this entry is not allowed to be inserted!
	// this is wrt to prefix pruning and this case is not allowed.
	prefixes := s.PrefixDriver.DriverPrefixesOf(entry.Subspace_id, entry.Path, s.Schemes.PathParams, s.EntryDriver.Storage.KDTree)
	for _, prefix := range prefixes {
		if prefix.Timestamp >= entry.Timestamp {
			s.IngestionMutexLock.Unlock()
			log.Fatal("failed to ingest entry\nnewer prefix already exists in store")
		}
	}

	// Check if the entry already exists in the store, if it does, check the necesarry conditions which
	// the protocol specifies to decide which of them is newer.
	// If the current inserting entry is found to be older, do not insert, otherwise
	// remove the other entry from all storages
	retEntry, err := s.EntryDriver.Get(entry.Subspace_id, entry.Path)
	otherEntry := retEntry.Entry
	authDigest := retEntry.AuthTokenHash
	if err != nil && strings.Compare(err.Error(), "entry does not exist") != 0 {
		log.Fatal(err)
	}
	if !reflect.DeepEqual(otherEntry, types.Position3d{}) {
		// Checking if path matches
		if utils.OrderPath(otherEntry.Path, entry.Path) == 0 {
			if otherEntry.Timestamp >= entry.Timestamp {
				// Check timestamps for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer entry already exists in store")
			} else if entry.Timestamp == otherEntry.Timestamp && otherEntry.Payload_digest >= entry.Payload_digest {
				// Check payload digests for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer entry already exists in store")
			} else if entry.Timestamp == otherEntry.Timestamp && otherEntry.Payload_digest == entry.Payload_digest && otherEntry.Payload_length == entry.Payload_length {
				// Check payload lengths for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer entry already exists in store")
			}
			// If the three conditions does not satisgy, it means the entry to be inserted is newer
			// and the other entry should be removed
			// Remove the other entry from all storages
			// s.Storage.Remove(types.Position3d{
			// 	Subspace: otherEntry.Subspace_id,
			// 	Path:     otherEntry.Path,
			// 	Time:     otherEntry.Timestamp,
			// })

			// s.EntryDriver.Opts.KVDriver.Delete(encodedKey)

			s.EntryDriver.Delete(otherEntry)
			s.PayloadDriver.Erase(authDigest)

			// Decrement payload ref counter of the other entry, if the count is 0, which means no entry is pointing to it
			// remove the payload itself from the payload driver
			if  otherEntry.Payload_digest != entry.Payload_digest {
				count, err := s.EntryDriver.PayloadReferenceCounter.Decrement(otherEntry.Payload_digest)
				if err != nil {
					log.Fatal(err)
				}
				if count == 0 {
					s.PayloadDriver.Erase(otherEntry.Payload_digest)
				}
			}
		}
	}

	// Insert the entry into the storage
	// This function also returns the entries which are pruned due to the insertion of the current entry
	prunedEntries, err := s.InsertEntry(struct {
		Path          types.Path
		Subspace      types.SubspaceId
		Timestamp     uint64
		PayloadDigest types.PayloadDigest
		PayloadLength uint64
		AuthToken     AuthorisationToken
	}{
		Path:          entry.Path,
		Subspace:      entry.Subspace_id,
		Timestamp:     entry.Timestamp,
		PayloadDigest: entry.Payload_digest,
		PayloadLength: entry.Payload_length,
		AuthToken:     authorisation,
	})
	// If there is an error in inserting the entry, print it and exit
	if err != nil {
		log.Fatal(err)
	}

	// Unlock the mutex lock
	s.IngestionMutexLock.Unlock()

	// Return the pruned entries and the entry which was inserted with no errors
	return prunedEntries, nil
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) InsertEntry(
	entry struct {
		Path          types.Path
		Subspace      types.SubspaceId
		Timestamp     uint64
		PayloadDigest types.PayloadDigest
		PayloadLength uint64
		AuthToken     AuthorisationToken
	},
) ([]types.Entry, error) {
	// Encode the authorisation token and get the digest of the token
	encodedToken := s.Schemes.AuthorisationScheme.TokenEncoding.Encode(entry.AuthToken)
	authDigest, _, _ := s.PayloadDriver.Set(encodedToken)
	
	// Insert the entry into the storage
	err := s.EntryDriver.Insert(types.Entry{
		Subspace_id:    entry.Subspace,
		Payload_digest: entry.PayloadDigest,
		Payload_length: entry.PayloadLength,
		Path: 		 entry.Path,
		Timestamp:      entry.Timestamp,
		Namespace_id:  s.NameSpaceId,
		}, authDigest)
		if err != nil {
			log.Fatal(err)
		}
		
	// Increment the payload reference counter of the entry
	s.EntryDriver.PayloadReferenceCounter.Increment(entry.PayloadDigest)

	// Variable to store pruned entries
	var prunedEntries []types.Entry

	// Get a list of all the prunable entries so that they can be pruned
	prunableEntries, err := s.PrunableEntries(s.EntryDriver.Storage.KDTree, types.Position3d{
		Subspace: entry.Subspace,
		Path:     entry.Path,
		Time:     entry.Timestamp,
	}, s.Schemes.PathParams)
	if err != nil {
		log.Fatalln(err)
	}

	// Iterate through all the prunable entries and remove them from storage
	for _, entry := range prunableEntries {
		// Remove from storage
		s.EntryDriver.Delete(entry.Entry)

		// Decrement the payload reference counter of the entry

		count, err := s.EntryDriver.PayloadReferenceCounter.Decrement(entry.Entry.Payload_digest)
		if err != nil {
			log.Fatal(err)
		}
		// If the count is 0, which means no entry is pointing to it, remove the payload itself from the payload driver
		if count == 0 {
			s.PayloadDriver.Erase(entry.Entry.Payload_digest)
		}
		// Append the pruned entry to prunedEntries array
		prunedEntries = append(prunedEntries, entry.Entry)
	}

	// Return the pruned entries with no errors
	return prunedEntries, nil
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) PrunableEntries(
	kdt *Kdtree.KDTree[Kdtree.KDNodeKey],
	entry types.Position3d,
	pathParams types.PathParams[K],
) ([]struct {
	Entry         types.Entry
	AuthTokenHash types.PayloadDigest
}, error,
) {
	// converting the 3D position to a 3D RANGE OMG SO COOL. this is done inside prefixedby func
	// prefixedby func basically does all the work, this is just a wrapper

	prunableEntries := s.PrefixDriver.PrefixedBy(entry.Subspace, entry.Path, pathParams, kdt)

	final_prunables := make([]struct {
		Entry         types.Entry
		AuthTokenHash types.PayloadDigest
	}, 0, len(prunableEntries))
	for _, prune_candidate := range prunableEntries {
		if prune_candidate.Timestamp < entry.Time {
			// encodedEntry, _ := kv_driver.EncodeKey(types.Position3d{
			// 	Time:     prune_candidate.Timestamp,
			// 	Subspace: prune_candidate.Subspace,
			// 	Path:   prune_candidate.Path,
			// }, pathParams)
			// encodedValue, err := s.EntryDriver.Opts.KVDriver.Get(encodedEntry)
			retEntry ,err := s.EntryDriver.Get(prune_candidate.Subspace, prune_candidate.Path)
			
			if err != nil {
				return nil, err
			}
			// decodedValue := kv_driver.DecodeValues(encodedValue)

			// Get the authorisation token hash of the entry
			final_prunables = append(final_prunables, retEntry)
		}
	}
	return final_prunables, nil
}

type Status int

// Setting up an enum for the return values
const (
	Failure Status = -1
	No_Op   Status = 0
	Success Status = 1
)

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) IngestPayload(
	entryDetails types.Position3d,
	payload []byte,
	allowPartial bool,
	offset int64,
) (Status, error) {
	encodedKey, err := kv_driver.EncodeKey(types.Position3d{
		Time:     entryDetails.Time,
		Subspace: entryDetails.Subspace,
		Path:   entryDetails.Path,
	}, s.Schemes.PathParams)
	if err != nil {
		log.Fatal(err)
	}
	getEntry, err := s.EntryDriver.Opts.KVDriver.Get(encodedKey)
	if err != nil {
		log.Fatal(err)
	}
	if getEntry != nil {
		return Failure, fmt.Errorf("entry does not exist")
	}

	// Samar needs to make a DecodeValue Function that returns payloadDigest
	// Len and Digest as mentioned by the entry
	decodedValue := kv_driver.DecodeValues(getEntry)

	existingPayload, err := s.PayloadDriver.Get(decodedValue.PayloadDigest)
	if err != nil {
		log.Fatal("Unable to Get")
	}

	if !reflect.DeepEqual(existingPayload, datamodeltypes.Payload{}) {
		return No_Op, fmt.Errorf("file already exists")
	}
	// Result after fully ingesting the paylaod
	resDigest, resLen, resCommit, resReject, err := s.PayloadDriver.Receive(payload, offset, decodedValue.PayloadLength, decodedValue.PayloadDigest)
	if err != nil {
		log.Fatal("Unable to receive")
	}
	if resLen > decodedValue.PayloadLength || (!allowPartial && decodedValue.PayloadLength != resLen) || (resLen == decodedValue.PayloadLength && s.Schemes.PayloadScheme.Order(resDigest, decodedValue.PayloadDigest) != 0) {
		resReject()
		return Failure, fmt.Errorf("data mismatch")
	}

	resCommit(resLen == decodedValue.PayloadLength)

	if resLen == decodedValue.PayloadLength && (s.Schemes.PayloadScheme.Order(resDigest, decodedValue.PayloadDigest) == 0) {
		complete, err := s.PayloadDriver.Get(decodedValue.PayloadDigest)
		if err != nil {
			log.Fatal(err)
		}

		if reflect.DeepEqual(complete, datamodeltypes.Payload{}) {
			log.Fatalln("Could not get payload for a payload that was just ingested", err)
		}
		authToken, err := s.PayloadDriver.Get(decodedValue.AuthDigest)
		if err != nil {
			log.Fatal(err)
		}
		if reflect.DeepEqual(authToken, datamodeltypes.Payload{}) {
			log.Fatalln("Could not get authorisation token for a stored entry.", err)
		}

	}
	// s.Storage.UpdateAvailablePayload(entryDetails.Subspace, entryDetails.Path)
	return Success, nil
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) AreaOfInterestToRange(
	areaOfInterest types.AreaOfInterest,
) types.Range3d {
	return s.Storage.GetInterestRange(areaOfInterest)
}
