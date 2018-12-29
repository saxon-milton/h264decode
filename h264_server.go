package main

import (
	"fmt"
	// "github.com/nareix/joy4/av"
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
	logger = log.New(os.Stderr, "streamer ", log.Lshortfile)
}
func isNALU(packet []byte) bool {
	if len(packet) < len(InitialNALU) {
		return false
	}
	// The first NALU is a 4 byte packet, all others have the NALU
	// in the packet's tailing 4 bytes
	naluSegment := packet[len(packet)-4:]
	for i := range InitialNALU {
		if naluSegment[i] != InitialNALU[i] {
			/*
				if len(packet) <= 32 {
					logger.Printf("\t not NALU %v\n", packet)
				}
			*/
			return false
		}
	}
	/*
		if len(packet) <= 32 {
			logger.Printf("\t found NALU %v\n", packet)
		}
	*/
	return true
}

// read bytes until a new header appears
func h264SegmentReader(r io.Reader) ([]byte, error) {
	packet := []byte{}
	byteCounter := 0
	for !isNALU(packet) {
		buf := make([]byte, 1)
		n, err := r.Read(buf)
		byteCounter += n
		packet = append(packet, buf...)
		if err != nil {
			return packet, err
		}
	}
	logger.Printf("read %d byte h264 segment\n", len(packet))
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
	logger.Printf("read opening %d byte NALU boundary\n", len(packet))
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

func decodeFrame(frame []byte) error {
	codecData, err := h264parser.NewCodecDataFromSPSAndPPS(frame, frame)
	if err != nil {
		logger.Printf("codec error %s\n", err)
		return err
	}
	logger.Printf("frame (h, w) (%d, %d)\n", codecData.Height(), codecData.Width())
	logger.Printf("codec type %v\n", codecData.Type())
	return nil
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
			logger.Printf("[frame:%d] received %d byte frame\n", frameCounter.c, len(frame))
			max := 8
			if len(frame) <= 8 {
				max = len(frame)
			}
			naluType := h264parser.CheckNALUsType(frame)
			logger.Printf("[NT(%d) frame:%d.%d] %v\n",
				naluType,
				frameCounter.c,
				len(frame),
				frame[0:max])
			if h264parser.IsDataNALU(frame) {
				logger.Printf("[frame:%d] data frame\n", frameCounter.c)
			}
			_, _ = frameFile.Write(frame)
			err = decodeFrame(frame)
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
