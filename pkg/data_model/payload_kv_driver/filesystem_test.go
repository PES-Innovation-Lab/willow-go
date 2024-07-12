package payloadDriver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/PES-Innovation-Lab/willow-go/pkg/data_model/datamodeltypes"
	"github.com/PES-Innovation-Lab/willow-go/utils"
)

// Mock PayloadScheme for testing
var mockPayloadScheme datamodeltypes.PayloadScheme[string, uint64] = datamodeltypes.PayloadScheme[string, uint64]{
	EncodingScheme: utils.EncodingScheme[string, uint64]{
		Encode: func(value string) []byte {
			decoded, err := hex.DecodeString(value)
			if err != nil {
				return []byte{}
			}
			return decoded
		},
	},
	FromBytes: func(bytes []byte) chan string {
		ch := make(chan string, 1)
		go func() {
			hash := sha256.Sum256(bytes)
			ch <- hex.EncodeToString(hash[:])
			close(ch)
		}()
		return ch
	},
}

// Helper function to create a temporary directory
func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "payload_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tempDir
}

// func TestGetKey(t *testing.T) {
// 	pd := PayloadDriver[string, uint64]{
// 		PayloadScheme: mockPayloadScheme,
// 	}

// 	hash := "0123456789abcdef"
// 	expected := base32.StdEncoding.EncodeToString(mockPayloadScheme.Encode(hash))
// 	result := pd.GetKey(hash)

// 	if result != expected {
// 		t.Errorf("getKey(%s) = %s; want %s", hash, result, expected)
// 	}
// }

func TestGetPayload(t *testing.T) {
	// Define the path where your test files are stored
	testDataPath := "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433"

	// Define the files you want to test
	testFiles := []struct {
		name     string
		fileName string
	}{
		{"Text file", "sample.txt"},
		// {"PDF file", "sample.pdf"},
		{"MP4 file", "sample.mp4"},
	}

	pd := PayloadDriver[string, uint64]{
		path: testDataPath,
	}

	for _, tf := range testFiles {
		t.Run(tf.name, func(t *testing.T) {
			filePath := filepath.Join(testDataPath, tf.fileName)

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Fatalf("Test file does not exist: %s", filePath)
			}

			payload := pd.GetPayload(filePath)

			// Print the first 50 bytes of the payload contents (or less if the file is smaller)
			fmt.Printf("The first 50 bytes of %s are: %v\n", tf.fileName, payload.Bytes()[:min(50, len(payload.Bytes()))])

			// Test BytesWithOffset function
			offsetBytes, err := payload.BytesWithOffset(5)
			if err != nil {
				t.Errorf("BytesWithOffset failed for %s: %v", tf.fileName, err)
			} else {
				fmt.Printf("The first 45 bytes with offset 5 for %s are: %v\n", tf.fileName, offsetBytes[:min(45, len(offsetBytes))])
			}

			// Test Length function
			length, err := payload.Length()
			if err != nil {
				t.Errorf("Length failed for %s: %v", tf.fileName, err)
			} else {
				fmt.Printf("The payload length for %s is: %d\n", tf.fileName, length)
			}

			// Additional checks
			if length == 0 {
				t.Errorf("Payload for %s is empty", tf.fileName)
			}

			// Check if BytesWithOffset returns expected length
			if len(offsetBytes) != int(length)-5 {
				t.Errorf("BytesWithOffset returned unexpected length for %s. Got %d, expected %d",
					tf.fileName, len(offsetBytes), int(length)-5)
			}
		})
	}
}

// Helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestSetAndGetPayload(t *testing.T) {
	fmt.Println("Starting TestSetAndGetPayload...")

	// Create a temporary directory for this test
	tempDir := createTempDir(t)
	fmt.Printf("Created temporary directory: %s\n", tempDir)
	defer func() {
		// os.RemoveAll(tempDir)
		fmt.Printf("Cleaned up temporary directory: %s\n", tempDir)
	}()

	// Initialize PayloadDriver
	pd := PayloadDriver[string, uint64]{
		path:          tempDir,
		PayloadScheme: mockPayloadScheme,
	}
	fmt.Println("Initialized PayloadDriver")

	// Create test content
	testContent := []byte("This is a test payload content. It includes some numbers 12345 and symbols !@#$%.")
	fmt.Printf("Created test content: %s\n", string(testContent))

	// Set the payload
	fmt.Println("Setting payload...")
	hash, setPayload, size := pd.Set(testContent)
	fmt.Printf("Set payload with hash: %s\n", hash)
	fmt.Printf("Set payload size: %d bytes\n", size)
	fmt.Printf("Set payload content: %s\n", string(setPayload.Bytes()))

	// Get the payload
	fmt.Printf("Getting payload with hash: %s\n", hash)
	getPayload, err := pd.Get(hash)
	if err != nil {
		t.Fatalf("Failed to get payload: %v", err)
	}
	fmt.Println("Successfully got payload")

	// Compare payloads
	fmt.Println("Comparing set and get payloads...")
	if !bytes.Equal(setPayload.Bytes(), getPayload.Bytes()) {
		t.Errorf("Get payload doesn't match set payload")
		fmt.Printf("Set payload: %s\n", string(setPayload.Bytes()))
		fmt.Printf("Get payload: %s\n", string(getPayload.Bytes()))
	} else {
		fmt.Println("Set and get payloads match!")
	}

	// Check payload length
	setLength, _ := setPayload.Length()
	getLength, _ := getPayload.Length()
	fmt.Printf("Set payload length: %d\n", setLength)
	fmt.Printf("Get payload length: %d\n", getLength)
	if setLength != getLength {
		t.Errorf("Payload lengths don't match. Set: %d, Get: %d", setLength, getLength)
	} else {
		fmt.Println("Payload lengths match!")
	}

	// Test BytesWithOffset
	offset := 10
	fmt.Printf("Testing BytesWithOffset with offset %d...\n", offset)
	setOffsetBytes, _ := setPayload.BytesWithOffset(offset)
	getOffsetBytes, _ := getPayload.BytesWithOffset(offset)
	fmt.Printf("Set payload with offset: %s\n", string(setOffsetBytes))
	fmt.Printf("Get payload with offset: %s\n", string(getOffsetBytes))
	if !bytes.Equal(setOffsetBytes, getOffsetBytes) {
		t.Errorf("Offset bytes don't match")
	} else {
		fmt.Println("Offset bytes match!")
	}

	fmt.Println("TestSetAndGetPayload completed successfully!")
}

func TestGet(t *testing.T) {
	pd := PayloadDriver[string, uint64]{
		path:          "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433",
		PayloadScheme: mockPayloadScheme,
	}

	// Read the video file content
	videoPath := "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433\\sample.mp4"
	videoContent, err := os.ReadFile(videoPath)
	if err != nil {
		t.Fatalf("failed to read video file: %v", err)
	}

	// Set the video content
	hash, _, _ := pd.Set(videoContent)
	fmt.Println("Finished Setting")
	// Get the payload
	payload, err := pd.Get(hash)
	if err != nil {
		t.Errorf("get failed: %v", err)
	}

	if !bytes.Equal(payload.Bytes(), videoContent) {
		t.Errorf("get(%s) content = %v; want %v", hash, payload.Bytes(), videoContent)
	}

	fmt.Println("Done now writing file")
	// Write the retrieved payload to output.mp4
	outputPath := "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433\\output"
	err = os.WriteFile(outputPath, payload.Bytes(), 0644)
	if err != nil {
		t.Fatalf("failed to write output file: %v", err)
	}

	// Verify the written file content
	writtenContent, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !bytes.Equal(writtenContent, videoContent) {
		t.Errorf("output file content = %v; want %v", writtenContent, videoContent)
	}
}

func TestErase(t *testing.T) {
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	pd := PayloadDriver[string, uint64]{
		path:          tempDir,
		PayloadScheme: mockPayloadScheme,
	}

	testContent := []byte("test content")
	hash, _, _ := pd.Set(testContent)

	success, err := pd.Erase(hash)
	if err != nil {
		t.Errorf("erase failed: %v", err)
	}
	if !success {
		t.Errorf("erase(%s) = false; want true", hash)
	}

	// Check if file is actually erased
	_, err = pd.Get(hash)
	if err == nil {
		t.Errorf("File still exists after erase")
	}
}

