package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIHDRChunkBytes(t *testing.T) {
	tests := []struct {
		width    uint32
		height   uint32
		expected []byte
	}{
		{
			width:  256,
			height: 120,
			expected: []byte{
				0b0, 0b0, 0b1, 0b0, // width
				0b0, 0b0, 0b0, 0b01111000, // height
				0b1000, // bit depth
				0b10,   // color type
				0b0,    // compression method
				0b0,    // filter method
				0b0,    // interlace method
			},
		},
		{
			width:  10,
			height: 6,
			expected: []byte{
				0b0, 0b0, 0b0, 0b1010, // width
				0b0, 0b0, 0b0, 0b110, // height
				0b1000, // bit depth
				0b10,   // color type
				0b0,    // compression method
				0b0,    // filter method
				0b0,    // interlace method
			},
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%dx%d", tt.width, tt.height)
		t.Run(name, func(t *testing.T) {
			got := IHDRChunk(tt.width, tt.height)
			if bytes.Compare(tt.expected, got) != 0 {
				t.Errorf("expected %+v, but got %+v", tt.expected, got)
			}
		})
	}
}
