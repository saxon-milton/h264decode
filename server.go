package main

import "fmt"
import "net"
import "encoding/binary"
import "io"
import "gocv.io/x/gocv"
import "os"
import "image"
import "image/color"

// import "sort"

var (
	blue  = color.RGBA{0, 0, 255, 0}
	red   = color.RGBA{255, 100, 100, 0}
	green = color.RGBA{100, 255, 100, 0}
)

const MaxArea = 1000

func listen(imageChan chan []byte) {
	fmt.Println("Starting listener")
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("failed to listen", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening...")
	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go LittleEndianStructHandler(c, imageChan)
	}
}

// LittleEndianStructHandler Read little endian packed Python struct
func LittleEndianStructHandler(c net.Conn, imageChan chan []byte) {
	defer c.Close()
	for {
		// Read the size of the image in bytes being sent
		b := make([]byte, 4)
		_, err := c.Read(b)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("error reading image size")
			panic(err)
		}
		imageSize := binary.LittleEndian.Uint32(b)
		// Read the little endian image
		img := make([]byte, imageSize)
		if err := binary.Read(c, binary.LittleEndian, &img); err != nil {
			fmt.Println("little endian read failed", err)
		} else {
			imageChan <- img
		}

	}
}
func jpegToMat(img []byte) (gocv.Mat, error) {
	return gocv.IMDecode(img, gocv.IMReadColor)
}
func featureExtractor(maxFeatures int, mat gocv.Mat) []image.Point {
	// Required for tracking features
	grayImage := gocv.NewMat()
	defer grayImage.Close()
	gocv.CvtColor(mat, &grayImage, gocv.ColorBGRToGray)
	corners := gocv.NewMat()
	defer corners.Close()
	gocv.GoodFeaturesToTrack(grayImage, &corners, maxFeatures, 0.01, 1.0)
	// Corners is a 2 dim array [ [x,y]...]
	points := []image.Point{}
	for f := 0; f < corners.Rows(); f++ {
		x := corners.GetFloatAt(f, 0)
		y := corners.GetFloatAt(f, 1)
		// Using GetInt yields out of range results
		points = append(points, image.Pt(int(x), int(y)))
	}
	return points
}

func circleFeatures(features []image.Point, mat gocv.Mat) gocv.Mat {
	for _, feature := range features {
		go func(mat *gocv.Mat, pt image.Point) {
			gocv.Circle(mat, pt, 4, red, 2)
		}(&mat, feature)
	}
	return mat

}

