package main

import (
	"encoding/binary"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

var (
	blue       = color.RGBA{0, 0, 255, 0}
	red        = color.RGBA{255, 100, 100, 0}
	green      = color.RGBA{100, 255, 100, 0}
	saveVideos = true
)

// 24k is within a few feet for a full sized human
const MaxArea = 24000
const MinArea = 200

func parseArgs() {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "nosave":
			saveVideos = false
		case "help":
			fmt.Printf("%s [nosave|help]\n", os.Args[0])
			os.Exit(0)
		}
	}
}
func timestamp() string {
	return time.Now().Format("2006.01.02_150405")
}

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
			fmt.Println("%s: error little endian read failed %v\n", timestamp(), err)
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
func writeLog(area float64) {

	f, err := os.OpenFile("events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s: error opening log: %v\n", timestamp(), err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%v: %2f sized object \n", time.Now(), area))
	if err != nil {
		fmt.Printf("%s: error writing entry: %v\n", timestamp(), err)
	}
}
func videoWriter(imchan chan gocv.Mat, donechan chan int) {
	fmt.Printf("video writer waiting...\n")
	var vw *gocv.VideoWriter
	var err error
	var filename string
	setupWriter := func() {
		filename = time.Now().Format(time.RFC3339) + ".avi"
		fps := 2.0
		vw, err = gocv.VideoWriterFile(filename, "MJPG", fps, 640, 480, true)
		if err != nil {
			fmt.Printf("%s: error unable to start video writer %s: %v\n", timestamp(), filename, err)
			return
		}
	}
	setupWriter()
	defer vw.Close()
	frame := 0
	for {
		select {
		case img := <-imchan:
			if !saveVideos {
				continue
			}

			/*
				im, err := img.ToImage()
				if err != nil {
					fmt.Printf("unable to convert image to im: %v\n", err)
					continue
				}
			*/
			if err := vw.Write(img); err != nil {
				fmt.Printf("[%s] failed to write frame %d: %v\n", filename, frame, err)
			}
			frame++
		case <-donechan:
			if frame > 1 {
				frame = 0
				fmt.Println("closing file:", filename)
				vw.Close()
				setupWriter()
			}
		}
	}

}

func motionDetector(window *gocv.Window, imchan chan gocv.Mat) {
	fmt.Printf("motion detector launched...\n")
	mog2 := gocv.NewBackgroundSubtractorMOG2()
	// Interesting should be true until a timer expires without it being refreshed
	sinceInteresting := 100
	videochan := make(chan gocv.Mat)
	closevideochan := make(chan int)
	go videoWriter(videochan, closevideochan)
	for img := range imchan {
		if img.Empty() {
			img.Close()
			continue
		}
		if sinceInteresting < 10 {
			videochan <- img
		} else {
			closevideochan <- 1
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
			if area < MinArea || area > MaxArea {
				sinceInteresting++
				continue
			}
			// image is about 30,000
			writeLog(area)
			gocv.DrawContours(&img, contours, i, red, 2)

			rect := gocv.BoundingRect(c)
			gocv.Rectangle(&img, rect, blue, 2)
			sinceInteresting = 0

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
	parseArgs()
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
