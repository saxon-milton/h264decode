package h264

import "fmt"
import "math"

// import "strings"

// Specification Page 46 7.3.2.2

type PPS struct {
	ID, SPSID                         int
	EntropyCodingMode                 int
	NumSliceGroupsMinus1              int
	BottomFieldPicOrderInFramePresent bool
	NumSlicGroupsMinus1               int
	SliceGroupMapType                 int
	RunLengthMinus1                   []int
	TopLeft                           []int
	BottomRight                       []int
	SliceGroupChangeDirection         bool
	SliceGroupChangeRateMinus1        int
	PicSizeInMapUnitsMinus1           int
	SliceGroupId                      []int
	NumRefIdxL0DefaultActiveMinus1    int
	NumRefIdxL1DefaultActiveMinus1    int
	WeightedPred                      bool
	WeightedBipred                    int
	PicInitQpMinus26                  int
	PicInitQsMinus26                  int
	ChromaQpIndexOffset               int
	DeblockingFilterControlPresent    bool
	ConstrainedIntraPred              bool
	RedundantPicCntPresent            bool
	Transform8x8Mode                  int
	PicScalingMatrixPresent           bool
	PicScalingListPresent             []bool
	SecondChromaQpIndexOffset         int
}

func NewPPS(sps *SPS, rbsp []byte) PPS {
	fmt.Printf(" == PPS RBSP %d bytes %d bits == \n", len(rbsp), len(rbsp)*8)
	fmt.Printf(" == %#v\n", rbsp[0:8])
	pps := PPS{}
	b := &BitReader{}

	nextField := func(name string, bits int) int {
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

	pps.ID = ue(b.golomb(rbsp))
	pps.SPSID = ue(b.golomb(rbsp))
	pps.EntropyCodingMode = nextField("EntropyCodingModeFlag", 1)
	pps.BottomFieldPicOrderInFramePresent = flagField()
	pps.NumSliceGroupsMinus1 = ue(b.golomb(rbsp))
	if pps.NumSliceGroupsMinus1 > 0 {
		pps.SliceGroupMapType = ue(b.golomb(rbsp))
		if pps.SliceGroupMapType == 0 {
			for iGroup := 0; iGroup <= pps.NumSliceGroupsMinus1; iGroup++ {
				pps.RunLengthMinus1[iGroup] = ue(b.golomb(rbsp))
			}
		} else if pps.SliceGroupMapType == 2 {
			for iGroup := 0; iGroup < pps.NumSliceGroupsMinus1; iGroup++ {
				pps.TopLeft[iGroup] = ue(b.golomb(rbsp))
				pps.BottomRight[iGroup] = ue(b.golomb(rbsp))
			}
		} else if pps.SliceGroupMapType > 2 && pps.SliceGroupMapType < 6 {
			pps.SliceGroupChangeDirection = flagField()
			pps.SliceGroupChangeRateMinus1 = ue(b.golomb(rbsp))
		} else if pps.SliceGroupMapType == 6 {
			pps.PicSizeInMapUnitsMinus1 = ue(b.golomb(rbsp))
			for i := 0; i <= pps.PicSizeInMapUnitsMinus1; i++ {
				pps.SliceGroupId[i] = nextField(
					fmt.Sprintf("SliceGroupId[%d]", i),
					int(math.Ceil(math.Log2(float64(pps.NumSliceGroupsMinus1+1)))))
			}
		}

	}
	pps.NumRefIdxL0DefaultActiveMinus1 = ue(b.golomb(rbsp))
	pps.NumRefIdxL1DefaultActiveMinus1 = ue(b.golomb(rbsp))
	pps.WeightedPred = flagField()
	pps.WeightedBipred = nextField("WeightedBipredIDC", 2)
	pps.PicInitQpMinus26 = se(b.golomb(rbsp))
	pps.PicInitQsMinus26 = se(b.golomb(rbsp))
	pps.ChromaQpIndexOffset = se(b.golomb(rbsp))
	pps.DeblockingFilterControlPresent = flagField()
	pps.ConstrainedIntraPred = flagField()
	pps.RedundantPicCntPresent = flagField()

	if b.HasMoreData(rbsp) {
		pps.Transform8x8Mode = nextField("Transform8x8ModeFlag", 1)
		pps.PicScalingMatrixPresent = flagField()
		if pps.PicScalingMatrixPresent {
			v := 6
			if sps.ChromaFormat != 3 {
				v = 2
			}
			for i := 0; i < 6+(v*pps.Transform8x8Mode); i++ {
				pps.PicScalingListPresent[i] = flagField()
				if pps.PicScalingListPresent[i] {
					if i < 6 {
						scalingList(
							b, rbsp,
							ScalingList4x4[i],
							16,
							DefaultScalingMatrix4x4[i])

					} else {
						scalingList(
							b, rbsp,
							ScalingList8x8[i],
							64,
							DefaultScalingMatrix8x8[i-6])

					}
				}
			}
			pps.SecondChromaQpIndexOffset = se(b.golomb(rbsp))
		}
		// rbspTrailingBits()
	}

	debugPacket("PPS", pps)
	return pps

}
