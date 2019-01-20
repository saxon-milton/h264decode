package h264

import (
	"os"
	"testing"
)

func TestHandleConnection(t *testing.T) {
	input := "../sample.h264"
	f, err := os.Open(input)
	if err != nil {
		t.Fatalf("Error opening sample: %v\n", err)
	}
	frameCounter := &counter{0}
	handleConnection(frameCounter, f)
}
