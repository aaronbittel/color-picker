package main

import (
	"encoding/binary"
	"hash/crc32"
)

type RGB struct {
	red   byte
	green byte
	blue  byte
}

type Image struct {
	width  uint32
	height uint32
	pixels [][]RGB
}

func NewImage(width, height uint32) Image {
	pixels := make([][]RGB, height)
	for i := range height {
		pixels[i] = make([]RGB, width)
	}

	return Image{
		width:  width,
		height: height,
		pixels: pixels,
	}
}

func (i *Image) Fill(color RGB) {
	for y := range i.height {
		for x := range i.width {
			i.pixels[y][x] = color
		}
	}
}

func (i Image) Bytes() []byte {
	buf := make([]byte, 3*i.height*i.width)

	for y := range i.height {
		for x := range i.width {
			p := i.pixels[y][x]
			buf = append(buf, p.red, p.green, p.blue)
		}
	}

	return buf
}

func Decode(data []byte, width, height uint32) []byte {
	var png []byte

	png = append(png, PNG_SIGNATURE...)
	png = append(png, IHDRChunk(width, height)...)
	png = append(png, IDATChunk()...)
	png = append(png, IENDChunk()...)

	return png
}

func IDATChunk() []byte {
	var buf []byte

	// length
	buf = append(buf, 0)

	// chunk type
	chunkType := []byte{73, 68, 65, 84}
	buf = append(buf, chunkType...)

	// chunk data
	var chunkData byte = 0
	buf = append(buf, chunkData)

	// crc
	buf = append(buf, byte(crc32.ChecksumIEEE(append(chunkType, chunkData))))

	return buf
}

func IENDChunk() []byte {
	var buf []byte

	// length
	buf = append(buf, 0)

	// chunk type
	chunkType := []byte{73, 69, 78, 68}
	buf = append(buf, chunkType...)

	// chunk data
	var chunkData byte = 0
	buf = append(buf, chunkData)

	// crc
	buf = append(buf, byte(crc32.ChecksumIEEE(append(chunkType, chunkData))))

	return buf
}

func uint32ToBytesBE(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	return buf
}

func IHDRChunk(width, height uint32) []byte {
	var buf []byte

	buf = append(buf, uint32ToBytesBE(width)...)
	buf = append(buf, uint32ToBytesBE(height)...)

	// INFO: bit depth 8 or 16?
	buf = append(buf, 8)
	// color type
	buf = append(buf, 2)
	// compression method
	buf = append(buf, 0)
	// filter method
	buf = append(buf, 0)
	// interlace method
	buf = append(buf, 0)

	return buf
}
