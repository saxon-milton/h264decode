package h264

import "github.com/mrmod/degolomb"

type SPS struct {
	// 8 bits
	ProfileIDC int
	// 6 bits
	Constraint0, Constraint1 int
	Constraint2, Constraint3 int
	Constraint4, Constraint5 int
	// 2 bit reserved 0 bits
	// 8 bits
	LevelIDC int
	// ExpGolomb variable bits
	ID                    int
	ChromaFormatIDC       int
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

func NewSPS(rbsp []byte) SPS {
	return SPS{
		ProfileIDC:  int(rbsp[0]),
		Constraint0: degolomb.BitArray(rbsp[1])[0],
		Constraint1: degolomb.BitArray(rbsp[1])[1],
		Constraint2: degolomb.BitArray(rbsp[1])[2],
		Constraint3: degolomb.BitArray(rbsp[1])[3],
		Constraint4: degolomb.BitArray(rbsp[1])[4],
		Constraint5: degolomb.BitArray(rbsp[1])[5],
		LevelIDC:    int(rbsp[2]),
	}
}
