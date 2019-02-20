package h264

const NaCtxId = 10000
const MbAddrNotAvailable = 10000

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

// 9.3.3.1.1 : returns ctxIdxInc
func Decoder9_3_3_1_1_1(condTermFlagA, condTermFlagB int) int {
	return condTermFlagA + condTermFlagB
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
