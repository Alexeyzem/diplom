package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gocv.io/x/gocv"
)

const (
	threshold      = 10
	frameWidth     = 1280
	frameHeight    = 720
	brightness     = 0.6
	maxPercentSize = 1
	minPercentSize = 0.03
	maxColor       = 255
	center         = "Centered"
	moveCamera     = "Move camera: "
	markerID       = "Marker ID: "
	left           = "left"
	right          = "right"
	up             = "up"
	down           = "down"
	defaultDevice  = 0
	pointRadius    = 8
	pointThickness = 4
	textThickness  = 5
	textFontScale  = 3
	lineThickness  = 4
	leftUpPtX      = 10
	leftUpPtY      = 50
)

func main() {
	fmt.Println("START PROGRAM")
	fmt.Println("---------------")

	fmt.Println("Start init webcam")
	webcam, err := gocv.VideoCaptureDevice(defaultDevice)
	if err != nil {
		panic(err)
	}
	fmt.Println("Init webcam successful")
	defer webcam.Close()

	webcam.Set(gocv.VideoCaptureFrameWidth, frameWidth)
	webcam.Set(gocv.VideoCaptureFrameHeight, frameHeight)
	webcam.Set(gocv.VideoCaptureBrightness, brightness)

	fmt.Println("Start init window")
	window := gocv.NewWindow("ArUco Navigator")
	defer window.Close()
	fmt.Println("Init window successful")

	//Init ArUco
	fmt.Println("Start init aruco detector")
	arucoDict := gocv.GetPredefinedDictionary(gocv.ArucoDict4x4_1000)
	parameters := gocv.NewArucoDetectorParameters()
	parameters.SetMaxMarkerPerimeterRate(maxPercentSize)
	parameters.SetMinMarkerPerimeterRate(minPercentSize)

	detector := gocv.NewArucoDetectorWithParams(arucoDict, parameters)
	fmt.Println("Init aruco detector successful")

	img := gocv.NewMat()
	defer img.Close()

	textColor := color.RGBA{G: maxColor}                //OK
	errorColor := color.RGBA{R: maxColor}               //Error
	centerColor := color.RGBA{G: maxColor, B: maxColor} //Img center
	markerColor := color.RGBA{R: maxColor}              //Marker center
	lineColor := color.RGBA{R: maxColor, G: maxColor}   //Line between center

	fmt.Println("Start cycle")
	fmt.Println("---------------")
	for {
		if ok := webcam.Read(&img); !ok || img.Empty() {
			continue
		}

		// detect marker
		corners, ids, _ := detector.DetectMarkers(img)
		directions := make(map[string]struct{})

		fmt.Printf("Detected markers: %d\n", len(ids))

		if len(ids) > 0 {
			markerCenterX, markerCenterY := getCenterMarker(corners[0])
			imgCenterX, imgCenterY := getCenterImg(img)

			dx := markerCenterX - imgCenterX
			dy := markerCenterY - imgCenterY

			// Формируем направления
			if dx < -threshold {
				directions[right] = struct{}{}
			} else if dx > threshold {
				directions[left] = struct{}{}
			}

			if dy < -threshold {
				directions[down] = struct{}{}
			} else if dy > threshold {
				directions[up] = struct{}{}
			}

			//img center
			gocv.Circle(&img, image.Pt(int(imgCenterX), int(imgCenterY)), pointRadius, centerColor, pointRadius)
			//marker center
			gocv.Circle(
				&img, image.Pt(int(markerCenterX), int(markerCenterY)), pointRadius, markerColor, pointThickness,
			)
			//line between center
			gocv.Line(
				&img,
				image.Pt(int(imgCenterX), int(imgCenterY)),
				image.Pt(int(markerCenterX), int(markerCenterY)),
				lineColor,
				lineThickness,
			)

			gocv.ArucoDrawDetectedMarkers(img, corners, ids, gocv.NewScalar(0, maxColor, 0, 0))

			statusText := fmt.Sprintf("%s. %s%d", center, markerID, ids[0])
			statusColor := textColor

			if len(directions) > 0 {
				statusText = fmt.Sprintf(
					"%s%d. %s %s", markerID, ids[0], moveCamera, strings.Join(getMapKeys(directions), " + "),
				)
				statusColor = errorColor
			}

			gocv.PutText(
				&img, statusText, image.Pt(leftUpPtX, leftUpPtY),
				gocv.FontHersheyPlain, textFontScale, statusColor, textThickness,
			)
		} else {
			gocv.PutText(
				&img, "No markers detected", image.Pt(leftUpPtX, leftUpPtY),
				gocv.FontHersheyPlain, textFontScale, errorColor, textThickness,
			)
		}

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}

	fmt.Println("---------------")
	fmt.Println("PROGRAM STOP")
}

func getMapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func getCenterMarker(markerCorners []gocv.Point2f) (x float32, y float32) {
	for _, corner := range markerCorners {
		x += corner.X
		y += corner.Y
	}
	x = x / 4
	y = y / 4

	return
}

func getCenterImg(img gocv.Mat) (x float32, y float32) {
	imgWidth := img.Cols()
	imgHeight := img.Rows()
	x = float32(imgWidth) / 2
	y = float32(imgHeight) / 2

	return
}
