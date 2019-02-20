package h264

import (
	"fmt"
	"math"
)

type Slice struct {
	SliceHeader
}
type SliceHeader struct {
	FirstMbInSlice                   int
	SliceType                        int
	PPSID                            int
	ColorPlaneID                     int
	FrameNum                         int
	FieldPic                         bool
	BottomField                      bool
	IDRPicID                         int
	PicOrderCntLsb                   int
	DeltaPicOrderCntBottom           int
	DeltaPicOrderCnt                 []int
	RedundantPicCnt                  int
	DirectSpatialMvPred              bool
	NumRefIdxActiveOverride          bool
	NumRefIdxL0ActiveMinus1          int
	NumRefIdxL1ActiveMinus1          int
	CabacInit                        int
	SliceQpDelta                     int
	SpForSwitch                      bool
	SliceQsDelta                     int
	DisableDeblockingFilter          int
	SliceAlphaC0OffsetDiv2           int
	SliceBetaOffsetDiv2              int
	SliceGroupChangeCycle            int
	RefPicListModificationFlagL0     bool
	ModificationOfPicNums            int
	AbsDiffPicNumMinus1              int
	LongTermPicNum                   int
	RefPicListModificationFlagL1     bool
	LumaLog2WeightDenom              int
	ChromaLog2WeightDenom            int
	ChromaArrayType                  int
	LumaWeightL0Flag                 bool
	LumaWeightL0                     []int
	LumaOffsetL0                     []int
	ChromaWeightL0Flag               bool
	ChromaWeightL0                   [][]int
	ChromaOffsetL0                   [][]int
	LumaWeightL1Flag                 bool
	LumaWeightL1                     []int
	LumaOffsetL1                     []int
	ChromaWeightL1Flag               bool
	ChromaWeightL1                   [][]int
	ChromaOffsetL1                   [][]int
	NoOutputOfPriorPicsFlag          bool
	LongTermReferenceFlag            bool
	AdaptiveRefPicMarkingModeFlag    bool
	MemoryManagementControlOperation int
	DifferenceOfPicNumsMinus1        int
	LongTermFrameIdx                 int
	MaxLongTermFrameIdxPlus1         int
}

type SliceData struct {
	CabacAlignmentOneBit     int
	MbSkipRun                int
	MbSkipFlag               bool
	MbFieldDecodingFlag      bool
	EndOfSliceFlag           bool
	MbType                   int
	PcmAlignmentZeroBit      int
	PcmSampleLuma            []int
	PcmSampleChroma          []int
	TransformSize8x8Flag     bool
	CodedBlockPattern        int
	MbQpDelta                int
	PrevIntra4x4PredModeFlag []int
	RemIntra4x4PredMode      []int
	PrevIntra8x8PredModeFlag []int
	RemIntra8x8PredMode      []int
	IntraChromaPredMode      int
	RefIdxL0                 []int
	RefIdxL1                 []int
	MvdL0                    [][][]int
	MvdL1                    [][][]int
}

// Table 7-6
var sliceTypeMap = map[int]string{
	0: "P",
	1: "B",
	2: "I",
	3: "SP",
	4: "SI",
	5: "P",
	6: "B",
	7: "I",
	8: "SP",
	9: "SI",
}

func flagVal(b bool) int {
	if b {
		return 1
	}
	return 0
}

// context-adaptive arithmetic entropy-coded element (CABAC)
// 9.3
// When parsing the slice date of a slice (7.3.4) the initialization is 9.3.1
func (d SliceData) ae(v int) int {
	// 9.3.1.1 : CABAC context initialization ctxIdx
	return 0
}

