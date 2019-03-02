package h264

import (
	"github.com/mrmod/degolomb"
	"io"
)

type StreamReader struct {
	Reader        io.Reader
	byteOffset    int
	bitOffset     int
	endStreamChan chan int
	// Holds bits when a partial byte is read
	bitCache []int
	// Request N more bits
	moreBitsChan chan int
	// Return N bits
	BitStreamChan chan []int
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
func (s *StreamReader) FillCache(buf []byte) {
	for _, b := range buf {
		s.bitCache = append(s.bitCache, degolomb.BitArray(b)...)
	}
}

// DrainCache Empty the bit cache if it's large enough, otherwise,
// return the additional bits required
func (s *StreamReader) DrainCache(bitCount int) int {
	if bitCount < len(s.bitCache) {
		s.BitStreamChan <- s.bitCache[0:bitCount]
		s.bitCache = s.bitCache[bitCount:len(s.bitCache)]
		return 0
	}

	return bitCount - len(s.bitCache)
}
func (s *StreamReader) GetByte(cnt int) []byte {
	if s.bitOffset != 0 {
		logger.Printf("warning: misaligned byte request\n")
		s.LogStreamPosition()
	}
	buf := make([]byte, cnt)
	_, err := s.Reader.Read(buf)
	if err != nil {
		logger.Printf("error: while getting %d bytes: %v\n", cnt, err)
	}
	return buf
}

func (s *StreamReader) Stream() {
	for request := range s.moreBitsChan {
		logger.Printf("%d bits requested\n", request)
		// The number of bits beyond what the cache has
		// which we need to request
		bitCount := s.DrainCache(request)
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
		buf := make([]byte, byteCnt)
		_, err := s.Reader.Read(buf)
		if err != nil {
			logger.Printf("error: while reading %d bytes: %v\n", byteCnt, err)
			return
		}
		s.FillCache(buf)
		s.DrainCache(bitCount)
	}
}
