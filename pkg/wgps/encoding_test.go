package wgps

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
)

func TestEncodingDecoding(t *testing.T) {
	tc := []struct {
		values []struct {
			Value struct {
				FingerPrint string
			}
			FingerPrint string
			Size        uint64
		}
	}{
		{
			values: []struct {
				Value struct {
					FingerPrint string
				}
				FingerPrint string
				Size        uint64
			}{
				{
					Value: struct {
						FingerPrint string
					}{
						FingerPrint: "Samarth",
					},
					FingerPrint: "Samarth",
					Size:        100,
				},
				{
					Value: struct {
						FingerPrint string
					}{
						FingerPrint: "Manas",
					},
					FingerPrint: "Manas",
					Size:        200,
				},
			},
		},
	}
	for _, tt := range tc {
		for i, val := range tt.values {
			encoded := Encode(val.FingerPrint, val.Size, val.Value)
			FingerPrint, Size, Value := Decode(encoded)
			if reflect.DeepEqual(FingerPrint, val.FingerPrint) == false {
				t.Errorf("Test %d: Expected %s, got %s", i, val.FingerPrint, FingerPrint)
			}
			if reflect.DeepEqual(Size, val.Size) == false {
				t.Errorf("Test %d: Expected %d, got %d", i, val.Size, Size)
			}
			if reflect.DeepEqual(Value, val.Value) == false {
				t.Errorf("Test %d: Expected %v, got %v", i, val.Value, Value)
			}
		}
	}
}

func Encode(FingerPrint string, Size uint64, Value struct {
	FingerPrint string
}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	// gob.Register("")
	// gob.Register(uint64(0))

	// encoder.Encode(FingerPrint)
	// encoder.Encode(Size)
	encoder.Encode(struct {
		FingerPrint string
		Size        uint64
		Value       struct {
			FingerPrint string
		}
	}{FingerPrint, Size, Value})
	return buffer.Bytes()
}

func Decode(value []byte) (FingerPrint string, Size uint64, Value struct {
	FingerPrint string
}) {
	var decoded struct {
		FingerPrint string
		Size        uint64
		Value       struct {
			FingerPrint string
		}
	}
	buffer := bytes.NewBuffer(value)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return decoded.FingerPrint, decoded.Size, decoded.Value
}
