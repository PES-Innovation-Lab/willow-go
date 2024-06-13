package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestGrowingBytes_Relative(t *testing.T) {
	fifo := make(chan []byte, 10) // Simulating FIFO buffer

	bytes := NewGrowingBytes(fifo)

	// Initial assertion
	if !reflect.DeepEqual(bytes.Array, []byte{}) {
		t.Errorf("Expected initial Array to be empty, got: %v", bytes.Array)
	}

	fifo <- []byte{0}

	// After pushing [0]
	if !reflect.DeepEqual(bytes.Array, []byte{}) {
		t.Errorf("Expected Array after push [0] to be empty, got: %v", bytes.Array)
	}

	fifo <- []byte{1}
	fifo <- []byte{2, 3}

	time.Sleep(50 * time.Millisecond) // Simulate asynchronous behavior
	receivedBytes := bytes.NextRelative(4)

	// After pushing [1, 2, 3]
	expected := []byte{0, 1, 2, 3}
	if !reflect.DeepEqual(receivedBytes, expected) {
		t.Errorf("Expected Array after push [1, 2, 3] to be %v, got: %v", expected, receivedBytes)
	}

	/* receivedBytes := bytes.NextRelative(4)

		if !reflect.DeepEqual(receivedBytes, expected) {
			t.Errorf("Expected received bytes to be %v, got: %v", expected, receivedBytes)
		}

		lastPromise := bytes.NextRelative(2)

		if !reflect.DeepEqual(bytes.Array, expected) {
			t.Errorf("Expected Array after nextRelative(4) to be %v, got: %v", expected, bytes.Array)
		}

		fifo <- []byte{4, 5}

		time.Sleep(10 * time.Millisecond) // Simulate asynchronous behavior

		expected = []byte{0, 1, 2, 3, 4, 5}
		if !reflect.DeepEqual(lastPromise, expected) {
			t.Errorf("Expected result from last promise to be %v, got: %v", expected, lastPromise)
		}
	}

	func TestGrowingBytes_Absolute(t *testing.T) {
		fifo := make(chan []byte, 10) // Simulating FIFO buffer

		bytes := NewGrowingBytes(fifo)

		// Initial assertion
		if !reflect.DeepEqual(bytes.Array, []byte{}) {
			t.Errorf("Expected initial Array to be empty, got: %v", bytes.Array)
		}

		fifo <- []byte{0}

		time.Sleep(10 * time.Millisecond) // Simulate asynchronous behavior

		// After pushing [0]
		if !reflect.DeepEqual(bytes.Array, []byte{}) {
			t.Errorf("Expected Array after push [0] to be empty, got: %v", bytes.Array)
		}

		fifo <- []byte{1}
		fifo <- []byte{2, 3}

		time.Sleep(10 * time.Millisecond) // Simulate asynchronous behavior

		// After pushing [1, 2, 3]
		expected := []byte{0, 1, 2, 3}
		if !reflect.DeepEqual(bytes.Array, expected) {
			t.Errorf("Expected Array after push [1, 2, 3] to be %v, got: %v", expected, bytes.Array)
		}

		receivedBytes := bytes.NextAbsolute(4)

		if !reflect.DeepEqual(receivedBytes, expected) {
			t.Errorf("Expected received bytes to be %v, got: %v", expected, receivedBytes)
		}

		bytes.Prune(4)

		fifo <- []byte{4}
		fifo <- []byte{5, 6}

		time.Sleep(10 * time.Millisecond) // Simulate asynchronous behavior

		expected = []byte{4, 5, 6}
		if !reflect.DeepEqual(bytes.Array, expected) {
			t.Errorf("Expected Array after prune(4) to be %v, got: %v", expected, bytes.Array)
		}

		lastPromise := bytes.NextAbsolute(4)

		fifo <- []byte{7}

		time.Sleep(10 * time.Millisecond) // Simulate asynchronous behavior

		expected = []byte{4, 5, 6, 7}
		if !reflect.DeepEqual(lastPromise, expected) {
			t.Errorf("Expected result from last promise to be %v, got: %v", expected, lastPromise)
		}
	} */
}
