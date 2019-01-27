package h264

// import "github.com/mrmod/degolomb"
import "fmt"
import "math"
import "strings"

// Specification Page 43 7.3.2.1.1
// Range is always inclusive
// XRange is always exclusive
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
	// Range 0 - 31 ; 6 bits
	ID                    int
	ChromaFormat          int
	UseSeparateColorPlane bool
	// minus8
	BitDepthLuma, BitDepthChroma int
	QPrimeYZeroTransformBypass   bool
	SeqScalingMatrixPresent      bool
	// Delta is (0-12)-1 ; 4 bits
	SeqScalingList []bool // se
	// Range 0 - 12; 4 bits
	Log2MaxFrameNumMinus4 int
	// Range 0 - 2; 2 bits
	PicOrderCountType int
	// Range 0 - 12; 4 bits
	Log2MaxPicOrderCntLSBMin4 int
	DeltaPicOrderAlwaysZero   bool
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForNonRefPic int // Value - 1 (se)
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForTopToBottomField int // Value - 1 (se)
	// Range 0 - 255 ; 8 bits
	NumRefFramesInPicOrderCntCycle int
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForRefFrameList []int // Value - 1 ([]se)
	// Range 0 - MaxDpbFrames
	MaxNumRefFrames            int
	GapsInFrameNumValueAllowed bool
	// Page 77
	PicWidthInMbsMinus1 int
	// Page 77
	PicHeightInMapUnitsMinus1          int
	FrameMbsOnly                       bool
	MBAdaptiveFrameField               bool
	Direct8x8Inference                 bool
	FrameCropping                      bool
	FrameCropLeftOffset                int
	FrameCropRightOffset               int
	FrameCropTopOffset                 int
	FrameCropBottomOffset              int
	VuiParametersPresent               bool
	VuiParameters                      []int
	AspectRatioInfoPresent             bool
	AspectRatio                        int
	SarWidth                           int
	SarHeight                          int
	OverscanInfoPresent                bool
	OverscanAppropriate                bool
	VideoSignalTypePresent             bool
	VideoFormat                        int
	VideoFullRange                     bool
	ColorDescriptionPresent            bool
	ColorPrimaries                     int
	TransferCharacteristics            int
	MatrixCoefficients                 int
	ChromaLocInfoPresent               bool
	ChromaSampleLocTypeTopField        int
	ChromaSampleLocTypeBottomField     int
	CpbCntMinus1                       int
	BitRateScale                       int
	CpbSizeScale                       int
	BitRateValueMinus1                 []int
	Cbr                                []bool
	InitialCpbRemovalDelayLengthMinus1 int
	CpbRemovalDelayLengthMinus1        int
	CpbSizeValueMinus1                 []int
	DpbOutputDelayLengthMinus1         int
	TimeOffsetLength                   int
	TimingInfoPresent                  bool
	NumUnitsInTick                     int
	TimeScale                          int
	NalHrdParametersPresent            bool
	FixedFrameRate                     bool
	VclHrdParametersPresent            bool
	LowHrdDelay                        bool
	PicStructPresent                   bool
	BitstreamRestriction               bool
	MotionVectorsOverPicBoundaries     bool
	MaxBytesPerPicDenom                int
	MaxBitsPerMbDenom                  int
	Log2MaxMvLengthHorizontal          int
	Log2MaxMvLengthVertical            int
	MaxDecFrameBuffering               int
	MaxNumReorderFrames                int
}

