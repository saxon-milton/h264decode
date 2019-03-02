package h264

import (
	// "github.com/nareix/joy4/av"
	// 	"github.com/nareix/joy4/codec/h264parser"
	// "github.com/nareix/joy4/format/ts"
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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

type RBSPReader struct {
	BitStream      chan []int
	BitRequestLine chan int
	*StreamReader
}

func init() {
	logger = log.New(os.Stderr, "streamer ", log.Lshortfile|log.Lmicroseconds)
}
func isStartSequence(packet []byte) bool {
	if len(packet) < len(InitialNALU) {
		return false
	}
	naluSegment := packet[len(packet)-4:]
	for i := range InitialNALU {
		if naluSegment[i] != InitialNALU[i] {
			return false
		}
	}
	return true
}
func isEmpty3Byte(buf []byte) bool {
	if len(buf) < 3 {
		return false
	}
	for _, i := range buf[len(buf)-3:] {
		if i != 0 {
			return false
		}
	}
	return true
}
func isStartCodeOnePrefix(buf []byte) bool {
	for i, b := range buf {
		if i < 2 && b != byte(0) {
			return false
		}
		// byte 3 may be 0 or 1
		if i == 3 && b != byte(0) || b != byte(1) {
			return false
		}
	}
	logger.Printf("debug: found start code one prefix byte\n")
	return true
}
func readNalUnit(r *RBSPReader) *NalUnit {
	nalUnitBuffer := &bytes.Buffer{}
	// for !isEmpty3Byte(nalUnitBuffer.Bytes()) && !isStartSequence(nalUnitBuffer.Bytes()) {
	for !isStartSequence(nalUnitBuffer.Bytes()) {
		if buf, err := r.GetBytes(1); err != nil {
			r.LogStreamPosition()
			r.endStreamChan <- 1
			return nil
		} else {
			nalUnitBuffer.Write(buf)
		}
	}
	// Annex B.2 Step 1
	//	_, _ = r.GetBytes(1)
	// Annex B.2 Step 2
	//	_, _ = r.GetBytes(3)
	r.LogStreamPosition()

	// Read the nalUnit
	nalUnitReader := &RBSPReader{BitRequestLine: make(chan int, 1)}
	endNaluChan := make(chan int, 1)
	nalUnitStream := NewStreamReader(nalUnitBuffer, nalUnitReader.BitRequestLine, endNaluChan)
	nalUnitStream.streamFile = r.streamFile
	nalUnitReader.StreamReader = nalUnitStream
	nalUnitReader.BitStream = nalUnitStream.BitStreamChan
	go nalUnitStream.Stream()
	logger.Printf("debug: found NALU unit with %d bytes\n\t%#v",
		nalUnitBuffer.Len(),
		nalUnitBuffer.Bytes()[nalUnitBuffer.Len()-4:])
	if isStartCodeOnePrefix(nalUnitBuffer.Bytes()[nalUnitBuffer.Len()-3:]) {
		logger.Printf("info: Nal unit ends in 0x000000 or 0x000001: %#v\n",
			nalUnitBuffer.Bytes()[nalUnitBuffer.Len()-3:])
	}
	nalUnit := NewNalUnit(nalUnitReader, nalUnitBuffer.Len()-4)
	endNaluChan <- 1
	return nalUnit
}
func (r *RBSPReader) Run() {
	rbsp := []byte{}
	// Find start of stream
	for !isStartSequence(rbsp) {
		buf, err := r.GetBytes(1)
		if err != nil {
			r.endStreamChan <- 1
			return
		}
		rbsp = append(rbsp, buf...)
	}

	for {
		_ = readNalUnit(r)
		// End read the nalUnit
		r.LogStreamPosition()
		logger.Printf("info: read NAL unit\n")
	}
}

func handleConnection(connection io.Reader) {
	logger.Printf("debug: handling connection\n")
	streamFilename := "/home/bruce/devel/go/src/github.com/mrmod/cvnightlife/output.mp4"
	endStreamChan := make(chan int, 1)
	_ = os.Remove(streamFilename)
	streamFile, err := os.Create(streamFilename)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		logger.Printf("debug: waiting on signals\n")
		s := <-c
		logger.Printf("info: %v received, closing stream file\n", s)
		streamFile.Close()
		os.Exit(0)
	}()
	rbspReader := &RBSPReader{BitRequestLine: make(chan int, 1)}
	streamReader := NewStreamReader(connection, rbspReader.BitRequestLine, endStreamChan)
	streamReader.streamFile = streamFile
	rbspReader.StreamReader = streamReader
	rbspReader.BitStream = streamReader.BitStreamChan
	// Handle bit arrays and allow requests for bits
	go rbspReader.Run()
	// Blocking
	streamReader.Stream()
}

func ByteStreamReader(connection net.Conn) {
	logger.Printf("opened bytestream\n")
	defer connection.Close()
	handleConnection(connection)
}
