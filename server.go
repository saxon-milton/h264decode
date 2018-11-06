package main

import "fmt"
import "net"
import "encoding/binary"
import "io"
import "gocv.io/x/gocv"

import "image"
import "image/color"

func listen(imageChan chan []byte) {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
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
	c.Close()
}
func jpegToMat(img []byte) (gocv.Mat, error) {
	return gocv.IMDecode(img, gocv.IMReadColor)
}
func featureExtractor(maxFeatures int, mat gocv.Mat) gocv.Mat {
	// Required for tracking features
	grayImage := gocv.NewMat()
	gocv.CvtColor(mat, &grayImage, gocv.ColorBGRToGray)
	red := color.RGBA{255, 100, 100, 90}
	corners := gocv.NewMat()
	gocv.GoodFeaturesToTrack(grayImage, &corners, maxFeatures, 0.01, 1.0)
	// Corners is a 2 dim array [ [x,y]...]
	for f := 0; f < maxFeatures; f++ {
		x := corners.GetFloatAt(f, 0)
		y := corners.GetFloatAt(f, 1)
		// Using GetInt yields out of range results
		point := image.Pt(int(x), int(y))
		gocv.Circle(&mat, point, 4, red, 2)
	}
	return mat
}

// Demo: Accepts images over the wire in [4 byte len of image, imagebytes] format
func main() {
	window := gocv.NewWindow("images")
	imageChan := make(chan []byte)
	defer close(imageChan)
	go listen(imageChan)
	for img := range imageChan {
		mat, err := jpegToMat(img)
		defer mat.Close()
		if err != nil {
			fmt.Println("unable to convert mat", err)
			break
		}
		mat = featureExtractor(500, mat)
		window.IMShow(mat)
		if key := window.WaitKey(1); key == 113 { // 'q'
			break

		}
	}
}