var (
	DefaultScalingMatrix4x4 = [][]int{
		[]int{6, 13, 20, 28, 13, 20, 28, 32, 20, 28, 32, 37, 28, 32, 37, 42},
		[]int{10, 14, 20, 24, 14, 20, 24, 27, 20, 24, 27, 30, 24, 27, 30, 34},
	}

	DefaultScalingMatrix8x8 = [][]int{
		[]int{6, 10, 13, 16, 18, 23, 25, 27,
			10, 11, 16, 18, 23, 25, 27, 29,
			13, 16, 18, 23, 25, 27, 29, 31,
			16, 18, 23, 25, 27, 29, 31, 33,
			18, 23, 25, 27, 29, 31, 33, 36,
			23, 25, 27, 29, 31, 33, 36, 38,
			25, 27, 29, 31, 33, 36, 38, 40,
			27, 29, 31, 33, 36, 38, 40, 42},
		[]int{9, 13, 15, 17, 19, 21, 22, 24,
			13, 13, 17, 19, 21, 22, 24, 25,
			15, 17, 19, 21, 22, 24, 25, 27,
			17, 19, 21, 22, 24, 25, 27, 28,
			19, 21, 22, 24, 25, 27, 28, 30,
			21, 22, 24, 25, 27, 28, 30, 32,
			22, 24, 25, 27, 28, 30, 32, 33,
			24, 25, 27, 28, 30, 32, 33, 35},
	}
	Default4x4IntraList = []int{6, 13, 13, 20, 20, 20, 38, 38, 38, 38, 32, 32, 32, 37, 37, 42}
	Default4x4InterList = []int{10, 14, 14, 20, 20, 20, 24, 24, 24, 24, 27, 27, 27, 30, 30, 34}
	Default8x8IntraList = []int{
		6, 10, 10, 13, 11, 13, 16, 16, 16, 16, 18, 18, 18, 18, 18, 23,
		23, 23, 23, 23, 23, 25, 25, 25, 25, 25, 25, 25, 27, 27, 27, 27,
		27, 27, 27, 27, 29, 29, 29, 29, 29, 29, 29, 31, 31, 31, 31, 31,
		31, 33, 33, 33, 33, 33, 36, 36, 36, 36, 38, 38, 38, 40, 40, 42}
	Default8x8InterList = []int{
		9, 13, 13, 15, 13, 15, 17, 17, 17, 17, 19, 19, 19, 19, 19, 21,
		21, 21, 21, 21, 21, 22, 22, 22, 22, 22, 22, 22, 24, 24, 24, 24,
		24, 24, 24, 24, 25, 25, 25, 25, 25, 25, 25, 27, 27, 27, 27, 27,
		27, 28, 28, 28, 28, 28, 30, 30, 30, 30, 32, 32, 32, 33, 33, 35}
	ScalingList4x4 = map[int][]int{
		0:  Default4x4IntraList,
		1:  Default4x4IntraList,
		2:  Default4x4IntraList,
		3:  Default4x4InterList,
		4:  Default4x4InterList,
		5:  Default4x4InterList,
		6:  Default8x8IntraList,
		7:  Default8x8InterList,
		8:  Default8x8IntraList,
		9:  Default8x8InterList,
		10: Default8x8IntraList,
		11: Default8x8InterList,
	}
	ScalingList8x8 = ScalingList4x4
)

func bitVal(bits []int) int {
	t := 0
	for i, b := range bits {
		if b == 1 {
			t += 1 << uint((len(bits)-1)-i)
		}
	}
	fmt.Printf("\t bitVal: %d\n", t)
	return t
}

// 9.1 Table 9-2
func ue(bits []int) int {
	return bitVal(bits) - 1
}

// 9.1.1 Table 9-3
func se(bits []int) int {
	codeNum := bitVal(bits) - 1
	return int(math.Pow(float64(-1), float64(codeNum+1)) * math.Ceil(float64(codeNum/2)))
}

func isInList(l []int, term int) bool {
	for _, m := range l {
		if m == term {
			return true
		}
	}
	return false
}

