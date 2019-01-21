package h264

import "github.com/mrmod/degolomb"

type SPS struct {
	// 8 bits
	Profile int
	// 6 bits
	Constraint0, Constraint1 int
	Constraint2, Constraint3 int
	Constraint4, Constraint5 int
	// 2 bit reserved 0 bits
	// 8 bits
	Level int
	// ExpGolomb variable bits
	ID                    int
	ChromaFormat          int
	UseSeparateColorPlane bool
	// minus8
	BitDepthLuma, BitDepthChroma int
	QPrimeYZeroTransformBypass   bool
	SeqScalingMatrixPresent      bool
	// log2(MaxFrameNum) - 4
	MaxFrame          int
	PicOrderCountType int
	MaxNumRefFrames   int
}

func spsId(third byte) int {
	return degolomb.Degolomb(third, []int{5})[0]
}
func chromaFormat(third byte) int {
	return degolomb.Degolomb(third, []int{5, 3})[1]
}
func useSeparateColorPlane(fourth byte) bool {
	return degolomb.BitArray(fourth)[0] == 1
}

func NewSPS(rbsp []byte) SPS {
	return SPS{
		Profile:               int(rbsp[0]),
		Constraint0:           degolomb.BitArray(rbsp[1])[0],
		Constraint1:           degolomb.BitArray(rbsp[1])[1],
		Constraint2:           degolomb.BitArray(rbsp[1])[2],
		Constraint3:           degolomb.BitArray(rbsp[1])[3],
		Constraint4:           degolomb.BitArray(rbsp[1])[4],
		Constraint5:           degolomb.BitArray(rbsp[1])[5],
		Level:                 int(rbsp[2]),
		ID:                    spsId(rbsp[3]),
		ChromaFormat:          chromaFormat(rbsp[3]),
		UseSeparateColorPlane: useSeparateColorPlane(rbsp[4]),
	}
}