// 8.2.2
func MbToSliceGroupMap(sps *SPS, pps *PPS, header *SliceHeader) []int {
	mbaffFrameFlag := 0
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		mbaffFrameFlag = 1
	}
	mapUnitToSliceGroupMap := MapUnitToSliceGroupMap(sps, pps, header)
	mbToSliceGroupMap := []int{}
	for i := 0; i <= PicSizeInMbs(sps, header)-1; i++ {
		if sps.FrameMbsOnly || header.FieldPic {
			mbToSliceGroupMap = append(mbToSliceGroupMap, mapUnitToSliceGroupMap[i])
			continue
		}
		if mbaffFrameFlag == 1 {
			mbToSliceGroupMap = append(mbToSliceGroupMap, mapUnitToSliceGroupMap[i/2])
			continue
		}
		if !sps.FrameMbsOnly && !sps.MBAdaptiveFrameField && !header.FieldPic {
			mbToSliceGroupMap = append(
				mbToSliceGroupMap,
				mapUnitToSliceGroupMap[(i/(2*PicWidthInMbs(sps)))*PicWidthInMbs(sps)+(i%PicWidthInMbs(sps))])
		}
	}
	return mbToSliceGroupMap

}
func PicWidthInMbs(sps *SPS) int {
	return sps.PicWidthInMbsMinus1 + 1
}
func PicHeightInMapUnits(sps *SPS) int {
	return sps.PicHeightInMapUnitsMinus1 + 1
}
func PicSizeInMapUnits(sps *SPS) int {
	return PicWidthInMbs(sps) * PicHeightInMapUnits(sps)
}
func FrameHeightInMbs(sps *SPS) int {
	return (2 - flagVal(sps.FrameMbsOnly)) * PicHeightInMapUnits(sps)
}
func PicHeightInMbs(sps *SPS, header *SliceHeader) int {
	return FrameHeightInMbs(sps) / (1 + flagVal(header.FieldPic))
}
func PicSizeInMbs(sps *SPS, header *SliceHeader) int {
	return PicWidthInMbs(sps) * PicHeightInMbs(sps, header)
}

// table 6-1
func SubWidthC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == 3 {
			return n
		}
	}

	switch sps.ChromaFormat {
	case 0:
		return n
	case 1:
		n = 2
	case 2:
		n = 2
	case 3:
		n = 1

	}
	return n
}
func SubHeightC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == 3 {
			return n
		}
	}
	switch sps.ChromaFormat {
	case 0:
		return n
	case 1:
		n = 2
	case 2:
		n = 1
	case 3:
		n = 1

	}
	return n
}

// 7-36
func CodedBlockPatternLuma(data *SliceData) int {
	return data.CodedBlockPattern % 16
}
func CodedBlockPatternChroma(data *SliceData) int {
	return data.CodedBlockPattern / 16
}

// dependencyId see Annex G.8.8.1
// Also G7.3.1.1 nal_unit_header_svc_extension
func DQId(nalUnit *NalUnit) int {
	return (nalUnit.DependencyId << 4) + nalUnit.QualityId
}

// Annex G p527
func NumMbPart(nalUnit *NalUnit, sps *SPS, header *SliceHeader, data *SliceData) int {
	sliceType := sliceTypeMap[header.SliceType]
	numMbPart := 0
	if MbTypeName(sliceType, CurrMbAddr(sps, header)) == "B_SKIP" || MbTypeName(sliceType, CurrMbAddr(sps, header)) == "B_Direct_16x16" {
		if DQId(nalUnit) == 0 && nalUnit.Type != 20 {
			numMbPart = 4
		} else if DQId(nalUnit) > 0 && nalUnit.Type == 20 {
			numMbPart = 1
		}
	} else if MbTypeName(sliceType, CurrMbAddr(sps, header)) != "B_SKIP" && MbTypeName(sliceType, CurrMbAddr(sps, header)) != "B_Direct_16x16" {
		numMbPart = CurrMbAddr(sps, header)

	}
	return numMbPart
}

