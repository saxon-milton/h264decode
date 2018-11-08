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
	red   = color.RGBA{255, 100, 100, 90}
	green = color.RGBA{100, 255, 100, 90}
)

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

// Demo: Accepts images over the wire in [4 byte len of image, imagebytes] format
func main() {
	window := gocv.NewWindow("images")
	defer window.Close()
	imageChan := make(chan []byte)
	defer close(imageChan)
	go listen(imageChan)
	var imgCounter int64
	lastImage := []gocv.Mat{}
	for imgBytes := range imageChan {
		imgCounter++
		fmt.Println("img", imgCounter)
		img, err := jpegToMat(imgBytes)
		defer img.Close()
		if err != nil {
			fmt.Println("unable to convert img", err)
			break
		}
		// Newest frame always at head
		if imgCounter > 5 {
			if len(lastImage) < 1 {
				lastImage = append(lastImage, img)
				continue
			}
			akps, ades := keyPointsAndDescriptors(img)
			// Annotate image
			// With a keypoint, a high octave means more de-res, 0 is hi-res > 0 is lessening
			// The keypoint gives the center of the keypoint circular region
			// The scale is the diameter of the region, and angle in radians, the orientation
			for i, kp := range akps {
				if i == 10 {
					fmt.Printf("Keypoint at %2f, %2f\n", kp.X, kp.Y)
				}
				pt := image.Pt(int(kp.X), int(kp.Y))
				gocv.Circle(&img, pt, 2, blue, 1)
			}
			window.IMShow(img)
			bkps, bdes := keyPointsAndDescriptors(lastImage[0])
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
				fmt.Printf("matches %d by %d\n", len(dmatch), len(dmatch[0]))
				fmt.Printf("bdes %#v\n", bdes.Size())
				for row := bdes.Rows() - 3; row < bdes.Rows(); row++ {
					fmt.Println("Descriptors ", row)
					for col := bdes.Cols() - 3; col < bdes.Cols(); col++ {
						fmt.Printf("\t%d,%d: %2f\n", col, row, bdes.GetFloatAt(row, col))
					}
				}
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

		if key := window.WaitKey(1); key == 113 { // 'q'
			break

		}
	}
}
