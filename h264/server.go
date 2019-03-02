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

type counter struct{ c int }

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

// read bytes until a new header appears
func h264SegmentReader(r io.Reader) ([]byte, error) {
	packet := []byte{}
	for !isNALU(packet) {
		buf := make([]byte, 1)
		n, err := r.Read(buf)
		streamOffset += n
		packet = append(packet, buf...)
		if err != nil {
			return packet, err
		}
	}
	logger.Printf("offset: %d read %d byte RBSP\n", streamOffset, len(packet))
	return packet, nil
}

// read bytes between packets
func h264Demuxer(r io.Reader, frames chan []byte) {
	// Read the opening NALU
	packet, err := h264SegmentReader(r)
	if err != nil {
		logger.Printf("head segment error %v\n", err)
		frames <- packet
		close(frames)
		return
	}
	packetCounter := 0
	// Packet is exactly a 4 byte NALU
	for {
		// Read the frame to the next NALU boundary
		segment, err := h264SegmentReader(r)
		packet = append(packet, segment...)
		if err != nil {
			logger.Printf("error demuxing: %v\n", err)
			frames <- packet
			close(frames)
			return
		}
		// Remove the header of the next packet
		packet = packet[0 : len(packet)-4]
		// logger.Printf("(%d) packet %v\n", packetCounter, packet)
		frames <- packet
		packetCounter++
		// Add the NALU header to the next packet
		packet = append([]byte{}, InitialNALU...)
	}
}
func handleConnection(frameCounter *counter, connection io.Reader) {
	bitRequestChan := make(chan int, 1)
	endStreamChan := make(chan int, 1)
	streamReader := NewStreamReader(connection, bitRequestChan, endStreamChan)

	// Handle bit arrays and allow requests for bits
	go rbspHandler(streamReader.BitStreamChan, bitRequestChan)
	// Block
	streamReader.Stream()
}

func rbspHandler(bitArrayChan chan []int, bitRequestChan chan int) {

}
func oldHandleConnection(frameCounter *counter, h264stream io.Reader) {
	var sps SPS
	var pps PPS
	frameFile, err := os.OpenFile("output.mp4", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Printf("Failed to open output.mp4: %v\n", err)
		return
	}

	defer frameFile.Close()
	frames := make(chan []byte, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for frame := range frames {
			// Drop leading 0x0, 0x0, 0x1, NALUTypeByte
			nalUnit := NewNalUnit(frame[4:])
			rbsp := NewRBSP(frame)
			logger.Printf("NALUTYPE[%d]%s FRAME: %d RBSP: %d\n",
				nalUnit.Type,
				NALUnitType[nalUnit.Type],
				frameCounter.c,
				len(rbsp))
			switch nalUnit.Type {
			case NALU_TYPE_SPS:
				sps = NewSPS(rbsp)
			case NALU_TYPE_PPS:
				pps = NewPPS(&sps, rbsp)
			case NALU_TYPE_SLICE_IDR_PICTURE:
				_ = NewSliceContext(&nalUnit, &sps, &pps, rbsp)
			case NALU_TYPE_SLICE_NON_IDR_PICTURE:
				_ = NewSliceContext(&nalUnit, &sps, &pps, rbsp)
			default:
				logger.Printf("== SKIP: %d:%s\n", nalUnit.Type, NALUnitType[nalUnit.Type])
			}

			_, _ = frameFile.Write(frame)
			frameCounter.c += 1
		}
		wg.Done()
	}()
	go h264Demuxer(h264stream, frames)
	wg.Wait()
	logger.Printf("read %d frames\n", frameCounter.c)
}
func ByteStreamReader(connection net.Conn) {
	logger.Printf("opened bytestream\n")
	frameCounter := &counter{0}
	defer connection.Close()
	handleConnection(frameCounter, connection)
}