func TestSet(t *testing.T) {
	// tempDir := createTempDir(t)
	// defer os.RemoveAll(tempDir)

	pd := PayloadDriver[string, uint64]{
		path:          "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433",
		PayloadScheme: mockPayloadScheme,
	}

	testContent := []byte("test content")
	hash, payload, size := pd.Set(testContent)

	if size != uint64(len(testContent)) {
		t.Errorf("set() size = %d; want %d", size, len(testContent))
	}

	if !bytes.Equal(payload.Bytes(), testContent) {
		t.Errorf("set() content = %v; want %v", payload.Bytes(), testContent)
	}

	// Check if file is actually created
	_, err := pd.Get(hash)
	if err != nil {
		t.Errorf("File not created after set: %v", err)
	}
}
func TestSetVido(t *testing.T) {
	pd := PayloadDriver[string, uint64]{
		path:          "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433",
		PayloadScheme: mockPayloadScheme,
	}

	// Read the video file content
	videoPath := "C:\\Users\\samar\\AppData\\Local\\Temp\\payload_test415599433\\sample.mp4"
	videoContent, err := ioutil.ReadFile(videoPath)
	if err != nil {
		t.Fatalf("failed to read video file: %v", err)
	}

	// Set the video content
	hash, payload, size := pd.Set(videoContent)

	if size != uint64(len(videoContent)) {
		t.Errorf("set() size = %d; want %d", size, len(videoContent))
	}

	if !bytes.Equal(payload.Bytes(), videoContent) {
		t.Errorf("set() content = %v; want %v", payload.Bytes(), videoContent)
	}

	// Check if file is actually created
	_, err = pd.Get(hash)
	if err != nil {
		t.Errorf("File not created after set: %v", err)
	}
}
func TestEnsureDir(t *testing.T) {
	tempDir := createTempDir(t)
	// defer os.RemoveAll(tempDir)

	pd := PayloadDriver[string, uint64]{
		path: tempDir,
	}

	testPath := "test/path"
	createdPath, err := pd.EnsureDir(testPath)
	if err != nil {
		t.Errorf("ensureDir failed: %v", err)
	}

	expectedPath := filepath.Join(tempDir, testPath)
	if createdPath != expectedPath {
		t.Errorf("ensureDir(%s) = %s; want %s", testPath, createdPath, expectedPath)
	}

	// Check if directory is actually created
	if _, err := os.Stat(createdPath); os.IsNotExist(err) {
		t.Errorf("Directory not created: %v", err)
	}
}

func TestReceive(t *testing.T) {

	// Create a temporary directory for this test
	tempDir := createTempDir(t)
	// defer func() {
	// 	os.RemoveAll(tempDir)
	// }()

	// Initialize PayloadDriver
	pd := PayloadDriver[string, uint64]{
		path:          tempDir,
		PayloadScheme: mockPayloadScheme,
	}
	// Create test content
	testContent := []byte("This is a test payload content. It includes some numbers 12345 and symbols !@#$%.")
	additionalContent := []byte(" Additional data.")
	expectedDigest := <-mockPayloadScheme.FromBytes(append(testContent, additionalContent...))
	expectedDigest_partial := <-mockPayloadScheme.FromBytes(testContent)

	// Test receiving the payload
	offset := int64(0)
	expectedLength := uint64(len(testContent))
	receivedDigest, receivedLength, commit, _, err := pd.Receive(testContent, offset, expectedLength, expectedDigest)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}

	// Verify the received digest and length
	if receivedDigest != expectedDigest_partial {
		t.Errorf("Received digest = %s; want %s", receivedDigest, expectedDigest_partial)
	}
	if receivedLength != expectedLength {
		t.Errorf("Received length = %d; want %d", receivedLength, expectedLength)
	}

	// Test committing the payload
	commit(false)

	// //Verify that the payload is stored correctly
	// storedPayload, err := pd.Get(expectedDigest)
	// if err != nil {
	// 	t.Fatalf("Failed to get stored payload: %v", err)
	// }
	// if !bytes.Equal(storedPayload.Bytes(), testContent) {
	// 	t.Errorf("Stored payload content = %s; want %s", storedPayload.Bytes(), testContent)
	// }

	// commit(false)
	// Test receiving additional data to simulate partial payload
	// expectedDigestWithAdditional := <-mockPayloadScheme.FromBytes(fmt.Append(testContent, additionalContent))
	fmt.Printf("%s", append(testContent, additionalContent...))
	fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	_, _, _, _, err = pd.Receive(additionalContent, int64(len(testContent)), uint64(len(testContent)+len(additionalContent)), expectedDigest)
	if err != nil {
		t.Fatalf("Receive failed for additional data: %v", err)
	}

	// Verify that the partial payload is stored correctly
	partialFilePath := filepath.Join(tempDir, "partial", pd.GetKey(expectedDigest))
	if _, err := os.Stat(partialFilePath); os.IsNotExist(err) {
		t.Fatalf("Partial payload file does not exist: %s", partialFilePath)
	}

	// // Test rejecting the payload
	// // reject()
	// // if _, err := os.Stat(partialFilePath); err == nil {
	// // 	t.Fatalf("Partial payload file still exists after reject: %s", partialFilePath)
	// // }

	// Verify that the commit function works correctly
	finalContent := append(testContent, additionalContent...)
	_, _, commit, _, err = pd.Receive(finalContent, 0, uint64(len(finalContent)), expectedDigest)
	if err != nil {
		t.Fatalf("Receive failed for final content: %v", err)
	}
	commit(true)

	storedPayload, err := pd.Get(expectedDigest)
	if err != nil {
		t.Fatalf("Failed to get final stored payload: %v", err)
	}
	if !bytes.Equal(storedPayload.Bytes(), finalContent) {
		t.Errorf("Final stored payload content = %s; want %s", storedPayload.Bytes(), finalContent)
	}

}
