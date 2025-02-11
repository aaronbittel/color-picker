package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

var (
	PNG_SIGNATURE = []byte{137, 80, 78, 71, 13, 10, 26, 10}
)

func main() {
	var filepath string
	flag.StringVar(&filepath, "file", "", "png file to get the colors from")
	flag.Parse()

	if filepath == "" || path.Ext(filepath) != ".png" {
		fmt.Println("error: no png file given\nUSAGE: go run . -file <png-file>")
		os.Exit(1)
	}

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("error opening file (%s): %v", filepath, err)
	}
	defer f.Close()

	if err = read_signature(f); err != nil {
		log.Fatalf("error reading signature: %v", err)
	}
	log.Println("successfully read png signuature")

	ihdrChunk, err := read_chunk(f)
	if err != nil {
		log.Fatalf("error reading IHDR chunk: %v", err)
	}
	ihdr := NewIHDR(ihdrChunk)

	fmt.Println(ihdr.chunk.data)

	if ihdr.compressionMethod != 0 {
		panic("error: compression method 0 (deflate/inflate) is the only supported compression method at the moment")
	}

	if ihdr.filterMethod != 0 {
		panic("error: filter method 0 is the only supported filter method at the moment")
	}

	if ihdr.interlaceMethod != 0 && ihdr.interlaceMethod != 1 {
		panic("error: interlace methods 1 or 2 are the only supported interlace methods at the moment")
	}

	log.Println(ihdr)
	log.Println("successfully read ihdr chunk")

outer:

	for {
		chunk, err := read_chunk(f)
		if err != nil {
			log.Printf("error reading chunk: %v\n", err)
			break
		}

		switch chunk.chunkType.String() {
		case "IEND":
			log.Println("IEND => end of png")
			break outer
		case "IDAT":
			fmt.Println("before      ", chunk.data)
			data, err := decompress_zlib(chunk.data)
			if err != nil {
				log.Fatal(err)
			}
			// fmt.Println("decompressed", data)
			unfilteredData, err := filter(data, ihdr.width, ihdr.height)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(unfilteredData)
		default:
			log.Printf("read %s chunk\n", chunk.chunkType)
		}
	}
}

type Filter byte

const (
	NONE Filter = iota
	SUB
	UP
	AVERAGE
	PAETH
)

func subFilter(data []byte) []byte {
	row := []byte{data[0]}
	for i := 1; i < len(data); i++ {
		row = append(row, data[i]-data[i-1])
	}
	return row
}

func unSubFilter(data []byte) []byte {
	row := []byte{data[0]}
	for i := 1; i < len(data); i++ {
		row = append(row, data[i]+data[i-1])
	}
	return row
}

func upFilter(prev, data []byte) []byte {
	row := []byte{}
	for i := 0; i < len(data); i++ {
		row = append(row, data[i]-prev[i])
	}
	return row
}

func unUpFilter(prev, data []byte) []byte {
	row := []byte{}
	for i := 0; i < len(data); i++ {
		row = append(row, data[i]+prev[i])
	}
	return row
}

func get_or(data []byte, idx int, def byte) byte {
	if idx < 0 || idx >= len(data) {
		return def
	}
	return data[idx]
}

func paethFilter(prev, data []byte) []byte {
	row := []byte{}
	for i := 0; i < len(data); i++ {
		L := get_or(data, i-1, 0)
		U := get_or(prev, i, 0)
		UL := get_or(prev, i-1, 0)
		v := U + L - UL

		m := min(v-L, v-U, v-UL)
		row = append(row, data[i]-m)
	}
	return row
}

func filter(data []byte, width, height int) ([]byte, error) {
	log.Println("WIDTH", width, "HEIGHT", height, "DATA LEN", len(data))
	unfilteredData := make([]byte, width*height)

	rowSize := 1 + (width * 3)

	for i := 0; i < height; i++ {
		start := i * rowSize
		end := (i + 1) * rowSize
		log.Println("START", start, "END", end)

		row := data[start+1 : end]
		filterByte := Filter(data[start])
		switch filterByte {
		case NONE:
			unfilteredData = append(unfilteredData, row...)
		case SUB:
			unfilteredData = append(unfilteredData, unSubFilter(row)...)
		case UP:
			if i == 0 {
				unfilteredData = append(unfilteredData, row...)
				continue
			}
			prev := data[start-rowSize : start]
			unfilteredData = append(unfilteredData, unUpFilter(prev, row)...)
		case AVERAGE:
			panic("TODO AVERAGE FILTER")
		case PAETH:
			panic("TODO PAETH FILTER")
		default:
			panic(fmt.Sprintf("error: unknown filter type: %d", filterByte))
		}
	}

	return unfilteredData, nil
}

func decompress_zlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error: reading zlib data: %v", err)
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

type ChunkType [4]byte

func (c ChunkType) String() string {
	return string(c[:])
}

type ColorType byte

func (c ColorType) String() string {
	switch c {
	case 0:
		return "nothing"
	case 2:
		return "color"
	case 3:
		return "palette & color"
	case 4:
		return "alpha channel"
	case 6:
		return "color & alpha channel"
	default:
		panic(fmt.Sprintf("unknown colortype: %d", c))
	}
}

