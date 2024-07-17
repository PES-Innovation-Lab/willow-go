package payloadDriver

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/types"
	"golang.org/x/exp/constraints"
)

type PayloadDriver[T constraints.Unsigned] struct {
	path          string
	PayloadScheme datamodeltypes.PayloadScheme[T]
}

func (pd *PayloadDriver[T]) GetKey(hash types.PayloadDigest) string {
	encoded := pd.PayloadScheme.Encode(hash)
	return base32.StdEncoding.EncodeToString(encoded)
}

func (pd *PayloadDriver[T]) GetPayload(filepath string) datamodeltypes.Payload {
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
				return nil, fmt.Errorf("offset is greater than file size")
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

func (pd *PayloadDriver[T]) Get(PayloadHash types.PayloadDigest) (datamodeltypes.Payload, error) {
	filepath := filepath.Join(pd.path, pd.GetKey(PayloadHash))
	_, err := os.Lstat(filepath)
	if err != nil {
		return datamodeltypes.Payload{}, err
	}

	return pd.GetPayload(filepath), nil
}

func (pd *PayloadDriver[T]) Erase(PayloadHash types.PayloadDigest) (bool, error) {
	filepath := filepath.Join(pd.path, pd.GetKey(PayloadHash))
	err := os.Remove(filepath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pd *PayloadDriver[T]) Set(payload []byte) (types.PayloadDigest, datamodeltypes.Payload, uint64) {
	digest := <-pd.PayloadScheme.FromBytes(payload)
	pd.EnsureDir()
	filepath := filepath.Join(pd.path, pd.GetKey(digest))
	os.WriteFile(filepath, payload, 0777)
	var retPayload datamodeltypes.Payload = pd.GetPayload(filepath)
	return digest, retPayload, uint64(len(payload))
}

func (pd *PayloadDriver[T]) EnsureDir(args ...string) (string, error) {
	path := filepath.Join(append([]string{pd.path}, args...)...)
	err := os.MkdirAll(path, 0777)
	fmt.Println(err, path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func getRandomBytes() ([]byte, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		fmt.Println("asdasd", err)
		return err
	}
	defer src.Close()

	dst, err := os.Create(to) // Use os.Create to truncate the destination file if it exists
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	fmt.Println(err)
	return err
}

func (pd *PayloadDriver[T]) Receive(payload []byte, offset int64, expectedLength uint64, expectedDigest types.PayloadDigest) (types.PayloadDigest, uint64, datamodeltypes.CommitType, datamodeltypes.RejectType, error) {

	_, err := pd.EnsureDir("staging")
	if err != nil {
		panic("Unable to locate the staging file: " + err.Error())
	}

	// Generate a temporary file name
	randBytes, err := getRandomBytes()
	if err != nil {
		panic("Unable to generate random bytes: " + err.Error())
	}
	tempKey := base32.StdEncoding.EncodeToString(randBytes)

	stagingFilePath := filepath.Join(pd.path, "staging", tempKey)

	// If offset is greater than 0, copy the existing partial file to staging
	if offset > 0 {
		partialFilePath := filepath.Join(pd.path, "partial", pd.GetKey(expectedDigest))
		err := copyFile(partialFilePath, stagingFilePath)
		if err != nil {
			panic("Unable to copy file: " + err.Error())
		}
	}

	// Open the file in the appropriate mode
	var file *os.File
	if offset == 0 {
		file, err = os.OpenFile(stagingFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	} else {
		file, err = os.OpenFile(stagingFilePath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0777)
	}
	if err != nil {
		panic("Unable to open file: " + err.Error())
	}
	defer file.Close()

	// If offset is greater than 0, truncate and seek
	if offset > 0 {
		err = file.Truncate(offset)
		if err != nil {
			panic("Unable to truncate file: " + err.Error())

		}
		_, err = file.Seek(offset, 0)
		if err != nil {
			panic("Unable to seek file: " + err.Error())
		}
	}

	// Write the payload to the file
	writer := io.Writer(file)
	receivedLen := offset + int64(len(payload))
	if _, err := writer.Write(payload); err != nil {
		panic("Unable to write payload: " + err.Error())
	}

	// Read the entire file to calculate the digest
	file.Seek(0, 0)
	readData := make([]byte, receivedLen)
	_, err = file.Read(readData)
	if err != nil {
		panic("Unable to read file: " + err.Error())
	}

	// Calculate the digest
	digest := <-pd.PayloadScheme.FromBytes(readData)

	// Commit function to move the file to the final destination
	commit := func(isCompletePayload bool) {
		_, err = pd.EnsureDir("partial")
		if err != nil {
			fmt.Printf("Unable to ensure partial directory: %v\n", err)
		}

		var committedFilePath string
		if isCompletePayload {
			committedFilePath = filepath.Join(pd.path, pd.GetKey(expectedDigest))
			err = os.Rename(stagingFilePath, committedFilePath)
		} else {
			pd.EnsureDir("partial")
			committedFilePath = filepath.Join(pd.path, "partial", pd.GetKey(expectedDigest))
			err = copyFile(stagingFilePath, committedFilePath)
			err = os.Remove(stagingFilePath)
		}
		if err != nil {
			fmt.Println("Unable to commit file:", err)
		} else {
			fmt.Println("File committed successfully")
		}
	}

	// Reject function to delete the staging file
	reject := func() {
		err = os.Remove(stagingFilePath)
		if err != nil {
			fmt.Printf("Unable to remove staging file: %v\n", err)
		} else {
			fmt.Println("Staging file removed successfully")
		}
	}

	return digest, uint64(receivedLen), commit, reject, nil
}

func MakePayloadDriver[PayloadDigest constraints.Ordered, T constraints.Unsigned](pathParam string, payloadSchemeParam datamodeltypes.PayloadScheme[T]) PayloadDriver[T] {
	return PayloadDriver[T]{
		path:          pathParam,
		PayloadScheme: payloadSchemeParam,
	}
}
