package store

import (
	"log"
	"sync"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/Kdtree"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	entrydriver "github.com/PES-Innovation-Lab/willow-go/pkg/data_model/entry_driver"
	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/kv_driver"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"github.com/PES-Innovation-Lab/willow-go/utils"
	"golang.org/x/exp/constraints"
)

type Store[PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts any, AuthorisationToken string, T datamodeltypes.KvPart] struct {
	Schemes            datamodeltypes.StoreSchemes[PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken]
	EntryDriver        entrydriver.EntryDriver[PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	PayloadDriver      datamodeltypes.PayloadDriver[PayloadDigest, K]
	Storage            datamodeltypes.KDTreeStorage[PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	NameSpaceId        types.NamespaceId
	IngestionMutexLock sync.Mutex
}

func (s *Store[PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) Set(
	input datamodeltypes.EntryInput,
	authorisation AuthorisationOpts,
) []types.Entry[PayloadDigest] {
	timestamp := input.Timestamp
	if timestamp == 0 {
		timestamp = uint64(time.Now().UnixMicro())
	}
	digest, _, length := s.PayloadDriver.Set(input.Payload)

	entry := types.Entry[PayloadDigest]{
		Subspace_id:    input.Subspace,
		Payload_digest: digest,
		Path:           input.Path,
		Payload_length: length,
		Timestamp:      timestamp,
		Namespace_id:   s.NameSpaceId,
	}
	authToken := s.Schemes.AuthorisationScheme.Authorise(entry, authorisation)

	prunedEntries, err := s.IngestEntry(entry, authToken)
	if err != nil {
		log.Fatal(err)
	}
	count := s.EntryDriver.PayloadReferenceCounter.Count(digest)
	if count == 0 {
		s.PayloadDriver.Erase(digest)
	}
	return prunedEntries
}

func (s *Store[PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) IngestEntry(
	entry types.Entry[PayloadDigest],
	authorisation AuthorisationToken,
) ([]types.Entry[PayloadDigest], error) {
	s.IngestionMutexLock.Lock() // Locked so that no parallel entry insertions can happen

	// Check if the namespace id of the entry and the current namespace match!
	if s.Schemes.NamespaceScheme.IsEqual(s.NameSpaceId, entry.Namespace_id) {
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
	for _, prefix := range kv_driver.DriverPrefixesOf(entry.Path, s.Schemes.PathParams, s.Storage.KDTree) {
		if prefix.Timestamp >= entry.Timestamp {
			s.IngestionMutexLock.Unlock()
			log.Fatal("failed to ingest entry\nnewer prefix already exists in store")
		}
	}

	// Check if the entry already exists in the store, if it does, check the necesarry conditions which
	// the protocol specifies to decide which of them is newer.
	// If the current inserting entry is found to be older, do not insert, otherwise
	// remove the other entry from all storages
	otherEntry, err := s.Storage.Get(entry.Subspace_id, entry.Path)

	// To better handle other errors, the error returned by the function should specify what kind of error it is
	if err != nil && err.Error() != "entry found" {
		// Checking if path matches
		if utils.OrderPath(otherEntry.Entry.Path, entry.Path) == 0 {
			if otherEntry.Entry.Timestamp >= entry.Timestamp {
				// Check timestamps for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer entry already exists in store")
			} else if otherEntry.Entry.Payload_digest >= entry.Payload_digest {
				// Check payload digests for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer prefix already exists in store")
			} else if otherEntry.Entry.Payload_length == entry.Payload_length {
				// Check payload lengths for newer entry
				s.IngestionMutexLock.Unlock()
				log.Fatal("failed to ingest entry\nnewer prefix already exists in store")
			}
			// If the three conditions does not satisgy, it means the entry to be inserted is newer
			// and the other entry should be removed
			// Remove the other entry from all storages
			s.Storage.Remove(types.Position3d{
				Subspace: otherEntry.Entry.Subspace_id,
				Path:     otherEntry.Entry.Path,
				Time:     otherEntry.Entry.Timestamp,
			})

			// Decrement payload ref counter of the other entry, if the count is 0, which means no entry is pointing to it
			// remove the payload itself from the payload driver
			count := s.EntryDriver.PayloadReferenceCounter.Decrement(otherEntry.AuthTokenHash)
			if count == 0 {
				s.PayloadDriver.Erase(otherEntry.Entry.Payload_digest)
			}
			// TO-DO: remove the entry from entry KV
		}
	} else {
		// If the error returned from get is some other error, then print it and exit
		log.Fatal(err)
	}

	// Insert the entry into the storage
	// This function also returns the entries which are pruned due to the insertion of the current entry
	prunedEntries, err := s.InsertEntry(struct {
		Path      types.Path
		Subspace  types.SubspaceId
		Timestamp uint64
		Hash      PayloadDigest
		Length    uint64
		AuthToken AuthorisationToken
	}{
		Path:      entry.Path,
		Subspace:  entry.Subspace_id,
		Timestamp: entry.Timestamp,
		Hash:      entry.Payload_digest,
		Length:    entry.Payload_length,
		AuthToken: authorisation,
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

func (s *Store[PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) InsertEntry(
	entry struct {
		Path      types.Path
		Subspace  types.SubspaceId
		Timestamp uint64
		Hash      PayloadDigest
		Length    uint64
		AuthToken AuthorisationToken
	},
) ([]types.Entry[PayloadDigest], error) {
	// Encode the authorisation token and get the digest of the token
	encodedToken := s.Schemes.AuthorisationScheme.TokenEncoding.Encode(entry.AuthToken)
	tokenDigest, _, _ := s.PayloadDriver.Set(encodedToken)

	// Insert the entry into the storage
	err := s.Storage.Insert(struct {
		Subspace      types.SubspaceId
		Path          types.Path
		PayloadDigest PayloadDigest
		Timestamp     uint64
		PayloadLength uint64
		AuthTokenHash PayloadDigest
	}{
		Subspace:      entry.Subspace,
		Path:          entry.Path,
		PayloadDigest: entry.Hash,
		Timestamp:     entry.Timestamp,
		PayloadLength: entry.Length,
		AuthTokenHash: tokenDigest,
	})
	// If there is an error in inserting the entry, print it and exit
	if err != nil {
		log.Fatal(err)
	}
	// Increment the payload reference counter of the entry
	s.EntryDriver.PayloadReferenceCounter.Increment(entry.Hash)

	// Variable to store pruned entries
	var prunedEntries []types.Entry[PayloadDigest]

	// Get a list of all the prunable entries so that they can be pruned
	prunableEntries, err := s.PrunableEntries(s.Storage.KDTree, types.Position3d{
		Subspace: entry.Subspace,
		Path:     entry.Path,
		Time:     entry.Timestamp,
	}, s.Schemes.PathParams)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through all the prunable entries and remove them from storage
	for _, entry := range prunableEntries {
		// Remove from storage
		err := s.Storage.Remove(types.Position3d{
			Subspace: entry.entry.Subspace_id,
			Path:    entry.entry.Path,
			Time:   entry.entry.Timestamp,
		})
		if err != nil {
			log.Fatal(err)
		}

		// Decrement the payload reference counter of the entry
		count := s.EntryDriver.PayloadReferenceCounter.Decrement(entry.authTokenHash)
		// If the count is 0, which means no entry is pointing to it, remove the payload itself from the payload driver
		if count == 0 {
			s.PayloadDriver.Erase(entry.entry.Payload_digest)
		}
		// Append the pruned entry to prunedEntries array
		prunedEntries = append(prunedEntries, entry.entry)
	}

	// Return the pruned entries with no errors
	return prunedEntries, nil
}

func (s *Store[PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) PrunableEntries(
	kdt *(Kdtree.KDTree[Kdtree.KDNodeKey]),
	entry types.Position3d,
	pathParams types.PathParams[K],
) ([]struct {
	entry         types.Entry[PayloadDigest]
	authTokenHash PayloadDigest
}, error,
) {
	// converting the 3D position to a 3D RANGE OMG SO COOL. this is done inside prefixedby func
	// prefizedby func basically does all the work, this is just a wrapper

	prunableEntries := kv_driver.PrefixedBy(entry.Subspace, entry.Path, pathParams, kdt)
	final_prunables := make([]struct{
		entry types.Entry[PayloadDigest]
		authTokenHash PayloadDigest
		}, 0, len(prunableEntries))
	for _, prune_candidate := range prunableEntries {
		if prune_candidate.Timestamp < entry.Time {
			encodedEntry, err := kv_driver.EncodeKey(prune_candidate.Timestamp, prune_candidate.Subspace, pathParams, prune_candidate.Path)
			encodedValue, err := s.EntryDriver.Opts.KVDriver.Get(encodedEntry)
			if err != nil {
				return nil, err
			}
			//get decoded value
			// decodedValue := 
			var decodedValue struct{
				PayloadDigest PayloadDigest
				AuthTokenHash PayloadDigest
				PayloadLentgh uint64
			}
			if err != nil {
				return nil, err
			}
			// Get the authorisation token hash of the entry
			final_prunables = append(final_prunables, struct{
				entry types.Entry[PayloadDigest]
				authTokenHash PayloadDigest}{
					entry: types.Entry[PayloadDigest]{
						Subspace_id:    prune_candidate.Subspace,
						Payload_digest: decodedValue.PayloadDigest,
						Timestamp: 	prune_candidate.Timestamp,
						Path: 			prune_candidate.Path,
						Payload_length: decodedValue.PayloadLentgh,
						Namespace_id: 	s.NameSpaceId,
					},
					authTokenHash: decodedValue.AuthTokenHash,
				})
		}
	}

	return final_prunables, nil
}