func scalingList(b *BitReader, rbsp []byte, scalingList []int, sizeOfScalingList int, defaultScalingMatrix []int) {
	lastScale := 8
	nextScale := 8
	for i := 0; i < sizeOfScalingList; i++ {
		if nextScale != 0 {
			deltaScale := se(b.golomb(rbsp))
			nextScale = (lastScale + deltaScale + 256) % 256
			if i == 0 && nextScale == 0 {
				// Scaling list should use the default list for this point in the matrix
				_ = defaultScalingMatrix
			}
		}
		if nextScale == 0 {
			scalingList[i] = lastScale
		} else {
			scalingList[i] = nextScale
		}
		lastScale = scalingList[i]
	}
}
func NewSPS(rbsp []byte) SPS {
	fmt.Printf(" == SPS RBSP %d bytes %d bits == \n", len(rbsp), len(rbsp)*8)
	sps := SPS{}
	b := &BitReader{}
	// Byte 0
	nextField := func(name string, bits int) int {
		fmt.Printf("\tget %d bits for %s\n", bits, name)
		buf := make([]int, bits)
		_, err := b.Read(rbsp, buf)
		if err != nil {
			fmt.Printf("error reading bits for %s: %v\n", name, err)
			return -1
		}

		return bitVal(buf)
	}
	hrdParameters := func() {
		sps.CpbCntMinus1 = ue(b.golomb(rbsp))
		sps.BitRateScale = nextField("BitRateScale", 4)
		sps.CpbSizeScale = nextField("CPBSizeScale", 4)
		// SchedSelIdx E1.2
		for sseli := 0; sseli <= sps.CpbCntMinus1; sseli++ {
			sps.BitRateValueMinus1 = append(sps.BitRateValueMinus1, ue(b.golomb(rbsp)))
			sps.CpbSizeValueMinus1 = append(sps.CpbSizeValueMinus1, ue(b.golomb(rbsp)))
			if v := nextField(fmt.Sprintf("CBR[%d]", sseli), 1); v == 1 {
				sps.Cbr = append(sps.Cbr, true)
			} else {
				sps.Cbr = append(sps.Cbr, false)
			}

			sps.InitialCpbRemovalDelayLengthMinus1 = nextField("InitialCpbRemovalDelayLengthMinus1", 5)
			sps.CpbRemovalDelayLengthMinus1 = nextField("CpbRemovalDelayLengthMinus1", 5)
			sps.DpbOutputDelayLengthMinus1 = nextField("DpbOutputDelayLengthMinus1", 5)
			sps.TimeOffsetLength = nextField("TimeOffsetLength", 5)
		}
	}
	sps.Profile = nextField("ProfileIDC", 8)
	sps.Constraint0 = nextField("Constraint0", 1)
	sps.Constraint1 = nextField("Constraint1", 1)
	sps.Constraint2 = nextField("Constraint2", 1)
	sps.Constraint3 = nextField("Constraint3", 1)
	sps.Constraint4 = nextField("Constraint4", 1)
	sps.Constraint5 = nextField("Constraint5", 1)
	_ = nextField("ReservedZeroBits", 2)
	sps.Level = nextField("LevelIDC", 8)
	// sps.ID = nextField("SPSID", 6) // proper
	fmt.Printf(" -- Pre SPSID %+v\n", b)
	sps.ID = ue(b.golomb(rbsp))
	fmt.Printf(" -- SPSID %+v\n", b)
	sps.ChromaFormat = ue(b.golomb(rbsp))
	// This should be done only for certain ProfileIDC:
	isProfileIDC := []int{100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135}
	// SpecialProfileCase1
	if isInList(isProfileIDC, sps.Profile) {
		if sps.ChromaFormat == 3 {
			if v := nextField("SeperateColorPlaneFlag", 1); v == 1 {
				sps.UseSeparateColorPlane = true
			} else {
				sps.UseSeparateColorPlane = false
			}
		}

		sps.BitDepthLuma = ue(b.golomb(rbsp))
		sps.BitDepthChroma = ue(b.golomb(rbsp))
		if v := nextField("QPrimeYZeroTransformBypassFlag", 1); v == 1 {
			sps.QPrimeYZeroTransformBypass = true
		} else {
			sps.QPrimeYZeroTransformBypass = false
		}
		if v := nextField("SequenceScalingMatrixPresentFlag", 1); v == 1 {
			sps.SeqScalingMatrixPresent = true
		} else {
			sps.SeqScalingMatrixPresent = false
		}
		if sps.SeqScalingMatrixPresent {
			max := 12
			if sps.ChromaFormat != 3 {
				max = 8
			}
			fmt.Printf("building Scaling matrix for %d elements\n", max)
			for i := 0; i < max; i++ {
				if v := nextField(fmt.Sprintf("SeqScalingListPresentFlag[%d]", i), 1); v == 1 {
					sps.SeqScalingList = append(sps.SeqScalingList, true)
				} else {
					sps.SeqScalingList = append(sps.SeqScalingList, false)
				}
				if sps.SeqScalingList[i] {
					if i < 6 {
						scalingList(
							b, rbsp,
							ScalingList4x4[i],
							16,
							DefaultScalingMatrix4x4[i])
						// 4x4: Page 75 bottom
					} else {
						// 8x8 Page 76 top
						scalingList(
							b, rbsp,
							ScalingList8x8[i],
							64,
							DefaultScalingMatrix8x8[i-6])
					}
				}
			}
		}
	} // End SpecialProfileCase1
	showSPS := func() {

		fmt.Printf("=== SPS RBSP ===\n")
		for _, line := range strings.Split(fmt.Sprintf("%+v", sps), " ") {
			fmt.Println(line)
		}
	}
	// showSPS()
	// return sps
	// Possibly wrong due to no scaling list being built
	sps.Log2MaxFrameNumMinus4 = ue(b.golomb(rbsp))
	sps.PicOrderCountType = ue(b.golomb(rbsp))
	if sps.PicOrderCountType == 0 {
		sps.Log2MaxPicOrderCntLSBMin4 = ue(b.golomb(rbsp))
	} else if sps.PicOrderCountType == 1 {
		if v := nextField("DeltaPicOrderAlwaysZeroFlag", 1); v == 1 {
			sps.DeltaPicOrderAlwaysZero = true
		} else {
			sps.DeltaPicOrderAlwaysZero = false
		}
		sps.OffsetForNonRefPic = se(b.golomb(rbsp))
		sps.OffsetForTopToBottomField = se(b.golomb(rbsp))
		sps.NumRefFramesInPicOrderCntCycle = ue(b.golomb(rbsp))

		for i := 0; i < sps.NumRefFramesInPicOrderCntCycle; i++ {
			sps.OffsetForRefFrameList = append(
				sps.OffsetForRefFrameList,
				se(b.golomb(rbsp)))
		}

	}
	sps.MaxNumRefFrames = ue(b.golomb(rbsp))
	if v := nextField("GapsInFrameNumValueAllowedFlag", 1); v == 1 {
		sps.GapsInFrameNumValueAllowed = true
	}
	sps.PicWidthInMbsMinus1 = ue(b.golomb(rbsp))
	sps.PicHeightInMapUnitsMinus1 = ue(b.golomb(rbsp))
	if v := nextField("FrameMbsOnlyFlag", 1); v == 1 {
		sps.FrameMbsOnly = true
	}
	if !sps.FrameMbsOnly {
		if v := nextField("MBAdaptiveFrameFieldFlag", 1); v == 1 {
			sps.MBAdaptiveFrameField = true
		}
	}
	if v := nextField("Direct8x8InferenceFlag", 1); v == 1 {
		sps.Direct8x8Inference = true
	}
	if v := nextField("FrameCroppingFlag", 1); v == 1 {
		sps.FrameCropping = true
	}
	if sps.FrameCropping {
		sps.FrameCropLeftOffset = ue(b.golomb(rbsp))
		sps.FrameCropRightOffset = ue(b.golomb(rbsp))
		sps.FrameCropTopOffset = ue(b.golomb(rbsp))
		sps.FrameCropBottomOffset = ue(b.golomb(rbsp))
	}
	if v := nextField("VUIParametersPresentFlag", 1); v == 1 {
		sps.VuiParametersPresent = true
	}
	if sps.VuiParametersPresent {
		// vui_parameters
		if v := nextField("AspectRatioInfoPresentFlag", 1); v == 1 {
			sps.AspectRatioInfoPresent = true
		}
		if sps.AspectRatioInfoPresent {
			sps.AspectRatio = nextField("AspectRatioIDC", 8)
			EXTENDED_SAR := 999
			if sps.AspectRatio == EXTENDED_SAR {
				sps.SarWidth = nextField("SARWidth", 16)
				sps.SarHeight = nextField("SARHeight", 16)
			}
		}
		if v := nextField("OverscanInfoPresentFlag", 1); v == 1 {
			sps.OverscanInfoPresent = true
		}
		if sps.OverscanInfoPresent {
			if v := nextField("OverscanAppropriateFlag", 1); v == 1 {
				sps.OverscanAppropriate = true
			}
		}
		if v := nextField("VideoSignalPresentFlag", 1); v == 1 {
			sps.VideoSignalTypePresent = true
		}
		if sps.VideoSignalTypePresent {
			sps.VideoFormat = nextField("VideoFormat", 3)
		}
		if sps.VideoSignalTypePresent {
			if v := nextField("VideoFullRangeFlag", 1); v == 1 {
				sps.VideoFullRange = true
			}
			if v := nextField("ColorDescriptionPresentFlag", 1); v == 1 {
				sps.ColorDescriptionPresent = true
			}
			if sps.ColorDescriptionPresent {
				sps.ColorPrimaries = nextField("ColorPrimaries", 8)
				sps.TransferCharacteristics = nextField("TransferCharacteristics", 8)
				sps.MatrixCoefficients = nextField("MatrixCoefficients", 8)
			}
		}
		if v := nextField("ChromaLocInfoPresentFlag", 1); v == 1 {
			sps.ChromaLocInfoPresent = true
		}
		if sps.ChromaLocInfoPresent {
			sps.ChromaSampleLocTypeTopField = ue(b.golomb(rbsp))
			sps.ChromaSampleLocTypeBottomField = ue(b.golomb(rbsp))
		}

		if v := nextField("TimingInfoPresentFlag", 1); v == 1 {
			sps.TimingInfoPresent = true
		}
		if sps.TimingInfoPresent {
			sps.NumUnitsInTick = nextField("NumUnitsInTick", 32)
			sps.TimeScale = nextField("TimeScale", 32)
			if v := nextField("FixedFramerateFlag", 1); v == 1 {
				sps.FixedFrameRate = true
			}
		}
		if v := nextField("NALHRDParametersPresent", 1); v == 1 {
			sps.NalHrdParametersPresent = true
		}
		if sps.NalHrdParametersPresent {
			hrdParameters()
		}
		if v := nextField("VCLHRDParametersPresent", 1); v == 1 {
			sps.VclHrdParametersPresent = true
		}
		if sps.VclHrdParametersPresent {
			hrdParameters()
		}
		if sps.NalHrdParametersPresent || sps.VclHrdParametersPresent {
			if v := nextField("LowHRDDelayFlag", 1); v == 1 {
				sps.LowHrdDelay = true
			}
		}
		if v := nextField("PicStructPresentFlag", 1); v == 1 {
			sps.PicStructPresent = true
		}
		if v := nextField("BitstreamRestrictionFlag", 1); v == 1 {
			sps.BitstreamRestriction = true
		}
		if sps.BitstreamRestriction {
			if v := nextField("MotionVectorsOverPicBoundaries", 1); v == 1 {
				sps.MotionVectorsOverPicBoundaries = true
			}
			sps.MaxBytesPerPicDenom = ue(b.golomb(rbsp))
			sps.MaxBitsPerMbDenom = ue(b.golomb(rbsp))
			sps.Log2MaxMvLengthHorizontal = ue(b.golomb(rbsp))
			sps.Log2MaxMvLengthVertical = ue(b.golomb(rbsp))
			sps.MaxNumReorderFrames = ue(b.golomb(rbsp))
			sps.MaxDecFrameBuffering = ue(b.golomb(rbsp))
		}

	} // End VuiParameters Annex E.1.1

	showSPS()
	return sps
}