type IHDR struct {
	chunk Chunk

	width             int
	height            int
	bitDepth          byte
	colorType         ColorType
	compressionMethod byte
	filterMethod      byte
	interlaceMethod   byte
}

func (ihdr IHDR) String() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("\n\tPNG width: %d, height: %d\n", ihdr.width, ihdr.height))
	sb.WriteString(fmt.Sprintf("\tbit depth: %d\n", ihdr.bitDepth))
	sb.WriteString(fmt.Sprintf("\tcolor type: %s\n", ihdr.colorType))
	sb.WriteString(fmt.Sprintf("\tcompression method: %d\n", ihdr.compressionMethod))
	sb.WriteString(fmt.Sprintf("\tfilter method: %d\n", ihdr.filterMethod))
	sb.WriteString(fmt.Sprintf("\tinterlace method: %d\n", ihdr.interlaceMethod))

	return sb.String()
}

func read_chunk_data(f *os.File, size int) ([]byte, error) {
	chunkDataBytes, err := read_n_bytes(f, size)
	if err != nil {
		return nil, fmt.Errorf("error reading %d bytes: %v", size, err)
	}

	return chunkDataBytes, nil
}

func read_chunk_crc(f *os.File) (uint32, error) {
	chunkCRC, err := read_n_bytes(f, 4)
	if err != nil {
		return 0, fmt.Errorf("error: reading 4 bytes chunk crc: %v", err)
	}

	return binary.BigEndian.Uint32(chunkCRC), nil
}

func read_chunk_size(f *os.File) (uint32, error) {
	chunkSizeBytes, err := read_n_bytes(f, 4)
	if err != nil {
		return 0, fmt.Errorf("error reading 4 bytes chunk size: %v", err)
	}
	// fmt.Println("chunk size bytes: ", chunkSizeBytes)
	// fmt.Println("chunk size int  : ", binary.BigEndian.Uint32(chunkSizeBytes))

	return binary.BigEndian.Uint32(chunkSizeBytes), nil
}

func read_chunk_type(f *os.File) (ChunkType, error) {
	var chunkType [4]byte
	chunkTypeBytes, err := read_n_bytes(f, 4)
	if err != nil {
		return ChunkType{}, fmt.Errorf("error reading 4 bytes chunk type: %v", err)
	}
	copy(chunkType[:], chunkTypeBytes)

	return chunkType, nil
}

type Chunk struct {
	size      uint32
	chunkType ChunkType
	data      []byte
	crc       uint32
}

func read_chunk(f *os.File) (Chunk, error) {
	chunkSize, err := read_chunk_size(f)
	if err != nil {
		return Chunk{}, err
	}

	chunkType, err := read_chunk_type(f)
	if err != nil {
		return Chunk{}, err
	}

	chunkData, err := read_chunk_data(f, int(chunkSize))
	if err != nil {
		return Chunk{}, err
	}

	chunkCRC, err := read_chunk_crc(f)
	if err != nil {
		return Chunk{}, err
	}

	fmt.Println("total chunk size:", 4+4+4+chunkSize)

	// crcBytes := append(chunkType[:], chunkData...)
	// fmt.Println("calculated", crc32.ChecksumIEEE(crcBytes), "expected", chunkCRC)

	return Chunk{
		size:      uint32(chunkSize),
		chunkType: chunkType,
		data:      chunkData,
		crc:       chunkCRC,
	}, nil
}

func read_signature(f *os.File) error {
	signature, err := read_n_bytes(f, 8)
	if err != nil {
		return err
	}

	if !slices.Equal(PNG_SIGNATURE, signature) {
		panic(fmt.Sprintf("panic expected png signature %v, got %v", PNG_SIGNATURE, signature))
	}

	return nil
}

func read_n_bytes(f *os.File, size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := f.Read(buf)
	if err != nil {
		// log.Fatalf("error reading %d bytes: %v", size, err)
		log.Printf("error reading %d bytes: %v", size, err)
		return nil, err
	}

	if n != size {
		panic(fmt.Sprintf("panic expected png signature length to be %d, but read %d", size, n))
	}

	return buf, nil
}

func NewIHDR(chunk Chunk) IHDR {
	if chunk.chunkType.String() != "IHDR" {
		panic(fmt.Sprintf("error: trying to convert chunk to IHDR, but type is %v (%s)",
			chunk.chunkType, chunk.chunkType))
	}

	if chunk.size != 13 {
		panic(fmt.Sprintf(
			"error: trying to convert chunk to IHDR. Size is expected to be 13, but got %d",
			chunk.size))
	}

	return IHDR{
		chunk: chunk,

		width:             int(binary.BigEndian.Uint32(chunk.data[:4])),
		height:            int(binary.BigEndian.Uint32(chunk.data[4:8])),
		bitDepth:          chunk.data[8],
		colorType:         ColorType(chunk.data[9]),
		compressionMethod: chunk.data[10],
		filterMethod:      chunk.data[11],
		interlaceMethod:   chunk.data[12],
	}
}
