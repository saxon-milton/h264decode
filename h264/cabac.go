package h264

const (
	NaCtxId            = 10000
	NA_SUFFIX          = -1
	MbAddrNotAvailable = 10000
)

// G.7.4.3.4 via G.7.3.3.4 via 7.3.2.13 for NalUnitType 20 or 21
// refLayerMbWidthC is equal to MbWidthC for the reference layer representation
func RefMbW(chromaFlag, refLayerMbWidthC int) int {
	if chromaFlag == 0 {
		return 16
	}
	return refLayerMbWidthC
}

// refLayerMbHeightC is equal to MbHeightC for the reference layer representation
func RefMbH(chromaFlag, refLayerMbHeightC int) int {
	if chromaFlag == 0 {
		return 16
	}
	return refLayerMbHeightC
}
func XOffset(xRefMin16, refMbW int) int {
	return (((xRefMin16 - 64) >> 8) << 4) - (refMbW >> 1)
}
func YOffset(yRefMin16, refMbH int) int {
	return (((yRefMin16 - 64) >> 8) << 4) - (refMbH >> 1)
}
func MbWidthC(sps *SPS) int {
	mbWidthC := 16 / SubWidthC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbWidthC = 0
	}
	return mbWidthC
}
func MbHeightC(sps *SPS) int {
	mbHeightC := 16 / SubHeightC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbHeightC = 0
	}
	return mbHeightC
}

// G.8.6.2.2.2
func Xr(x, xOffset, refMbW int) int {
	return (x + xOffset) % refMbW
}
func Yr(y, yOffset, refMbH int) int {
	return (y + yOffset) % refMbH
}

// G.8.6.2.2.2
func Xd(xr, refMbW int) int {
	if xr >= refMbW/2 {
		return xr - refMbW
	}
	return xr + 1
}
func Yd(yr, refMbH int) int {
	if yr >= refMbH/2 {
		return yr - refMbH
	}
	return yr + 1
}
func Ya(yd, refMbH, signYd int) int {
	return yd - (refMbH/2+1)*signYd
}

// 6.4.11.1
func MbAddr(xd, yd, predPartWidth int) {
	// TODO: Unfinished
	var n string
	if xd == -1 && yd == 0 {
		n = "A"
	}
	if xd == 0 && yd == -1 {
		n = "B"
	}
	if xd == predPartWidth && yd == -1 {
		n = "C"
	}
	if xd == -1 && yd == -1 {
		n = "D"
	}
	_ = n
}

func CondTermFlag(mbAddr, mbSkipFlag int) int {
	if mbAddr == MbAddrNotAvailable || mbSkipFlag == 1 {
		return 0
	}
	return 1
}

// s9.3.3 p 278: Returns the value of the syntax element
func (bin *Binarization) Decode(sliceContext *SliceContext, b *BitReader, rbsp []byte) {
	if bin.SyntaxElement == "MbType" {
		bin.binString = binIdxMbMap[sliceContext.Slice.Data.SliceTypeName][sliceContext.Slice.Data.MbType]
	} else {
		logger.Printf("TODO: no means to find binString for %s\n", bin.SyntaxElement)
	}
}

// 9.3.3.1.1 : returns ctxIdxInc
func Decoder9_3_3_1_1_1(condTermFlagA, condTermFlagB int) int {
	return condTermFlagA + condTermFlagB
}

// 7-30 p 112
func SliceQPy(pps *PPS, header *SliceHeader) int {
	return 26 + pps.PicInitQpMinus26 + header.SliceQpDelta
}

type CABAC struct {
}

// table 9-1
func initCabac(pps *PPS, header *SliceHeader, data *SliceData) {
	pStateIdx := SliceQPy(pps, header)
	valMPS := SliceQPy(pps, header)
	_ = pStateIdx
	_ = valMPS
	// Initialization of context variables
	// Initialization of decoding engine
}

