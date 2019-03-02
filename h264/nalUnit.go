package h264

type NalUnit struct {
	NumBytesInNalUnit            int
	NumBytesInRBSP               int
	ForbiddenZeroBit             int
	RefIdc                       int
	Type                         int
	SvcExtensionFlag             int
	Avc3dExtensionFlag           int
	IdrFlag                      int
	PriorityId                   int
	NoInterLayerPredFlag         int
	DependencyId                 int
	QualityId                    int
	TemporalId                   int
	UseRefBasePicFlag            int
	DiscardableFlag              int
	OutputFlag                   int
	ReservedThree2Bits           int
	HeaderBytes                  int
	NonIdrFlag                   int
	ViewId                       int
	AnchorPicFlag                int
	InterViewFlag                int
	ReservedOneBit               int
	ViewIdx                      int
	DepthFlag                    int
	RBSPBytes                    []byte
	EmulationPreventionThreeByte int
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, r *RBSPReader) {
	// TODO: Annex G
	nalUnit.IdrFlag = r.GetFieldValue("IdrFlag", 1)
	nalUnit.PriorityId = r.GetFieldValue("PriorityId", 6)
	nalUnit.NoInterLayerPredFlag = r.GetFieldValue("NoInterLayerPredFlag", 1)
	nalUnit.DependencyId = r.GetFieldValue("DependencyId", 3)
	nalUnit.QualityId = r.GetFieldValue("QualityId", 4)
	nalUnit.TemporalId = r.GetFieldValue("TemporalId", 3)
	nalUnit.UseRefBasePicFlag = r.GetFieldValue("UseRefBasePicFlag", 1)
	nalUnit.DiscardableFlag = r.GetFieldValue("DiscardableFlag", 1)
	nalUnit.OutputFlag = r.GetFieldValue("OutputFlag", 1)
	nalUnit.ReservedThree2Bits = r.GetFieldValue("ReservedThree2Bits", 2)
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, r *RBSPReader) {
	// TODO: Annex J
	nalUnit.ViewIdx = r.GetFieldValue("ViewIdx", 8)
	nalUnit.DepthFlag = r.GetFieldValue("DepthFlag", 1)
	nalUnit.NonIdrFlag = r.GetFieldValue("NonIdrFlag", 1)
	nalUnit.TemporalId = r.GetFieldValue("TemporalId", 3)
	nalUnit.AnchorPicFlag = r.GetFieldValue("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = r.GetFieldValue("InterViewFlag", 1)
}
func NalUnitHeaderMvcExtension(nalUnit *NalUnit, r *RBSPReader) {
	// TODO: Annex H
	nalUnit.NonIdrFlag = r.GetFieldValue("NonIdrFlag", 1)
	nalUnit.PriorityId = r.GetFieldValue("PriorityId", 6)
	nalUnit.ViewId = r.GetFieldValue("ViewId", 10)
	nalUnit.TemporalId = r.GetFieldValue("TemporalId", 3)
	nalUnit.AnchorPicFlag = r.GetFieldValue("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = r.GetFieldValue("InterViewFlag", 1)
	nalUnit.ReservedOneBit = r.GetFieldValue("ReservedOneBit", 1)
}

func NewNalUnit(r *RBSPReader, numBytesInNalUnit int) *NalUnit {
	logger.Printf("debug: reading %d byte NALU\n", numBytesInNalUnit)
	nalUnit := NalUnit{
		NumBytesInNalUnit: numBytesInNalUnit,
		NumBytesInRBSP:    0,
		HeaderBytes:       1,
	}
	nalUnit.ForbiddenZeroBit = r.GetFieldValue("ForbiddenZeroBit", 1)
	nalUnit.RefIdc = r.GetFieldValue("NalRefIdc", 2)
	nalUnit.Type = r.GetFieldValue("NalUnitType", 5)

	logger.Printf("info: NAL Unit %d type %s: %d bytes\n", nalUnit.Type, NALUnitType[nalUnit.Type], numBytesInNalUnit)
	if nalUnit.Type == 14 || nalUnit.Type == 20 || nalUnit.Type == 21 {
		if nalUnit.Type != 21 {
			nalUnit.SvcExtensionFlag = r.GetFieldValue("SvcExtensionFlag", 1)
		} else {
			nalUnit.Avc3dExtensionFlag = r.GetFieldValue("Avc3dExtensionFlag", 1)
		}
		if nalUnit.SvcExtensionFlag == 1 {
			NalUnitHeaderSvcExtension(&nalUnit, r)
			nalUnit.HeaderBytes += 3
		} else if nalUnit.Avc3dExtensionFlag == 1 {
			NalUnitHeader3davcExtension(&nalUnit, r)
			nalUnit.HeaderBytes += 2
		} else {
			NalUnitHeaderMvcExtension(&nalUnit, r)
			nalUnit.HeaderBytes += 3

		}
	}
	logger.Printf("debug: finding end with %d headerbytes and %d NumBytesInRBSP\n",
		nalUnit.HeaderBytes,
		nalUnit.NumBytesInRBSP)
	for i := nalUnit.HeaderBytes; i < numBytesInNalUnit; i++ {
		if i+2 < nalUnit.NumBytesInRBSP && r.NextBitsValue(24) == 3 {
			nalUnit.NumBytesInRBSP += 1
			nalUnit.RBSPBytes = append(nalUnit.RBSPBytes, byte(nalUnit.NumBytesInRBSP))
			nalUnit.NumBytesInRBSP += 1
			nalUnit.RBSPBytes = append(nalUnit.RBSPBytes, byte(nalUnit.NumBytesInRBSP))
			i += 2
			nalUnit.EmulationPreventionThreeByte = r.GetFieldValue("EmulationPreventionThreeByte", 8)
		} else {
			nalUnit.NumBytesInRBSP += 1
			nalUnit.RBSPBytes = append(nalUnit.RBSPBytes, byte(nalUnit.NumBytesInRBSP))
		}
	}

	return &nalUnit
}
