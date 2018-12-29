package main

import (
	"fmt"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/codec/h264parser"
	// "github.com/nareix/joy4/format/ts"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

// InitialNALU indicates the start of a h264 packet
var InitialNALU = []byte{0, 0, 0, 1}
var logger *log.Logger

type counter struct{ c int }

func init() {
	logger = log.New(os.Stderr, "vlc-writer", log.Lshortfile)
}
func isNALU(packet []byte) bool {
	if len(packet) != len(InitialNALU) {
		return false
	}
	for i := range InitialNALU {
		if packet[i] != InitialNALU[i] {
			return false
		}
	}
	return true
}

// read bytes until a new header appears
func h264SegmentReader(r io.Reader) ([]byte, error) {
	packet := []byte{}
	for !isNALU(packet) {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		if err != nil {
			return packet, err
		}
		packet = append(packet, buf...)
	}
	// 0, 0, 0, 1
	return packet, nil
}

// read bytes between packets
func h264Demuxer(r io.Reader, frames chan []byte) {
	packet := []byte{}
	firstPacketRcvd := false
	for {
		// Read the very first NALU
		if !firstPacketRcvd {
			segment, err := h264SegmentReader(r)
			if err != nil {
				frames <- packet
				close(frames)
				return
			}
			packet = segment
			firstPacketRcvd = true
		}

		// Read the frame from the NALU boundary
		segment, err := h264SegmentReader(r)
		if err != nil {
			frames <- append(packet, segment...)
			close(frames)
			return
		}
		// Add the header for first packets
		// Remove the header from the first packet
		frames <- packet[0 : len(packet)-4]
		// S
		packet = addNALUHeader([]byte{})

	}
	// What about the next packet?
	// The next segment will be a packet up to
	// and including the next header.

}

func addNALUHeader(frame []byte) []byte {
	if isNALU(frame[0:4]) {
		return frame
	}
	return append(InitialNALU, frame...)
}

func loadFrame(frame []byte) error {
	if len(frame) >= 32 {
		logger.Printf("frame %v\n", frame[0:32])
	}
	codecData, err := h264parser.NewCodecDataFromSPSAndPPS(frame, frame)
	if err != nil {
		logger.Printf("codec error %s\n", err)
		return err
	} else {
		logger.Printf("frame (h, w) (%d, %d)\n", codecData.Height(), codecData.Width())
		logger.Printf("codec type %v\n", codecData.Type())
	}
	return nil
}

func decodeFrame(frame []byte) {
	codecType := av.H264
	_ = codecType
}

func handleConnection(frameCounter *counter, connection net.Conn) {
	defer connection.Close()
	var err error
	frameFile, _ := os.OpenFile("output.mp4", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer frameFile.Close()
	frames := make(chan []byte, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for frame := range frames {
			if len(frame) <= 8 {
				logger.Printf("frame %v\n", frame[0:len(frame)])
			} else {
				logger.Printf("frame %v\n", frame[0:8])
			}
			naluType := h264parser.CheckNALUsType(frame)
			logger.Printf("nalu type %d\n", naluType)
			if h264parser.IsDataNALU(frame) {
				logger.Println("is data nalu")
			}
			_, _ = frameFile.Write(frame)
			err = loadFrame(frame)
			frameCounter.c += 1
		}
		wg.Done()
	}()
	go h264Demuxer(connection, frames)
	wg.Wait()
	logger.Printf("read %d frames\n", frameCounter.c)
}

func main() {
	server, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(fmt.Sprintf("failed to listen %s\n", err))
	}
	frameCounter := counter{0}
	defer server.Close()
	for {
		connection, err := server.Accept()
		if err != nil {
			panic(fmt.Sprintf("connection failed %s\n", err))
		}
		go handleConnection(&frameCounter, connection)
		// hand connection to ReadMuxer
	}
}