func MbPred(nalUnit *NalUnit, sps *SPS, pps *PPS, header *SliceHeader, data *SliceData, b *BitReader, rbsp []byte) {
	sliceType := sliceTypeMap[header.SliceType]
	mbPartPredMode := MbPartPredMode(data, sliceType, data.MbType, 0)
	if mbPartPredMode == "Intra_4x4" || mbPartPredMode == "Intra_8x8" || mbPartPredMode == "Intra_16x16" {
		if mbPartPredMode == "Intra_4x4" {
			for luma4x4BlkIdx := 0; luma4x4BlkIdx < 16; luma4x4BlkIdx++ {
				var v int
				if pps.EntropyCodingMode == 1 {
					// TODO: 1 bit or ae(v)
					fmt.Printf("TODO: ae for PevIntra4x4PredModeFlag[%d]\n", luma4x4BlkIdx)
				} else {
					v = NextField(fmt.Sprintf("PrevIntra4x4PredModeFlag[%d]", luma4x4BlkIdx), 1, b, rbsp)
				}
				data.PrevIntra4x4PredModeFlag = append(data.PrevIntra4x4PredModeFlag, v)
				if data.PrevIntra4x4PredModeFlag[luma4x4BlkIdx] == 0 {
					if pps.EntropyCodingMode == 1 {
						// TODO: 3 bits or ae(v)
						fmt.Printf("TODO: ae for RemIntra4x4PredMode[%d]\n", luma4x4BlkIdx)
					} else {
						v = NextField(fmt.Sprintf("RemIntra4x4PredMode[%d]", luma4x4BlkIdx), 3, b, rbsp)
					}
					if len(data.RemIntra4x4PredMode) < luma4x4BlkIdx {
						data.RemIntra4x4PredMode = append(
							data.RemIntra4x4PredMode,
							make([]int, luma4x4BlkIdx-len(data.RemIntra4x4PredMode)+1)...)
					}
					data.RemIntra4x4PredMode[luma4x4BlkIdx] = v
				}
			}
		}
		if mbPartPredMode == "Intra_8x8" {
			for luma8x8BlkIdx := 0; luma8x8BlkIdx < 4; luma8x8BlkIdx++ {
				var v int
				if pps.EntropyCodingMode == 1 {
					// TODO: 1 bit or ae(v)
					fmt.Printf("TODO: ae for PrevIntra8x8PredModeFlag[%d]\n", luma8x8BlkIdx)
				} else {
					v = NextField(fmt.Sprintf("PrevIntra8x8PredModeFlag[%d]", luma8x8BlkIdx), 1, b, rbsp)
				}
				data.PrevIntra8x8PredModeFlag = append(data.PrevIntra8x8PredModeFlag, v)
				if data.PrevIntra8x8PredModeFlag[luma8x8BlkIdx] == 0 {
					if pps.EntropyCodingMode == 1 {
						// TODO: 3 bits or ae(v)
						fmt.Printf("TODO: ae for RemIntra8x8PredMode[%d]\n", luma8x8BlkIdx)
					} else {
						v = NextField(fmt.Sprintf("RemIntra8x8PredMode[%d]", luma8x8BlkIdx), 3, b, rbsp)
					}
					if len(data.RemIntra8x8PredMode) < luma8x8BlkIdx {
						data.RemIntra8x8PredMode = append(
							data.RemIntra8x8PredMode,
							make([]int, luma8x8BlkIdx-len(data.RemIntra8x8PredMode)+1)...)
					}
					data.RemIntra8x8PredMode[luma8x8BlkIdx] = v
				}
			}

		}
		if header.ChromaArrayType == 1 || header.ChromaArrayType == 2 {
			if pps.EntropyCodingMode == 1 {
				// TODO: ue(v) or ae(v)
				fmt.Printf("TODO: ae for IntraChromaPredMode\n")
			} else {
				data.IntraChromaPredMode = ue(b.golomb(rbsp))
			}
		}

	} else if mbPartPredMode != "Direct" {
		for mbPartIdx := 0; mbPartIdx < NumMbPart(nalUnit, sps, header, data); mbPartIdx++ {
			if (header.NumRefIdxL0ActiveMinus1 > 0 || data.MbFieldDecodingFlag != header.FieldPic) && MbPartPredMode(data, sliceType, data.MbType, mbPartIdx) != "Pred_L1" {
				fmt.Printf("\tTODO: refIdxL0[%d] te or ae(v)\n", mbPartIdx)
				if len(data.RefIdxL0) < mbPartIdx {
					data.RefIdxL0 = append(data.RefIdxL0, make([]int, mbPartIdx-len(data.RefIdxL0)+1)...)
				}
				if pps.EntropyCodingMode == 1 {
					// TODO: te(v) or ae(v)
					fmt.Printf("TODO: ae for RefIdxL0[%d]\n", mbPartIdx)
				} else {
					// TODO: Only one reference picture is used for inter-prediction,
					// then the value should be 0
					if MbaffFrameFlag(sps, header) == 0 || !data.MbFieldDecodingFlag {
						data.RefIdxL0[mbPartIdx] = te(b.golomb(rbsp), header.NumRefIdxL0ActiveMinus1)
					} else {
						rangeMax := 2*header.NumRefIdxL0ActiveMinus1 + 1
						data.RefIdxL0[mbPartIdx] = te(b.golomb(rbsp), rangeMax)
					}
				}
			}
		}
		for mbPartIdx := 0; mbPartIdx < NumMbPart(nalUnit, sps, header, data); mbPartIdx++ {
			if MbPartPredMode(data, sliceType, data.MbType, mbPartIdx) != "Pred_L1" {
				for compIdx := 0; compIdx < 2; compIdx++ {
					if len(data.MvdL0) < mbPartIdx {
						data.MvdL0 = append(
							data.MvdL0,
							make([][][]int, mbPartIdx-len(data.MvdL0)+1)...)
					}
					if len(data.MvdL0[mbPartIdx][0]) < compIdx {
						data.MvdL0[mbPartIdx][0] = append(
							data.MvdL0[mbPartIdx][0],
							make([]int, compIdx-len(data.MvdL0[mbPartIdx][0])+1)...)
					}
					if pps.EntropyCodingMode == 1 {
						// TODO: se(v) or ae(v)
						fmt.Printf("TODO: ae for MvdL0[%d][0][%d]\n", mbPartIdx, compIdx)
					} else {
						data.MvdL0[mbPartIdx][0][compIdx] = se(b.golomb(rbsp))
					}
				}
			}
		}
		for mbPartIdx := 0; mbPartIdx < NumMbPart(nalUnit, sps, header, data); mbPartIdx++ {
			if MbPartPredMode(data, sliceType, data.MbType, mbPartIdx) != "Pred_L0" {
				for compIdx := 0; compIdx < 2; compIdx++ {
					if len(data.MvdL1) < mbPartIdx {
						data.MvdL1 = append(
							data.MvdL1,
							make([][][]int, mbPartIdx-len(data.MvdL1)+1)...)
					}
					if len(data.MvdL1[mbPartIdx][0]) < compIdx {
						data.MvdL1[mbPartIdx][0] = append(
							data.MvdL0[mbPartIdx][0],
							make([]int, compIdx-len(data.MvdL1[mbPartIdx][0])+1)...)
					}
					if pps.EntropyCodingMode == 1 {
						// TODO: se(v) or ae(v)
						fmt.Printf("TODO: ae for MvdL1[%d][0][%d]\n", mbPartIdx, compIdx)
					} else {
						data.MvdL1[mbPartIdx][0][compIdx] = se(b.golomb(rbsp))
					}
				}
			}
		}

	}
}

