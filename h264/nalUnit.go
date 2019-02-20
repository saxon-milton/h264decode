package h264

import "fmt"

type NalUnit struct {
	NumBytes             int
	ForbiddenZeroBit     int
	RefIdc               int
	Type                 int
	SvcExtensionFlag     int
	Avc3dExtensionFlag   int
	IdrFlag              int
	PriorityId           int
	NoInterLayerPredFlag int
	DependencyId         int
	QualityId            int
	TemporalId           int
	UseRefBasePicFlag    int
	DiscardableFlag      int
	OutputFlag           int
	ReservedThree2Bits   int
	HeaderBytes          int
	NonIdrFlag           int
	ViewId               int
	AnchorPicFlag        int
	InterViewFlag        int
	ReservedOneBit       int
	ViewIdx              int
	DepthFlag            int
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, b *BitReader, frame []byte) {
	// TODO: Annex G
	nalUnit.IdrFlag = b.NextField("IdrFlag", 1, frame)
	nalUnit.PriorityId = b.NextField("PriorityId", 6, frame)
	nalUnit.NoInterLayerPredFlag = b.NextField("NoInterLayerPredFlag", 1, frame)
	nalUnit.DependencyId = b.NextField("DependencyId", 3, frame)
	nalUnit.QualityId = b.NextField("QualityId", 4, frame)
	nalUnit.TemporalId = b.NextField("TemporalId", 3, frame)
	nalUnit.UseRefBasePicFlag = b.NextField("UseRefBasePicFlag", 1, frame)
	nalUnit.DiscardableFlag = b.NextField("DiscardableFlag", 1, frame)
	nalUnit.OutputFlag = b.NextField("OutputFlag", 1, frame)
	nalUnit.ReservedThree2Bits = b.NextField("ReservedThree2Bits", 2, frame)
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, b *BitReader, frame []byte) {
	// TODO: Annex J
	nalUnit.ViewIdx = b.NextField("ViewIdx", 8, frame)
	nalUnit.DepthFlag = b.NextField("DepthFlag", 1, frame)
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1, frame)
	nalUnit.TemporalId = b.NextField("TemporalId", 3, frame)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1, frame)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1, frame)
}
func NalUnitHeaderMvcExtension(nalUnit *NalUnit, b *BitReader, frame []byte) {
	// TODO: Annex H
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1, frame)
	nalUnit.PriorityId = b.NextField("PriorityId", 6, frame)
	nalUnit.ViewId = b.NextField("ViewId", 10, frame)
	nalUnit.TemporalId = b.NextField("TemporalId", 3, frame)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1, frame)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1, frame)
	nalUnit.ReservedOneBit = b.NextField("ReservedOneBit", 1, frame)
}

func NewNalUnit(frame []byte) NalUnit {
	fmt.Printf("== NalUnit %d\n", len(frame))
	fmt.Printf(" == %#v\n", frame[0:8])
	nalUnit := NalUnit{
		NumBytes:    len(frame),
		HeaderBytes: 1,
	}
	b := &BitReader{}
	nalUnit.ForbiddenZeroBit = b.NextField("ForbiddenZeroBit", 1, frame)
	nalUnit.RefIdc = b.NextField("NalRefIdc", 2, frame)
	nalUnit.Type = b.NextField("NalUnitType", 5, frame)

	if nalUnit.Type == 14 || nalUnit.Type == 20 || nalUnit.Type == 21 {
		if nalUnit.Type != 21 {
			nalUnit.SvcExtensionFlag = b.NextField("SvcExtensionFlag", 1, frame)
		} else {
			nalUnit.Avc3dExtensionFlag = b.NextField("Avc3dExtensionFlag", 1, frame)
		}
		if nalUnit.SvcExtensionFlag == 1 {
			NalUnitHeaderSvcExtension(&nalUnit, b, frame)
			nalUnit.HeaderBytes += 3
		} else if nalUnit.Avc3dExtensionFlag == 1 {
			NalUnitHeader3davcExtension(&nalUnit, b, frame)
			nalUnit.HeaderBytes += 2
		} else {
			NalUnitHeaderMvcExtension(&nalUnit, b, frame)
			nalUnit.HeaderBytes += 3

		}
	}

	for i := nalUnit.HeaderBytes; i < nalUnit.NumBytes; i++ {
		var nextBitsIs3 bool
		if frame[i] == byte(0) && frame[i+1] == byte(0) && frame[1+2] == byte(3) {
			nextBitsIs3 = true
		}
		if i+2 < nalUnit.NumBytes && nextBitsIs3 {
		}
	}

	return nalUnit
}