// Table 9-36, 9-37
// func BinIdx(mbType int, sliceTypeName string) []int {
// Map of SliceTypeName[MbType][]int{binString}
// {"SliceTypeName": {MbTypeCode: []BinString}}
var (
	binIdxMbMap = map[string]map[int][]int{
		"I": map[int][]int{
			0:  []int{0},
			1:  []int{1, 0, 0, 0, 0, 0},
			2:  []int{1, 0, 0, 0, 0, 1},
			3:  []int{1, 0, 0, 0, 1, 0},
			4:  []int{1, 0, 0, 0, 1, 1},
			5:  []int{1, 0, 0, 1, 0, 0, 0},
			6:  []int{1, 0, 0, 1, 0, 0, 1},
			7:  []int{1, 0, 0, 1, 0, 1, 0},
			8:  []int{1, 0, 0, 1, 0, 1, 1},
			9:  []int{1, 0, 0, 1, 1, 0, 0},
			10: []int{1, 0, 0, 1, 1, 0, 1},
			11: []int{1, 0, 0, 1, 1, 1, 0},
			12: []int{1, 0, 0, 1, 1, 1, 1},
			13: []int{1, 0, 1, 0, 0, 0},
			14: []int{1, 0, 1, 0, 0, 1},
			15: []int{1, 0, 1, 0, 1, 0},
			16: []int{1, 0, 1, 0, 1, 1},
			17: []int{1, 0, 1, 1, 0, 0, 0},
			18: []int{1, 0, 1, 1, 0, 0, 1},
			19: []int{1, 0, 1, 1, 0, 1, 0},
			20: []int{1, 0, 1, 1, 0, 1, 1},
			21: []int{1, 0, 1, 1, 1, 0, 0},
			22: []int{1, 0, 1, 1, 1, 0, 1},
			23: []int{1, 0, 1, 1, 1, 1, 0},
			24: []int{1, 0, 1, 1, 1, 1, 1},
			25: []int{1, 1},
		},
		// Table 9-37
		"P": map[int][]int{
			0:  []int{0, 0, 0},
			1:  []int{0, 1, 1},
			2:  []int{0, 1, 0},
			3:  []int{0, 0, 1},
			4:  []int{},
			5:  []int{1},
			6:  []int{1},
			7:  []int{1},
			8:  []int{1},
			9:  []int{1},
			10: []int{1},
			11: []int{1},
			12: []int{1},
			13: []int{1},
			14: []int{1},
			15: []int{1},
			16: []int{1},
			17: []int{1},
			18: []int{1},
			19: []int{1},
			20: []int{1},
			21: []int{1},
			22: []int{1},
			23: []int{1},
			24: []int{1},
			25: []int{1},
			26: []int{1},
			27: []int{1},
			28: []int{1},
			29: []int{1},
			30: []int{1},
		},
		// Table 9-37
		"SP": map[int][]int{
			0:  []int{0, 0, 0},
			1:  []int{0, 1, 1},
			2:  []int{0, 1, 0},
			3:  []int{0, 0, 1},
			4:  []int{},
			5:  []int{1},
			6:  []int{1},
			7:  []int{1},
			8:  []int{1},
			9:  []int{1},
			10: []int{1},
			11: []int{1},
			12: []int{1},
			13: []int{1},
			14: []int{1},
			15: []int{1},
			16: []int{1},
			17: []int{1},
			18: []int{1},
			19: []int{1},
			20: []int{1},
			21: []int{1},
			22: []int{1},
			23: []int{1},
			24: []int{1},
			25: []int{1},
			26: []int{1},
			27: []int{1},
			28: []int{1},
			29: []int{1},
			30: []int{1},
		},
		// TODO: B Slice table 9-37
	}

	// Map of SliceTypeName[SubMbType][]int{binString}
	binIdxSubMbMap = map[string]map[int][]int{
		"P": map[int][]int{
			0: []int{1},
			1: []int{0, 0},
			2: []int{0, 1, 1},
			3: []int{0, 1, 0},
		},
		"SP": map[int][]int{
			0: []int{1},
			1: []int{0, 0},
			2: []int{0, 1, 1},
			3: []int{0, 1, 0},
		},
		// TODO: B slice table 9-38
	}

	// Table 9-36, 9-37
	MbBinIdx = []int{1, 2, 3, 4, 5, 6}

	// Table 9-38
	SubMbBinIdx = []int{0, 1, 2, 3, 4, 5}
)