func keyPointsAndDescriptors(img gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {

	// In frame a, then b
	orb := gocv.NewORB()
	defer orb.Close()

	m := gocv.NewMat()
	defer m.Close()
	return orb.DetectAndCompute(img, m)
}

func dropLowKeyPoints(kps []gocv.KeyPoint) []gocv.KeyPoint {
	hqkps := []gocv.KeyPoint{}
	minOctave := 100
	for _, kp := range kps {
		if kp.Octave < minOctave {
			minOctave = kp.Octave
		}
		// 50% derez at 1
		if kp.Octave < 1 {
			hqkps = append(hqkps, kp)
		}
	}
	return hqkps

}
func motionDetector(window *gocv.Window, imchan chan gocv.Mat) {
	fmt.Printf("motion detector launched...\n")
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	for img := range imchan {

		if img.Empty() {
			img.Close()
			continue
		}

		// Work off a smaller gray image
		grayImage := gocv.NewMat()
		gocv.CvtColor(img, &grayImage, gocv.ColorBGRToGray)
		fgMask := gocv.NewMat()
		imgThresh := gocv.NewMat()

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(grayImage, &fgMask)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		// gocv.Threshold(fgMask, &imgThresh, 25, 255, gocv.ThresholdBinary)
		// AdaptiveThresholdMean=0, Gaussian1
		blockSize := 255 // %2 == 1
		gocv.AdaptiveThreshold(fgMask, &imgThresh, 255, gocv.AdaptiveThresholdGaussian, gocv.ThresholdBinary, blockSize, 2)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		gocv.Dilate(imgThresh, &imgThresh, kernel)

		// now find contours
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		for i, c := range contours {
			area := gocv.ContourArea(c)
			if area < MaxArea {
				continue
			}

			gocv.DrawContours(&img, contours, i, red, 2)

			rect := gocv.BoundingRect(c)
			gocv.Rectangle(&img, rect, blue, 2)
		}
		window.IMShow(img)
		window.WaitKey(1)
		kernel.Close()
		imgThresh.Close()
		fgMask.Close()
		grayImage.Close()
		img.Close()
	}
	mog2.Close()

}
func main() {
	window := gocv.NewWindow("images")
	defer window.Close()
	// Open imagestream
	imstream := make(chan gocv.Mat)
	defer close(imstream)
	go motionDetector(window, imstream)
	// Start bystream
	bytestream := make(chan []byte)
	defer close(bytestream)
	go listen(bytestream)

	for imgBytes := range bytestream {
		img, err := jpegToMat(imgBytes)
		if err == nil {
			imstream <- img
		}
	}
}

// Demo: Accepts images over the wire in [4 byte len of image, imagebytes] format
func keypointDetector(window *gocv.Window, imchan chan gocv.Mat) {
	var imgCounter int64
	for img := range imchan {
		lastImage := []gocv.Mat{}
		window.IMShow(img)
		window.WaitKey(10)
		img.Close()
		// Newest frame always at head
		if imgCounter > 5 {
			if len(lastImage) < 1 {
				lastImage = append(lastImage, img)
				continue
			}
			// The descriptor will be n matches by 32 descriptor bits
			akps, ades := keyPointsAndDescriptors(img)
			akps = dropLowKeyPoints(akps)
			fmt.Printf("retained %d keypoints\n", len(akps))
			// Annotate image
			// With a keypoint, a high octave means more de-res, 0 is hi-res > 0 is lessening
			// The keypoint gives the center of the keypoint circular region
			// The scale is the diameter of the region, and angle in radians, the orientation
			drawKeypoints := func(kps []gocv.KeyPoint, c color.RGBA) {
				for _, kp := range kps {
					pt := image.Pt(int(kp.X), int(kp.Y))
					gocv.Circle(&img, pt, 2, c, 1)
				}
			}
			bkps, bdes := keyPointsAndDescriptors(lastImage[0])
			bkps = dropLowKeyPoints(bkps)
			drawKeypoints(akps, blue)
			drawKeypoints(bkps, green)
			window.IMShow(img)
			fmt.Printf("A kps %d, B kps %d\n", len(akps), len(bkps))
			/*
				if len(akps) > 0 && len(bkps) > 0 {
					sa := sortableKeyPoints{akps}
					sort.Sort(sa)
					sb := sortableKeyPoints{bkps}
					sort.Sort(sb)
					// TODO: Create rectangles from the keypoints
					fmt.Printf("A kps0: %#v\nB kps0: %#v\n", sa.kps[0], sb.kps[0])
				}
			*/
			lastImage[0].Close()
			lastImage[0] = img
			bf := gocv.NewBFMatcher()
			// Query, train, points
			dmatch := bf.KnnMatch(bdes, ades, 10)
			if len(dmatch) > 0 {
				// Descriptors by 32-bytes (256 bit) descriptor (see BRIEF-32 descriptor)
				//			fmt.Printf("matches %d by %d\n", len(dmatch), len(dmatch[0]))
				//			fmt.Printf("bdes %#v\n", bdes.Size())
				// fmt.Printf("d bdes %d, %d\n", bdes.Cols(), bdes.Rows())
				// fmt.Printf("Type %#v Total %#v\n", bdes.Type(), bdes.Total())
				// Each dmatch has the QueryIdx and TrainIdx for finding keypoints
				// The number of dmatch rows is the lesser keypoints set size
				// The number of columns matches our requested terms
				// Finding the smaller *kps set
			}
			// deferred .Close() was causing memory leak
			ades.Close()
			bdes.Close()
			bf.Close()
		}

		if key := window.WaitKey(100); key == 113 { // 'q'
			break

		}
	}
}