// 8.2.2.1
func MapUnitToSliceGroupMap(sps *SPS, pps *PPS, header *SliceHeader) []int {
	mapUnitToSliceGroupMap := []int{}
	picSizeInMapUnits := PicSizeInMapUnits(sps)
	if pps.NumSliceGroupsMinus1 == 0 {
		// 0 to PicSizeInMapUnits -1 inclusive
		for i := 0; i <= picSizeInMapUnits-1; i++ {
			mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, 0)
		}
	} else {
		switch pps.SliceGroupMapType {
		case 0:
			// 8.2.2.1
			i := 0
			for i < picSizeInMapUnits {
				// iGroup should be incremented in the pps.RunLengthMinus1 index operation. There may be a bug here
				for iGroup := 0; iGroup <= pps.NumSliceGroupsMinus1 && i < picSizeInMapUnits; i += pps.RunLengthMinus1[iGroup+1] + 1 {
					for j := 0; j < pps.RunLengthMinus1[iGroup] && i+j < picSizeInMapUnits; j++ {
						if len(mapUnitToSliceGroupMap) < i+j {
							mapUnitToSliceGroupMap = append(
								mapUnitToSliceGroupMap,
								make([]int, (i+j)-len(mapUnitToSliceGroupMap)+1)...)
						}
						mapUnitToSliceGroupMap[i+j] = iGroup
					}
				}
			}
		case 1:
			// 8.2.2.2
			for i := 0; i < picSizeInMapUnits; i++ {
				v := ((i % PicWidthInMbs(sps)) + (((i / PicWidthInMbs(sps)) * (pps.NumSliceGroupsMinus1 + 1)) / 2)) % (pps.NumSliceGroupsMinus1 + 1)
				mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, v)
			}
		case 2:
			// 8.2.2.3
			for i := 0; i < picSizeInMapUnits; i++ {
				mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, pps.NumSliceGroupsMinus1)
			}
			for iGroup := pps.NumSliceGroupsMinus1 - 1; iGroup >= 0; iGroup-- {
				yTopLeft := pps.TopLeft[iGroup] / PicWidthInMbs(sps)
				xTopLeft := pps.TopLeft[iGroup] % PicWidthInMbs(sps)
				yBottomRight := pps.BottomRight[iGroup] / PicWidthInMbs(sps)
				xBottomRight := pps.BottomRight[iGroup] % PicWidthInMbs(sps)
				for y := yTopLeft; y <= yBottomRight; y++ {
					for x := xTopLeft; x <= xBottomRight; x++ {
						idx := y*PicWidthInMbs(sps) + x
						if len(mapUnitToSliceGroupMap) < idx {
							mapUnitToSliceGroupMap = append(
								mapUnitToSliceGroupMap,
								make([]int, idx-len(mapUnitToSliceGroupMap)+1)...)
							mapUnitToSliceGroupMap[idx] = iGroup
						}
					}
				}
			}

		case 3:
			// 8.2.2.4
			// TODO
		case 4:
			// 8.2.2.5
			// TODO
		case 5:
			// 8.2.2.6
			// TODO
		case 6:
			// 8.2.2.7
			// TODO
		}
	}
	// 8.2.2.8
	// Convert mapUnitToSliceGroupMap to MbToSliceGroupMap
	return mapUnitToSliceGroupMap
}
func nextMbAddress(n int, sps *SPS, pps *PPS, header *SliceHeader) int {
	i := n + 1
	// picSizeInMbs is the number of macroblocks in picture 0
	// 7-13
	// PicWidthInMbs = sps.PicWidthInMbsMinus1 + 1
	// PicHeightInMapUnits = sps.PicHeightInMapUnitsMinus1 + 1
	// 7-29
	// picSizeInMbs = PicWidthInMbs * PicHeightInMbs
	// 7-26
	// PicHeightInMbs = FrameHeightInMbs / (1 + header.fieldPicFlag)
	// 7-18
	// FrameHeightInMbs = (2 - ps.FrameMbsOnly) * PicHeightInMapUnits
	picWidthInMbs := sps.PicWidthInMbsMinus1 + 1
	picHeightInMapUnits := sps.PicHeightInMapUnitsMinus1 + 1
	frameHeightInMbs := (2 - flagVal(sps.FrameMbsOnly)) * picHeightInMapUnits
	picHeightInMbs := frameHeightInMbs / (1 + flagVal(header.FieldPic))
	picSizeInMbs := picWidthInMbs * picHeightInMbs
	mbToSliceGroupMap := MbToSliceGroupMap(sps, pps, header)
	for i < picSizeInMbs && mbToSliceGroupMap[i] != mbToSliceGroupMap[i] {
		i++
	}
	return i
}

