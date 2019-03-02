package h264

import (
	"github.com/mrmod/degolomb"
	"io"
	"os"
)

type StreamReader struct {
	Reader        io.Reader
	byteOffset    int
	bitOffset     int
	endStreamChan chan int
	// Holds bits when a partial byte is read
	byteCache []byte
	bitCache  []int
	// Request N more bits
	moreBitsChan chan int
	// Return N bits
	BitStreamChan chan []int
	streamFile    *os.File
}

func NewStreamReader(r io.Reader, request, done chan int) *StreamReader {
	return &StreamReader{
		Reader:        r,
		moreBitsChan:  request,
		endStreamChan: done,
		BitStreamChan: make(chan []int, 1),
	}
}
func (s *StreamReader) LogStreamPosition() {
	logger.Printf("debug: stream position: %d bytes : %d bits\n", s.byteOffset, s.bitOffset)
}
func (s *StreamReader) ByteOffset() int {
	return s.byteOffset
}
func (s *StreamReader) BitOffset() int {
	return s.bitOffset
}
func (s *StreamReader) FillCache(buf []byte) {
	for _, b := range buf {

		s.bitCache = append(s.bitCache, degolomb.BitArray(b)...)
	}
	s.byteOffset += len(buf)
	logger.Printf("debug: filled bitCache with %d bytes, now %d bits available\n", len(buf), len(s.bitCache))
}

// DrainCache Empty the bit cache if it's large enough, otherwise,
// return the additional bits required
func (s *StreamReader) DrainCache(bitCount int) int {
	logger.Printf("debug: draining %d of %d bits\n", bitCount, len(s.bitCache))
	if bitCount < len(s.bitCache) {
		s.BitStreamChan <- s.bitCache[0:bitCount]
		s.bitCache = s.bitCache[bitCount:len(s.bitCache)]
		logger.Printf("debug: %d bits remain in cache\n", len(s.bitCache))
		return 0
	}
	s.BitStreamChan <- s.bitCache
	s.bitCache = []int{}
	logger.Printf("debug: bitCache completely drained\n")
	return bitCount - len(s.bitCache)
}
func (s *StreamReader) peekBits(cnt int) []int {
	if cnt >= len(s.bitCache) {
		if buf, err := s.GetBytes(1); err != nil {
			logger.Printf("error: while getting %d bytes: %v\n", cnt, err)
		} else {
			s.FillCache(buf)
			return s.peekBits(cnt)
		}
	}
	return s.bitCache[0:cnt]

}
func (s *StreamReader) GetBytes(cnt int) ([]byte, error) {
	if s.bitOffset != 0 {
		logger.Printf("warning: misaligned byte request\n")
		s.LogStreamPosition()
	}
	buf := make([]byte, cnt)
	_, err := s.Reader.Read(buf)
	if s.streamFile != nil {
		if _, err := s.streamFile.Write(buf); err != nil {
			logger.Printf("error: unable to write to streamFile: %v\n", err)
		}
	}
	if err != nil {
		logger.Printf("error: while getting %d bytes: %v\n", cnt, err)
	} else {
		s.byteOffset += 1
	}
	return buf, err
}

// Peek the next bits
func (s *StreamReader) NextBits(cnt int) []int {
	return s.peekBits(cnt)
}
func (s *StreamReader) NextBitsValue(cnt int) int {
	return bitVal(s.NextBits(cnt))
}

func (s *StreamReader) GetBits(cnt int) []int {
	s.moreBitsChan <- cnt
	return <-s.BitStreamChan
}
func (s *StreamReader) GetFieldValue(name string, width int) int {
	s.moreBitsChan <- width
	bitByte := []int{}
	value := 0
	logger.Printf("debug: reading %d bits for field %s\n", width, name)
	for bit := range s.BitStreamChan {
		logger.Printf("\tdebug: got %d bits for %s\n", len(bit), name)
		bitByte = append(bitByte, bit...)
		if len(bitByte) == width {
			break
		}
	}
	logger.Printf("debug: getting value of %d bits for %s\n", len(bitByte), name)
	if len(bitByte) < 8 {
		return bitVal(bitByte)
	}
	for i := range bitByte {
		if i > 0 && i%8 == 0 {
			value += bitVal(bitByte[i-8 : i])
		}
	}

	return value
}

func (s *StreamReader) Stream() {
	for request := range s.moreBitsChan {
		logger.Printf("debug: %d bits requested\n", request)
		// The number of bits beyond what the cache has
		// which we need to request
		bitCount := s.DrainCache(request)
		logger.Printf("debug: %d additional bits needed\n", bitCount)
		if bitCount == 0 {
			continue
		}

		byteCnt := bitCount / 8
		if bitCount%8 > 0 {
			byteCnt++
		}
		if byteCnt < 1 {
			byteCnt = 1
		}
		// Request the full bytes and cache the bits
		// which are not needed
		logger.Printf("debug: filling bit buffer with %d bytes\n", byteCnt)
		buf := make([]byte, byteCnt)
		_, err := s.Reader.Read(buf)
		if err != nil {
			logger.Printf("error: while reading %d bytes: %v\n", byteCnt, err)
			return
		}
		logger.Printf("debug: requesting to fill cache with %d bytes\n", len(buf))
		s.FillCache(buf)
		logger.Printf("debug: draining %d additional bits\n", bitCount)
		s.DrainCache(bitCount)
	}
}
