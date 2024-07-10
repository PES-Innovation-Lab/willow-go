package payloadDriver

import (
	"encoding/base32"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"golang.org/x/exp/constraints"
)

type PayloadDriver[PayloadDigest constraints.Ordered, T constraints.Unsigned] struct {
	path          string
	PayloadScheme datamodeltypes.PayloadScheme[PayloadDigest, T]
}

func (pd *PayloadDriver[PayloadDigest, T]) getKey(hash PayloadDigest) string {
	encoded := pd.PayloadScheme.Encode(hash)
	return base32.StdEncoding.EncodeToString(encoded)
}

func (pd *PayloadDriver[PayloadDigest, T]) getPayload(filepath string) datamodeltypes.Payload {
	return datamodeltypes.Payload{
		Bytes: func() []byte {
			bytes, _ := os.ReadFile(filepath)
			return bytes
		},
		BytesWithOffset: func(offset int) ([]byte, error) {
			file, err := os.Open(filepath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			fileInfo, _ := file.Stat()
			size := fileInfo.Size()
			if offset >= int(size) {
				return nil, fmt.Errorf("Offset is greater than file size")
			}

			bytes := make([]byte, size-int64(offset))
			_, err = file.ReadAt(bytes, int64(offset))
			if err != nil {
				return nil, err
			}
			return bytes, nil
		},
		Length: func() (uint64, error) {
			file, err := os.Open(filepath)
			if err != nil {
				return 0, err
			}
			defer file.Close()

			fileInfo, _ := file.Stat()
			size := fileInfo.Size()
			return uint64(size), nil
		},
	}

}

func (pd *PayloadDriver[PayloadDigest, T]) get(PayloadHash PayloadDigest) (datamodeltypes.Payload, error) {
	filepath := filepath.Join(pd.path, pd.getKey(PayloadHash))
	_, err := os.Lstat(filepath)
	if err != nil {
		return datamodeltypes.Payload{}, err
	}

	return pd.getPayload(filepath), nil
}

func (pd *PayloadDriver[PayloadDigest, T]) erase(PayloadHash PayloadDigest) (bool, error) {
	filepath := filepath.Join(pd.path, pd.getKey(PayloadHash))
	err := os.Remove(filepath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pd *PayloadDriver[PayloadDigest, T]) set(payload []byte) (PayloadDigest, datamodeltypes.Payload, uint64) {
	digest := <-pd.PayloadScheme.FromBytes(payload)
	pd.ensureDir()
	filepath := filepath.Join(pd.path, pd.getKey(digest))
	os.WriteFile(filepath, payload, 0755)
	var retPayload datamodeltypes.Payload = pd.getPayload(filepath)
	return digest, retPayload, uint64(len(payload))
}

func (pd *PayloadDriver[PayloadDigest, T]) ensureDir(args ...string) (string, error) {
	path := filepath.Join(append([]string{pd.path}, args...)...)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	return path, nil
}