func CurrMbAddr(sps *SPS, header *SliceHeader) int {
	mbaffFrameFlag := 0
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		mbaffFrameFlag = 1
	}

	return header.FirstMbInSlice * (1 * mbaffFrameFlag)
}

func MbaffFrameFlag(sps *SPS, header *SliceHeader) int {
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		return 1
	}
	return 0
}

func NextField(name string, bits int, b *BitReader, rbsp []byte) int {
	fmt.Printf("\t[%s] %d bits\n", name, bits)
	buf := make([]int, bits)
	if _, err := b.Read(rbsp, buf); err != nil {
		fmt.Printf("error reading %d bits for %s: %v\n", bits, name, err)
		return -1
	}
	return bitVal(buf)
}

func NewSliceData(nalUnit *NalUnit, sps *SPS, pps *PPS, header *SliceHeader, b *BitReader, rbsp []byte) SliceData {
	data := SliceData{}
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
	flagField := func() bool {
		if v := nextField("", 1); v == 1 {
			return true
		}
		return false
	}
	if pps.EntropyCodingMode == 1 {
		for !b.IsByteAligned() {
			data.CabacAlignmentOneBit = nextField("CabacAlignmentOneBit", 1)
		}
	}
	mbaffFrameFlag := 0
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		mbaffFrameFlag = 1
	}
	currMbAddr := header.FirstMbInSlice * (1 * mbaffFrameFlag)

	moreDataFlag := true
	prevMbSkipped := 0
	sliceType := sliceTypeMap[header.SliceType]
	for moreDataFlag {
		if sliceType != "I" && sliceType != "SI" {
			if pps.EntropyCodingMode == 0 {
				data.MbSkipRun = ue(b.golomb(rbsp))
				if data.MbSkipRun > 0 {
					prevMbSkipped = 1
				}
				for i := 0; i < data.MbSkipRun; i++ {
					// nextMbAddress(currMbAdd
					currMbAddr = nextMbAddress(currMbAddr, sps, pps, header)
				}
				if data.MbSkipRun > 0 {
					moreDataFlag = b.MoreRBSPData(rbsp)
				}
			} else {
				data.MbSkipFlag = flagField()
				moreDataFlag = !data.MbSkipFlag
			}
			if moreDataFlag {
				if mbaffFrameFlag == 1 && (currMbAddr%2 == 0 || (currMbAddr%2 == 1 && prevMbSkipped == 1)) {
					if pps.EntropyCodingMode == 1 {
						// TODO: ae implementation
						fmt.Printf("TODO: ae for MbFieldDecodingFlag\n")
					} else {
						data.MbFieldDecodingFlag = flagField()
					}
				}
				// BEGIN: macroblockLayer()
				if pps.EntropyCodingMode == 1 {
					// TODO: ae implementation
					fmt.Printf("TODO: ae implementation for MBType\n")
				} else {
					data.MbType = ue(b.golomb(rbsp))
				}
				if MbTypeName(sliceType, data.MbType) == "I_PCM" {
					for !b.IsByteAligned() {
						_ = nextField("PCMAlignmentZeroBit", 1)
					}
					// 7-3 p95
					bitDepthY := 8 + sps.BitDepthLumaMinus8
					for i := 0; i < 256; i++ {
						data.PcmSampleLuma = append(
							data.PcmSampleLuma,
							nextField(fmt.Sprintf("PcmSampleLuma[%d]", i), bitDepthY))
					}
					// 6-1 p 47
					mbWidthC := 16 / SubWidthC(sps)
					mbHeightC := 16 / SubHeightC(sps)
					// if monochrome
					if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
						mbWidthC = 0
						mbHeightC = 0
					}

					bitDepthC := 8 + sps.BitDepthChromaMinus8
					for i := 0; i < 2*mbWidthC*mbHeightC; i++ {
						data.PcmSampleChroma = append(
							data.PcmSampleChroma,
							nextField(fmt.Sprintf("PcmSampleChroma[%d]", i), bitDepthC))
					}

				} else {
					noSubMbPartSizeLessThan8x8Flag := 1
					if MbTypeName(sliceType, data.MbType) == "I_NxN" && MbPartPredMode(&data, sliceType, data.MbType, 0) != "Intra_16x16" && NumMbPart(nalUnit, sps, header, &data) == 4 {
						fmt.Printf("\tTODO: subMbPred\n")
						/*
							subMbType := SubMbPred(data.MbType)
							for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
								if subMbType[mbPartIdx] != "B_Direct_8x8" {
									if NumbSubMbPart(subMbType[mbPartIdx]) > 1 {
										noSubMbPartSizeLessThan8x8Flag = 0
									}
								} else if !sps.Direct8x8Inference {
									noSubMbPartSizeLessThan8x8Flag = 0
								}
							}
						*/
					} else {
						if pps.Transform8x8Mode == 1 && MbTypeName(sliceType, data.MbType) == "I_NxN" {
							// TODO
							// 1 bit or ae(v)
							// If pps.EntropyCodingMode == 1, use ae(v)
							if pps.EntropyCodingMode == 1 {
								fmt.Println("TODO: ae(v) for TransformSize8x8Flag")
							} else {
								data.TransformSize8x8Flag = flagField()
							}
						}
						MbPred(nalUnit, sps, pps, header, &data, b, rbsp)
					}
					if MbPartPredMode(&data, sliceType, data.MbType, 0) != "Intra_16x16" {
						// TODO: me, ae
						fmt.Printf("TODO: CodedBlockPattern pending me/ae implementation\n")
						if pps.EntropyCodingMode == 1 {
							fmt.Printf("TODO: ae for CodedBlockPattern\n")
						} else {
							data.CodedBlockPattern = me(
								b.golomb(rbsp),
								header.ChromaArrayType,
								MbPartPredMode(&data, sliceType, data.MbType, 0))
						}

						// data.CodedBlockPattern = me(v) | ae(v)
						if CodedBlockPatternLuma(&data) > 0 && pps.Transform8x8Mode == 1 && MbTypeName(sliceType, data.MbType) != "I_NxN" && noSubMbPartSizeLessThan8x8Flag == 1 && (MbTypeName(sliceType, data.MbType) != "B_Direct_16x16" || sps.Direct8x8Inference) {
							// TODO: 1 bit or ae(v)
							if pps.EntropyCodingMode == 1 {
								fmt.Printf("TODO: ae for TranformSize8x8Flag\n")
							} else {
								data.TransformSize8x8Flag = flagField()
							}
						}
					}
					if CodedBlockPatternLuma(&data) > 0 || CodedBlockPatternChroma(&data) > 0 || MbPartPredMode(&data, sliceType, data.MbType, 0) == "Intra_16x16" {
						// TODO: se or ae(v)
						if pps.EntropyCodingMode == 1 {
							fmt.Printf("TODO: ae for MbQpDelta\n")
						} else {
							data.MbQpDelta = se(b.golomb(rbsp))
						}

					}
				}

			}
			if pps.EntropyCodingMode == 0 {
				moreDataFlag = b.MoreRBSPData(rbsp)
			} else {
				if sliceType != "I" && sliceType != "SI" {
					if data.MbSkipFlag {
						prevMbSkipped = 1
					} else {
						prevMbSkipped = 0
					}
				}
				if mbaffFrameFlag == 1 && currMbAddr%2 == 0 {
					moreDataFlag = true
				} else {
					// TODO: ae implementation
					data.EndOfSliceFlag = flagField() // ae(b.golomb(rbsp))
					moreDataFlag = !data.EndOfSliceFlag
				}
			}
			currMbAddr = nextMbAddress(currMbAddr, sps, pps, header)
		}
	}
	return data
}

