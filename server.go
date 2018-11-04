package main

import "fmt"
import "net"
import "encoding/binary"
import "io"
import "gocv.io/x/gocv"

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
		image := make([]byte, imageSize)
		if err := binary.Read(c, binary.LittleEndian, &image); err != nil {
			fmt.Println("little endian read failed", err)
		} else {
			imageChan <- image
		}

	}
	c.Close()
}
func jpegToMat(image []byte) (gocv.Mat, error) {
	return gocv.IMDecode(image, gocv.IMReadColor)
}

// Demo: Accepts images over the wire in [4 byte len of image, imagebytes] format
func main() {
	window := gocv.NewWindow("images")
	imageChan := make(chan []byte)
	defer close(imageChan)
	go listen(imageChan)
	for image := range imageChan {
		mat, err := jpegToMat(image)
		if err != nil {
			fmt.Println("unable to convert mat", err)
			break
		}
		window.IMShow(mat)
		if key := window.WaitKey(1); key == 113 { // 'q'
			break

		}
	}
}
