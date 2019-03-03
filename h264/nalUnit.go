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
	rbsp                 []byte
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex G
	nalUnit.IdrFlag = b.NextField("IdrFlag", 1)
	nalUnit.PriorityId = b.NextField("PriorityId", 6)
	nalUnit.NoInterLayerPredFlag = b.NextField("NoInterLayerPredFlag", 1)
	nalUnit.DependencyId = b.NextField("DependencyId", 3)
	nalUnit.QualityId = b.NextField("QualityId", 4)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.UseRefBasePicFlag = b.NextField("UseRefBasePicFlag", 1)
	nalUnit.DiscardableFlag = b.NextField("DiscardableFlag", 1)
	nalUnit.OutputFlag = b.NextField("OutputFlag", 1)
	nalUnit.ReservedThree2Bits = b.NextField("ReservedThree2Bits", 2)
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex J
	nalUnit.ViewIdx = b.NextField("ViewIdx", 8)
	nalUnit.DepthFlag = b.NextField("DepthFlag", 1)
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1)
}
func NalUnitHeaderMvcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex H
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1)
	nalUnit.PriorityId = b.NextField("PriorityId", 6)
	nalUnit.ViewId = b.NextField("ViewId", 10)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1)
	nalUnit.ReservedOneBit = b.NextField("ReservedOneBit", 1)
}

func NewNalUnit(frame []byte) *NalUnit {
	fmt.Printf("== NalUnit %d\n", len(frame))
	fmt.Printf(" == %#v\n", frame[0:8])
	nalUnit := NalUnit{
		NumBytes:    len(frame),
		HeaderBytes: 1,
	}
	b := &BitReader{bytes: frame}
	nalUnit.ForbiddenZeroBit = b.NextField("ForbiddenZeroBit", 1)
	nalUnit.RefIdc = b.NextField("NalRefIdc", 2)
	nalUnit.Type = b.NextField("NalUnitType", 5)

	if nalUnit.Type == 14 || nalUnit.Type == 20 || nalUnit.Type == 21 {
		if nalUnit.Type != 21 {
			nalUnit.SvcExtensionFlag = b.NextField("SvcExtensionFlag", 1)
		} else {
			nalUnit.Avc3dExtensionFlag = b.NextField("Avc3dExtensionFlag", 1)
		}
		if nalUnit.SvcExtensionFlag == 1 {
			NalUnitHeaderSvcExtension(&nalUnit, b)
			nalUnit.HeaderBytes += 3
		} else if nalUnit.Avc3dExtensionFlag == 1 {
			NalUnitHeader3davcExtension(&nalUnit, b)
			nalUnit.HeaderBytes += 2
		} else {
			NalUnitHeaderMvcExtension(&nalUnit, b)
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

	nalUnit.rbsp = frame[nalUnit.HeaderBytes:]
	logger.Printf("info: decoded %s NAL\n", NALUnitType[nalUnit.Type])
	return &nalUnit
}