func NewSlice(nalUnit *NalUnit, sps *SPS, pps *PPS, rbsp []byte) Slice {
	fmt.Printf(" == %s RBSP %d bytes %d bits == \n", NALUnitType[nalUnit.Type], len(rbsp), len(rbsp)*8)
	fmt.Printf(" == %#v\n", rbsp[0:8])
	var idrPic bool
	if nalUnit.Type == 5 {
		idrPic = true
	}
	header := SliceHeader{}
	if sps.UseSeparateColorPlane {
		header.ChromaArrayType = 0
	} else {
		header.ChromaArrayType = sps.ChromaFormat
	}
	slice := Slice{SliceHeader: header}
	b := &BitReader{}
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
	flagField := func() bool {
		if v := nextField("", 1); v == 1 {
			return true
		}
		return false
	}
	header.FirstMbInSlice = ue(b.golomb(rbsp))
	header.SliceType = ue(b.golomb(rbsp))
	sliceType := sliceTypeMap[header.SliceType]
	fmt.Printf("== %s (%s) slice of %d bytes\n", NALUnitType[nalUnit.Type], sliceType, len(rbsp))
	header.PPSID = ue(b.golomb(rbsp))
	if sps.UseSeparateColorPlane {
		header.ColorPlaneID = nextField("ColorPlaneID", 2)
	}
	// TODO: See 7.4.3
	// header.FrameNum = nextField("FrameNum", 0)
	if !sps.FrameMbsOnly {
		header.FieldPic = flagField()
		if header.FieldPic {
			header.BottomField = flagField()
		}
	}
	if idrPic {
		header.IDRPicID = ue(b.golomb(rbsp))
	}
	if sps.PicOrderCountType == 0 {
		header.PicOrderCntLsb = nextField("PicOrderCntLsb", sps.Log2MaxPicOrderCntLSBMin4+4)
		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCntBottom = se(b.golomb(rbsp))
		}
	}
	if sps.PicOrderCountType == 1 && !sps.DeltaPicOrderAlwaysZero {
		header.DeltaPicOrderCnt[0] = se(b.golomb(rbsp))
		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCnt[1] = se(b.golomb(rbsp))
		}
	}
	if pps.RedundantPicCntPresent {
		header.RedundantPicCnt = ue(b.golomb(rbsp))
	}
	if sliceType == "B" {
		header.DirectSpatialMvPred = flagField()
	}
	if sliceType == "B" || sliceType == "SP" || sliceType == "B" {
		header.NumRefIdxActiveOverride = flagField()
		if header.NumRefIdxActiveOverride {
			header.NumRefIdxL0ActiveMinus1 = ue(b.golomb(rbsp))
			if sliceType == "B" {
				header.NumRefIdxL1ActiveMinus1 = ue(b.golomb(rbsp))
			}
		}
	}

	if nalUnit.Type == 20 || nalUnit.Type == 21 {
		// Annex H
		// H.7.3.3.1.1
		// refPicListMvcModifications()
	} else {
		// 7.3.3.1
		if header.SliceType%5 != 2 && header.SliceType%5 != 4 {
			header.RefPicListModificationFlagL0 = flagField()
			if header.RefPicListModificationFlagL0 {
				for header.ModificationOfPicNums != 3 {
					header.ModificationOfPicNums = ue(b.golomb(rbsp))
					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1 = ue(b.golomb(rbsp))
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum = ue(b.golomb(rbsp))
					}
				}
			}

		}
		if header.SliceType%5 == 1 {
			header.RefPicListModificationFlagL1 = flagField()
			if header.RefPicListModificationFlagL1 {
				for header.ModificationOfPicNums != 3 {
					header.ModificationOfPicNums = ue(b.golomb(rbsp))
					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1 = ue(b.golomb(rbsp))
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum = ue(b.golomb(rbsp))
					}
				}
			}
		}
		// refPicListModification()
	}

	if (pps.WeightedPred && (sliceType == "P" || sliceType == "SP")) || (pps.WeightedBipred == 1 && sliceType == "B") {
		// predWeightTable()
		header.LumaLog2WeightDenom = ue(b.golomb(rbsp))
		if header.ChromaArrayType != 0 {
			header.ChromaLog2WeightDenom = ue(b.golomb(rbsp))
		}
		for i := 0; i <= header.NumRefIdxL0ActiveMinus1; i++ {
			header.LumaWeightL0Flag = flagField()
			if header.LumaWeightL0Flag {
				header.LumaWeightL0 = append(header.LumaWeightL0, se(b.golomb(rbsp)))
				header.LumaOffsetL0 = append(header.LumaOffsetL0, se(b.golomb(rbsp)))
			}
			if header.ChromaArrayType != 0 {
				header.ChromaWeightL0Flag = flagField()
				if header.ChromaWeightL0Flag {
					header.ChromaWeightL0 = append(header.ChromaWeightL0, []int{})
					header.ChromaOffsetL0 = append(header.ChromaOffsetL0, []int{})
					for j := 0; j < 2; j++ {
						header.ChromaWeightL0[i] = append(header.ChromaWeightL0[i], se(b.golomb(rbsp)))
						header.ChromaOffsetL0[i] = append(header.ChromaOffsetL0[i], se(b.golomb(rbsp)))
					}
				}
			}
		}
		if header.SliceType%5 == 1 {
			for i := 0; i <= header.NumRefIdxL1ActiveMinus1; i++ {
				header.LumaWeightL1Flag = flagField()
				if header.LumaWeightL1Flag {
					header.LumaWeightL1 = append(header.LumaWeightL1, se(b.golomb(rbsp)))
					header.LumaOffsetL1 = append(header.LumaOffsetL1, se(b.golomb(rbsp)))
				}
				if header.ChromaArrayType != 0 {
					header.ChromaWeightL1Flag = flagField()
					if header.ChromaWeightL1Flag {
						header.ChromaWeightL1 = append(header.ChromaWeightL1, []int{})
						header.ChromaOffsetL1 = append(header.ChromaOffsetL1, []int{})
						for j := 0; j < 2; j++ {
							header.ChromaWeightL1[i] = append(header.ChromaWeightL1[i], se(b.golomb(rbsp)))
							header.ChromaOffsetL1[i] = append(header.ChromaOffsetL1[i], se(b.golomb(rbsp)))
						}
					}
				}
			}
		}
	} // end predWeightTable
	if nalUnit.RefIdc != 0 {
		// devRefPicMarking()
		if idrPic {
			header.NoOutputOfPriorPicsFlag = flagField()
			header.LongTermReferenceFlag = flagField()
		} else {
			header.AdaptiveRefPicMarkingModeFlag = flagField()
			if header.AdaptiveRefPicMarkingModeFlag {
				header.MemoryManagementControlOperation = ue(b.golomb(rbsp))
				for header.MemoryManagementControlOperation != 0 {
					if header.MemoryManagementControlOperation == 1 || header.MemoryManagementControlOperation == 3 {
						header.DifferenceOfPicNumsMinus1 = ue(b.golomb(rbsp))
					}
					if header.MemoryManagementControlOperation == 2 {
						header.LongTermPicNum = ue(b.golomb(rbsp))
					}
					if header.MemoryManagementControlOperation == 3 || header.MemoryManagementControlOperation == 6 {
						header.LongTermFrameIdx = ue(b.golomb(rbsp))
					}
					if header.MemoryManagementControlOperation == 4 {
						header.MaxLongTermFrameIdxPlus1 = ue(b.golomb(rbsp))
					}
				}
			}
		} // end decRefPicMarking
	}
	if pps.EntropyCodingMode == 1 && sliceType != "I" && sliceType != "SI" {
		header.CabacInit = ue(b.golomb(rbsp))
	}
	header.SliceQpDelta = se(b.golomb(rbsp))
	if sliceType == "SP" || sliceType == "SI" {
		if sliceType == "SP" {
			header.SpForSwitch = flagField()
		}
		header.SliceQsDelta = se(b.golomb(rbsp))
	}
	if pps.DeblockingFilterControlPresent {
		header.DisableDeblockingFilter = ue(b.golomb(rbsp))
		if header.DisableDeblockingFilter != 1 {
			header.SliceAlphaC0OffsetDiv2 = se(b.golomb(rbsp))
			header.SliceBetaOffsetDiv2 = se(b.golomb(rbsp))
		}
	}
	if pps.NumSliceGroupsMinus1 > 0 && pps.SliceGroupMapType >= 3 && pps.SliceGroupMapType <= 5 {
		header.SliceGroupChangeCycle = nextField(
			"SliceGroupChangeCycle",
			int(math.Ceil(math.Log2(float64(pps.PicSizeInMapUnitsMinus1/pps.SliceGroupChangeRateMinus1+1)))))
	}

	debugPacket("Slice", header)
	return slice
}
