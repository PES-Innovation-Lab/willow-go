package store

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)
type Store[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint constraints.Ordered, K constraints.Unsigned, AuthorisationOpts any, AuthorisationToken string, T datamodeltypes.KvPart] struct {
    Schemes datamodeltypes.StoreSchemes[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K , AuthorisationOpts, AuthorisationToken]
	EntryDriver datamodeltypes.EntryDriver[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	PayloadDriver datamodeltypes.PayloadDriver[PayloadDigest, K]
	Storage datamodeltypes.KDTreeStorage[NamespaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, T, K]
	NameSpaceId NamespaceId
	IngestionMutexLock sync.Mutex
	PrunableEntries func (Subspace SubspaceId, Path types.Path, Timestamp uint64) ([]struct{entry types.Entry[NamespaceId, SubspaceId, PayloadDigest];authTokenHash PayloadDigest}, error)

}

func (s *Store[NameSpaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) Set(input datamodeltypes.EntryInput[SubspaceId], authorisation AuthorisationOpts)  {
	timestamp := input.Timestamp
	if timestamp == 0 {
		timestamp = uint64(time.Now().UnixMicro())
	}
	digest, payload, length := s.PayloadDriver.Set(input.Payload)

	entry := types.Entry[NameSpaceId, SubspaceId, PayloadDigest]{
		Subspace_id: input.Subspace,
		Payload_digest: digest,
		Path: input.Path,
		Payload_length: length,
		Timestamp: timestamp,
		Namespace_id: s.NameSpaceId,
	}
	authToken := s.Schemes.AuthorisationScheme.Authorise(entry, authorisation)

	ingestResult := s.IngestEntry()

}

func (s *Store[NameSpaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) IngestEntry(
	entry types.Entry[NameSpaceId, SubspaceId, PayloadDigest],
	authorisation AuthorisationToken,
	) error {
	s.IngestionMutexLock.Lock()
	if s.Schemes.NamespaceScheme.IsEqual(s.NameSpaceId, entry.Namespace_id){
		s.IngestionMutexLock.Unlock()
		return errors.New("failed to ingest entry\nnamespace does not match store namespace")
	}
	if !(<-s.Schemes.AuthorisationScheme.IsAuthoriseWrite(entry, authorisation)){
		s.IngestionMutexLock.Unlock()
		return errors.New("failed to ingest entry\nauthorisation failed")
	}
	//TO-DO checking for collisions, entry already exists or newer prefix foiund
	//Query for path, check if subspace is same, if subspace is same, check timestamp
	//Timestamp of this entry should be greater than the timestamp of the entry in the store
	//If it is not return error!

	//TO-DO: check if the entry is already present in the store
	//If it is present, check if the timestamp is greater than the timestamp of the entry in the store
	//If it is not return error!
	//If the timstamp is lesser, remove that entry and add the newer one..

	//TO-DO call insert entry, check if the insertion was called from local system
	//Get a list of pruned entries!

	//Finally add the entry, release the lock and return a success, with pruned entries and the entry itself!!!!!
	return nil
}

func (s *Store[NameSpaceId, SubspaceId, PayloadDigest, PreFingerPrint, FingerPrint, K, AuthorisationOpts, AuthorisationToken, T]) InsertEntry(
	entry struct{
		Path types.Path
		Subspace SubspaceId
		Timestamp uint64
		Hash PayloadDigest
		Length uint64
		AuthToken AuthorisationToken
	}) ([]types.Entry[NameSpaceId, SubspaceId, PayloadDigest] ,error) {

	encodedToken := s.Schemes.AuthorisationScheme.TokenEncoding.Encode(entry.AuthToken)
	tokenDigest, _, _ := s.PayloadDriver.Set(encodedToken)

	var prunedEntries []types.Entry[NameSpaceId, SubspaceId, PayloadDigest]

	err := s.Storage.Insert(struct{Subspace SubspaceId; Path types.Path; PayloadDigest PayloadDigest; Timestamp uint64; PayloadLength uint64; AuthTokenHash PayloadDigest}{
		Subspace: entry.Subspace,
		Path: entry.Path,
		PayloadDigest: entry.Hash,
		Timestamp: entry.Timestamp,
		PayloadLength: entry.Length,
		AuthTokenHash: tokenDigest,
	})
	if err != nil {
		log.Fatal(err)
	}

	s.EntryDriver.PayloadReferenceCounter.Increment(entry.Hash)

	prunableEntries, err := s.PrunableEntries(entry.Subspace, entry.Path, entry.Timestamp)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range prunableEntries {
		err := s.Storage.Remove(entry.entry)
		if err != nil {
			log.Fatal(err)
		}
		
		count := s.EntryDriver.PayloadReferenceCounter.Decrement(entry.authTokenHash)
		if count == 0 {
			s.PayloadDriver.Erase(entry.entry.Payload_digest)
		}

		prunedEntries = append(prunedEntries, entry.entry)
	}
	return prunedEntries, nil
}