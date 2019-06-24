/*
NAME
  parse_test.go

DESCRIPTION
  parse_test.go provides testing for parsing utilities provided in parse.go

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2019 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  along with revid in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

package h264

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
)

// TestReadUe checks that readUe correctly parses an Exp-Golomb-coded element
// to a code number.
func TestReadUe(t *testing.T) {
	// tests has been derived from Table 9-2 in ITU-T H.H264, showing bit strings
	// and corresponding codeNums.
	tests := []struct {
		in     []byte // The bitstring we wish to read.
		expect uint   // The expected codeNum.
	}{
		{[]byte{0x80}, 0},  // Bit string: 1, codeNum: 0
		{[]byte{0x40}, 1},  // Bit string: 010, codeNum: 1
		{[]byte{0x60}, 2},  // Bit string: 011, codeNum: 2
		{[]byte{0x20}, 3},  // Bit string: 00100, codeNum: 3
		{[]byte{0x28}, 4},  // Bit string: 00101, codeNum: 4
		{[]byte{0x30}, 5},  // Bit string: 00110, codeNum: 5
		{[]byte{0x38}, 6},  // Bit string: 00111, codeNum: 6
		{[]byte{0x10}, 7},  // Bit string: 0001000, codeNum: 7
		{[]byte{0x12}, 8},  // Bit string: 0001001, codeNum: 8
		{[]byte{0x14}, 9},  // Bit string: 0001010, codeNum: 9
		{[]byte{0x16}, 10}, // Bit string: 0001011, codeNum: 10
	}

	for testn, test := range tests {
		got, err := readUe(bitio.NewReader(bytes.NewReader(test.in)))
		if err != nil {
			t.Fatalf("did not expect to get error: %v from readUe", err)
		}

		if test.expect != uint(got) {
			t.Errorf("did not get expected result for test: %v\nGot: %v\nWant: %v\n", testn, got, test.expect)
		}
	}
}
