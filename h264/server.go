package h264

import (
	// "github.com/nareix/joy4/av"
	// 	"github.com/nareix/joy4/codec/h264parser"
	// "github.com/nareix/joy4/format/ts"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

// InitialNALU indicates the start of a h264 packet
// 0 - Forbidden 0 bit; always 0
// 1,2 - NRI
// 3,4,5,6,7 - Type
var (
	InitialNALU   = []byte{0, 0, 0, 1}
	Initial3BNALU = []byte{0, 0, 1}
	logger        *log.Logger
	streamOffset  = 0
)

func init() {
	logger = log.New(os.Stderr, "streamer ", log.Lshortfile)
}
func isNALU(packet []byte) bool {
	var is3bNalu, is4bNalu bool
	if len(packet) < len(InitialNALU) {
		return false
	}
	nalu3BSegment := packet[len(packet)-3:]
	for i := range Initial3BNALU {
		if nalu3BSegment[i] != Initial3BNALU[i] {
			return false
		}
	}
	is3bNalu = true
	// The first NALU is a 3 or 4 byte packet, all others have the NALU
	// in the packet's tailing 4 bytes
	naluSegment := packet[len(packet)-4:]
	for i := range InitialNALU {
		if naluSegment[i] != InitialNALU[i] {
			return false
		}
	}
	is4bNalu = true
	return is3bNalu || is4bNalu
}

type RBSPReader struct {
	BitStream      chan []int
	BitRequestLine chan int
	*StreamReader
}

func (r *RBSPReader) Stream() {
	rbsp := []byte{}
	for !isNALU(rbsp) {
		rbsp = r.GetBytes(1)
	}
}

func handleConnection(frameCounter *counter, connection io.Reader) {
	endStreamChan := make(chan int, 1)
	rbspReader := &RBSPReader{BitRequestLine: make(chan int, 1)}
	streamReader := NewStreamReader(connection, rbspReader.BitRequestLine, endStreamChan)
	rbspReader.StreamReader = &streamReader
	rbspReader.BitStream = streamReader.BitStreamChan
	// Handle bit arrays and allow requests for bits
	go rbspReader.Stream()
	// Blocking
	streamReader.Stream()
}

func ByteStreamReader(connection net.Conn) {
	logger.Printf("opened bytestream\n")
	frameCounter := &counter{0}
	defer connection.Close()
	handleConnection(frameCounter, connection)
}