// Table 9-34
type MaxBinIdxCtx struct {
	// When false, Prefix is the MaxBinIdxCtx
	IsPrefixSuffix bool
	Prefix, Suffix int
}
type CtxIdxOffset struct {
	// When false, Prefix is the MaxBinIdxCtx
	IsPrefixSuffix bool
	Prefix, Suffix int
}

// Table 9-34
type Binarization struct {
	SyntaxElement string
	BinarizationType
	MaxBinIdxCtx
	CtxIdxOffset
	UseDecodeBypass int
	binIdx          int
	binString       []int
}
type BinarizationType struct {
	PrefixSuffix   bool
	FixedLength    bool
	Unary          bool
	TruncatedUnary bool
	CMax           bool
	// 9.3.2.3
	UEGk      bool
	CMaxValue int
}

func NewBinarization(syntaxElement string, data *SliceData) *Binarization {
	sliceTypeName := data.SliceTypeName
	logger.Printf("NewBinarization for %s in sliceType %s\n", syntaxElement, sliceTypeName)
	binarization := &Binarization{SyntaxElement: syntaxElement}
	switch syntaxElement {
	case "CodedBlockPattern":
		binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 3, Suffix: 1}
		binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 73, Suffix: 77}
	case "IntraChromaPredMode":
		binarization.BinarizationType = BinarizationType{
			TruncatedUnary: true, CMax: true, CMaxValue: 3}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 1}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 64}
	case "MbQpDelta":
		binarization.BinarizationType = BinarizationType{}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 2}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 60}
	case "MvdLnEnd0":
		binarization.UseDecodeBypass = 1
		binarization.BinarizationType = BinarizationType{UEGk: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 4, Suffix: NA_SUFFIX}
		binarization.CtxIdxOffset = CtxIdxOffset{
			IsPrefixSuffix: true,
			Prefix:         40,
			Suffix:         NA_SUFFIX,
		}
	case "MvdLnEnd1":
		binarization.UseDecodeBypass = 1
		binarization.BinarizationType = BinarizationType{UEGk: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{
			IsPrefixSuffix: true,
			Prefix:         4,
			Suffix:         NA_SUFFIX,
		}
		binarization.CtxIdxOffset = CtxIdxOffset{
			IsPrefixSuffix: true,
			Prefix:         47,
			Suffix:         NA_SUFFIX,
		}
	case "MbType":
		logger.Printf("\tMbType is %s\n", data.MbTypeName)
		switch sliceTypeName {
		case "SI":
			binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 0, Suffix: 6}
			binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 0, Suffix: 3}
		case "I":
			binarization.BinarizationType = BinarizationType{}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 6}
			binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 3}
		case "SP":
			fallthrough
		case "P":
			binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 2, Suffix: 5}
			binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 14, Suffix: 17}
		}
	case "MbFieldDecodingFlag":
		binarization.BinarizationType = BinarizationType{
			FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 70}
	case "PrevIntra4x4PredModeFlag":
		fallthrough
	case "PrevIntra8x8PredModeFlag":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 68}
	case "RefIdxL0":
		fallthrough
	case "RefIdxL1":
		binarization.BinarizationType = BinarizationType{Unary: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 2}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 54}
	case "RemIntra4x4PredMode":
		fallthrough
	case "RemIntra8x8PredMode":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 7}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 69}
	case "TransformSize8x8Flag":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 399}
	}
	return binarization
}

