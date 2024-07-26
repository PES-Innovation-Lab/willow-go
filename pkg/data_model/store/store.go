package store

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kdnode"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	payloadDriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/payload_kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	kdtree "github.com/rishitc/go-kd-tree"

	"golang.org/x/exp/constraints"
)

type Store[PreFingerPrint, FingerPrint string, K constraints.Unsigned, AuthorisationOpts []byte, AuthorisationToken string] struct {
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
) ([]types.Entry, error) {
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
		return nil, errors.New(err.Error())
	}
	prunedEntries, err := s.IngestEntry(entry, authToken)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	count, err := s.EntryDriver.PayloadReferenceCounter.Count(digest)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	if count == 0 {
		s.PayloadDriver.Erase(digest)
	}
	return prunedEntries, nil
}

func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) IngestEntry(
	entry types.Entry,
	authorisation AuthorisationToken,
) ([]types.Entry, error) {
	s.IngestionMutexLock.Lock() // Locked so that no parallel entry insertions can happen

	// Check if the namespace id of the entry and the current namespace match!
	if !(s.Schemes.NamespaceScheme.IsEqual(s.NameSpaceId, entry.Namespace_id)) {
		s.IngestionMutexLock.Unlock()
		return nil, errors.New("failed to ingest entry\nnamespace does not match store namespace")
	}

	// Check if the authorisation token is valid
	if !(s.Schemes.AuthorisationScheme.IsAuthoriseWrite(entry, authorisation)) {
		s.IngestionMutexLock.Unlock()
		return nil, errors.New("failed to ingest entry\nauthorisation failed")
	}

	// Get all the prefixes of the entry path to be inserted, iterate through them
	// and check if a newer prefix exists, if it does, then this entry is not allowed to be inserted!
	// this is wrt to prefix pruning and this case is not allowed.
	prefixes := s.PrefixDriver.DriverPrefixesOf(entry.Subspace_id, entry.Path, s.Schemes.PathParams, s.EntryDriver.Storage.KDTree)
	for _, prefix := range prefixes {
		if prefix.Timestamp >= entry.Timestamp {
			s.IngestionMutexLock.Unlock()
			return nil, errors.New("failed to ingest entry\nnewer prefix already exists in store")
		}
	}

	// Check if the entry already exists in the store, if it does, check the necesarry conditions which
	// the protocol specifies to decide which of them is newer.
	// If the current inserting entry is found to be older, do not insert, otherwise
	// remove the other entry from all storages
	otherEntry, err := s.EntryDriver.Get(entry.Subspace_id, entry.Path)
	if err != nil && strings.Compare(err.Error(), "entry does not exist") != 0 {
		return nil, errors.New(err.Error())
	}
	if !reflect.DeepEqual(otherEntry.Entry, types.Position3d{}) {
		// Checking if path matches
		if utils.OrderPath(otherEntry.Entry.Path, entry.Path) == 0 {
			if otherEntry.Entry.Timestamp >= entry.Timestamp {
				// Check timestamps for newer entry
				s.IngestionMutexLock.Unlock()
				return nil, errors.New("failed to ingest entry\nnewer entry already exists in store")
			} else if entry.Timestamp == otherEntry.Entry.Timestamp && otherEntry.Entry.Payload_digest >= entry.Payload_digest {
				// Check payload digests for newer entry
				s.IngestionMutexLock.Unlock()
				return nil, errors.New("failed to ingest entry\nnewer entry already exists in store")
			} else if entry.Timestamp == otherEntry.Entry.Timestamp && otherEntry.Entry.Payload_digest == entry.Payload_digest && otherEntry.Entry.Payload_length >= entry.Payload_length {
				// Check payload lengths for newer entry
				s.IngestionMutexLock.Unlock()
				return nil, errors.New("failed to ingest entry\nnewer entry already exists in store")
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

			s.EntryDriver.Delete(otherEntry.Entry)
			s.PayloadDriver.Erase(otherEntry.AuthDigest)

			// Decrement payload ref counter of the other entry, if the count is 0, which means no entry is pointing to it
			// remove the payload itself from the payload driver
			if otherEntry.Entry.Payload_digest != entry.Payload_digest {
				count, err := s.EntryDriver.PayloadReferenceCounter.Decrement(otherEntry.Entry.Payload_digest)
				if err != nil {
					return nil, errors.New(err.Error())
				}
				if count == 0 {
					s.PayloadDriver.Erase(otherEntry.Entry.Payload_digest)
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
		return nil, errors.New(err.Error())
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
	err := s.EntryDriver.Insert(datamodeltypes.ExtendedEntry{
		Entry: types.Entry{
			Timestamp:      entry.Timestamp,
			Path:           entry.Path,
			Payload_digest: entry.PayloadDigest,
			Payload_length: entry.PayloadLength,
			Subspace_id:    entry.Subspace,
			Namespace_id:   s.NameSpaceId,
		},
		AuthDigest: authDigest,
	})
	if err != nil {
		return nil, errors.New(err.Error())
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
		return nil, errors.New(err.Error())
	}
	// Iterate through all the prunable entries and remove them from storage
	for _, entry := range prunableEntries {
		// Remove from storage
		s.EntryDriver.Delete(entry.Entry)

		// Decrement the payload reference counter of the entry

		count, err := s.EntryDriver.PayloadReferenceCounter.Decrement(entry.Entry.Payload_digest)
		if err != nil {
			return nil, errors.New(err.Error())
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
	kdt *kdtree.KDTree[kdnode.Key],
	entry types.Position3d,
	pathParams types.PathParams[K],
) ([]datamodeltypes.ExtendedEntry, error,
) {
	// converting the 3D position to a 3D RANGE OMG SO COOL. this is done inside prefixedby func
	// prefixedby func basically does all the work, this is just a wrapper

	prunableEntries := s.PrefixDriver.PrefixedBy(entry.Subspace, entry.Path, pathParams, kdt)

	final_prunables := make([]datamodeltypes.ExtendedEntry, 0, len(prunableEntries))
	for _, prune_candidate := range prunableEntries {
		if prune_candidate.Timestamp < entry.Time {
			retEntry, err := s.EntryDriver.Get(prune_candidate.Subspace, prune_candidate.Path)

			if err != nil {
				return nil, err
			}

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
		Path:     entryDetails.Path,
	}, s.Schemes.PathParams)
	if err != nil {
		return 0, errors.New(err.Error())
	}
	getEntry, err := s.EntryDriver.Opts.KVDriver.Get(encodedKey)
	if err != nil {
		return 0, errors.New(err.Error())
	}
	if getEntry != nil {
		return Failure, errors.New("entry does not exist")
	}

	// Samar needs to make a DecodeValue Function that returns payloadDigest
	// Len and Digest as mentioned by the entry
	decodedValue := kv_driver.DecodeValues(getEntry)

	existingPayload, err := s.PayloadDriver.Get(decodedValue.PayloadDigest)
	if err != nil {
		return 0, errors.New("unable to Get")
	}

	if !reflect.DeepEqual(existingPayload, datamodeltypes.Payload{}) {
		return No_Op, errors.New("file already exists")
	}
	// Result after fully ingesting the paylaod
	resDigest, resLen, resCommit, resReject, err := s.PayloadDriver.Receive(payload, offset, decodedValue.PayloadLength, decodedValue.PayloadDigest)
	if err != nil {
		return 0, errors.New("unable to receive")
	}
	if resLen > decodedValue.PayloadLength || (!allowPartial && decodedValue.PayloadLength != resLen) || (resLen == decodedValue.PayloadLength && s.Schemes.PayloadScheme.Order(resDigest, decodedValue.PayloadDigest) != 0) {
		resReject()
		return Failure, errors.New("data mismatch")
	}

	resCommit(resLen == decodedValue.PayloadLength)

	if resLen == decodedValue.PayloadLength && (s.Schemes.PayloadScheme.Order(resDigest, decodedValue.PayloadDigest) == 0) {
		complete, err := s.PayloadDriver.Get(decodedValue.PayloadDigest)
		if err != nil {
			return 0, errors.New(err.Error())
		}

		if reflect.DeepEqual(complete, datamodeltypes.Payload{}) {
			return 0, fmt.Errorf("could not get payload for a payload that was just ingested")
		}
		authToken, err := s.PayloadDriver.Get(decodedValue.AuthDigest)
		if err != nil {
			return 0, fmt.Errorf("could not get payload for a payload that was just ingested: %s", err.Error())
		}
		if reflect.DeepEqual(authToken, datamodeltypes.Payload{}) {
			return 0, fmt.Errorf("could not get payload for a payload that was just ingested")
		}

	}
	// s.Storage.UpdateAvailablePayload(entryDetails.Subspace, entryDetails.Path)
	return Success, nil
}

// Returns range of the passed areaOfInterest
func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) AreaOfInterestToRange(
	areaOfInterest types.AreaOfInterest,
) (types.Range3d, error) {
	return s.EntryDriver.Storage.GetInterestRange(areaOfInterest), nil
}

// Function which returns the payload if we pass in the entry details
// it takes subaspace path and time, gets the payloadDigest from KVStore and returns payload from filesystem.
func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) GetPayload(position types.Position3d) ([]byte, error) {
	encodedkey, err := kv_driver.EncodeKey(position, s.Schemes.PathParams)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	encodedValue, err := s.EntryDriver.Opts.KVDriver.Get(encodedkey)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	Entry := kv_driver.DecodeValues(encodedValue)

	payload, err := s.PayloadDriver.Get(Entry.PayloadDigest)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return (payload.Bytes()), nil
}

// function to return all values present in the tree
func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) List() []kdnode.Key {
	return s.EntryDriver.Storage.KDTree.Values()
}

// function to return values present in the area of interest
func (s *Store[PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]) ListWithAOI(aoi types.AreaOfInterest) ([]types.Entry, error) {
	entries, err := s.EntryDriver.Opts.KVDriver.ListValues(aoi, s.Schemes.PathParams, s.NameSpaceId)

	if err != nil {
		return nil, err
	}
	return entries, nil
}
