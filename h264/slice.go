package h264

import (
	"fmt"
	"math"
)

type Slice struct {
	SliceHeader
}
type SliceHeader struct {
	FirstMbInSlice               int
	SliceType                    int
	PPSID                        int
	ColorPlaneID                 int
	FrameNum                     int
	FieldPic                     bool
	BottomField                  bool
	IDRPicID                     int
	PicOrderCntLsb               int
	DeltaPicOrderCntBottom       int
	DeltaPicOrderCnt             []int
	RedundantPicCnt              int
	DirectSpatialMvPred          bool
	NumRefIdxActiveOverride      bool
	NumRefIdxL0ActiveMinus1      int
	NumRefIdxL1ActiveMinus1      int
	CabacInit                    int
	SliceQpDelta                 int
	SpForSwitch                  bool
	SliceQsDelta                 int
	DisableDeblockingFilter      int
	SliceAlphaC0OffsetDiv2       int
	SliceBetaOffsetDiv2          int
	SliceGroupChangeCycle        int
	RefPicListModificationFlagL0 bool
	ModificationOfPicNums        int
	AbsDiffPicNumMinus1          int
	LongTermPicNum               int
	RefPicListModificationFlagL1 bool
	LumaLog2WeightDenom          int
	ChromaLog2WeightDenom        int
	ChromaArrayType              int
	LumaWeightL0Flag             bool
	LumaWeightL0                 []int
	LumaOffsetL0                 []int
	ChromaWeightL0Flag           bool
	ChromaWeightL0               [][]int
	ChromaOffsetL0               [][]int
	LumaWeightL1Flag             bool
	LumaWeightL1                 []int
	LumaOffsetL1                 []int
	ChromaWeightL1Flag           bool
	ChromaWeightL1               [][]int
	ChromaOffsetL1               [][]int
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

func NewSlice(nalUnitType, nalRef int, sps *SPS, pps *PPS, rbsp []byte) Slice {
	fmt.Printf(" == %s RBSP %d bytes %d bits == \n", NALUnitType[nalUnitType], len(rbsp), len(rbsp)*8)
	fmt.Printf(" == %#v\n", rbsp[0:8])
	var idrPic bool
	if nalUnitType == 5 {
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
	fmt.Printf("== %s (%s) slice of %d bytes\n", NALUnitType[nalUnitType], sliceType, len(rbsp))
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

	if nalUnitType == 20 || nalUnitType == 21 {
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
	if nalRef != 0 {
		// devRefPicMarking()
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