func CtxIdx(binIdx, maxBinIdxCtx, ctxIdxOffset int) int {
	ctxIdx := NaCtxId
	// table 9-39
	switch ctxIdxOffset {
	case 0:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.3
	case 3:
		switch binIdx {
		case 0:
			// 9.3.3.1.1.3
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 3
		case 3:
			ctxIdx = 4
		case 4:
			// 9.3.3.1.2
		case 5:
			// 9.3.3.1.2
		default:
			ctxIdx = 7
		}
	case 11:
		if binIdx != 0 {
			return NaCtxId
		}

		// 9.3.3.1.1.3
	case 14:
		if binIdx == 0 {
			ctxIdx = 0
		}
		if binIdx == 1 {
			ctxIdx = 1
		}
		if binIdx == 2 {
			// 9.3.3.1.2
		}
		if binIdx > 2 {
			return NaCtxId
		}
	case 17:
		switch binIdx {
		case 0:
			ctxIdx = 0
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 1
		case 3:
			ctxIdx = 2
		case 4:
			// 9.3.3.1.2
		default:
			ctxIdx = 3
		}

	case 21:
		if binIdx < 3 {
			ctxIdx = binIdx
		} else {
			return NaCtxId
		}
	case 24:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.1
	case 27:
		switch binIdx {
		case 0:
			// 9.3.3.1.1.3
		case 1:
			ctxIdx = 3
		case 2:
			// 9.3.3.1.2
		default:
			ctxIdx = 5
		}
	case 32:
		switch binIdx {
		case 0:
			ctxIdx = 0
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 1
		case 3:
			ctxIdx = 2
		case 4:
			// 9.3.3.1.2
		default:
			ctxIdx = 3
		}
	case 36:
		if binIdx == 0 || binIdx == 1 {
			ctxIdx = binIdx
		}
		if binIdx == 2 {
			// 9.3.3.1.2
		}
		if binIdx > 2 && binIdx < 6 {
			ctxIdx = 3
		}

	case 40:
		fallthrough
	case 47:
		switch binIdx {
		case 0:
			// 9.3.3.1.1.7
		case 1:
			ctxIdx = 3
		case 2:
			ctxIdx = 4
		case 3:
			ctxIdx = 5
		default:
			ctxIdx = 6
		}
	case 54:
		if binIdx == 0 {
			// 9.3.3.1.1.6
		}
		if binIdx == 1 {
			ctxIdx = 4
		}
		if binIdx > 1 {
			ctxIdx = 5
		}
	case 60:
		if binIdx == 0 {
			// 9.3.3.1.1.5
		}
		if binIdx == 1 {
			ctxIdx = 2
		}
		if binIdx > 1 {
			ctxIdx = 3
		}
	case 64:
		if binIdx == 0 {
			// 9.3.3.1.1.8
		} else if binIdx == 1 || binIdx == 2 {
			ctxIdx = 3
		} else {
			return NaCtxId
		}
	case 68:
		if binIdx != 0 {
			return NaCtxId
		}
		ctxIdx = 0
	case 69:
		if binIdx >= 0 && binIdx < 3 {
			ctxIdx = 0
		}
		return NaCtxId
	case 70:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.2
	case 73:
		switch binIdx {
		case 0:
			fallthrough
		case 1:
			fallthrough
		case 2:
			fallthrough
		case 3:
			// 9.3.3.1.1.4
		default:
			return NaCtxId
		}
	case 77:
		if binIdx == 0 {
			// 9.3.3.1.1.4
		} else if binIdx == 1 {
			// 9.3.3.1.1.4
		} else {
			return NaCtxId
		}
	case 276:
		if binIdx != 0 {
			return NaCtxId
		}
		ctxIdx = 0
	case 399:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.10
	}

	return ctxIdx
}
