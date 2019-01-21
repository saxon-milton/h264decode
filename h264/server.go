package h264

import (
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
// 0 - Forbidden 0 bit; always 0
// 1,2 - NRI
// 3,4,5,6,7 - Type
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

/*func isSPSFrame(frame []byte) (h264.SPSInfo, error) {
	return h264.ParseSPS(frame)
}*/

func decodeFrame(frame []byte) error {
	codecData, err := h264parser.NewCodecDataFromSPSAndPPS(frame, frame)
	if err != nil {
		logger.Printf("codec error %s\n", err)
		return err
	}
	logger.Printf("\t%v NALRefIDC: %d %s\n", frame[4], nalRefIDC(frame), NALRefIDC[nalRefIDC(frame)])
	logger.Printf("\t%v NALUnitType %d %s\n", frame[4], nalUnitType(frame), NALUnitType[nalUnitType(frame)])
	logger.Printf("\tframe (h, w) (%d, %d)\n", codecData.Height(), codecData.Width())
	logger.Printf("\tcodec type %v\n", codecData.Type())
	return nil
}

func handleConnection(frameCounter *counter, h264stream io.Reader) {
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
			rbsp := NewRBSP(frame)
			logger.Printf("[frame:%d] received %d byte frame\n", frameCounter.c, len(frame))
			naluType := h264parser.CheckNALUsType(frame[4:])
			logger.Printf("\t[NaluType(%d) frame:%d.%d]",
				naluType,
				frameCounter.c,
				len(frame))
			if nalUnitType(frame) == NALU_TYPE_SPS {
				sps := NewSPS(rbsp)
				logger.Printf("\tSPS: %#v\n", sps)
				logger.Printf("\tProfileIDC: %s\n", ProfileIDC[sps.Profile])
				logger.Printf("\tLevelIDC: %d\n", sps.Level)
				logger.Printf("\tSPS ID: %d\n", sps.ID)
				logger.Printf("\tChromaFormat %d\n", sps.ChromaFormat)
				logger.Printf("\tBitDepth Luma: %d\n", sps.BitDepthLuma)
				logger.Printf("\tBitDetch Chroma: %d\n", sps.BitDepthChroma)
				logger.Printf("%d---\n%#v\n\n---%d", frameCounter.c, rbsp, frameCounter.c)
			}
			if h264parser.IsDataNALU(frame[4:]) {
				logger.Printf("\t[frame:%d] data frame\n", frameCounter.c)
			}
			_, _ = frameFile.Write(frame)
			err = decodeFrame(frame)
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
