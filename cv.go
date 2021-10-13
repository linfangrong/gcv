// Copyright 2016 Evans. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gcv

import (
	"image"
	"image/color"

	"github.com/vcaesar/imgo"
	"gocv.io/x/gocv"
)

// FindImgFile find image file in subfile
func FindImgFile(tempFile, file string, flag ...int) (float32, float32, image.Point, image.Point) {
	return FindImgMatC(IMRead(file, flag...), IMRead(tempFile, flag...))
}

// IMRead read the image file to gocv.Mat
func IMRead(file string, flag ...int) gocv.Mat {
	f1 := 4
	if len(flag) > 0 {
		f1 = flag[0]
	}
	return gocv.IMRead(file, gocv.IMReadFlag(f1))
}

// IMWrite write the gocv.Mat to file
func IMWrite(name string, img gocv.Mat) bool {
	return gocv.IMWrite(name, img)
}

// IMWrite write the image.Image to file
func ImgWrite(name string, img image.Image) bool {
	im, err := ImgToMat(img)
	if err != nil {
		return false
	}

	return IMWrite(name, im)
}

// FindImg find image in the subImg
func FindImg(subImg, imgSource image.Image) (float32, float32, image.Point, image.Point) {
	m1, _ := ImgToMat(imgSource)
	m2, _ := ImgToMat(subImg)
	return FindImgMatC(m1, m2)
}

// FindImgByte find image in the subImg by []byte
func FindImgByte(subImg, imgSource []byte) (float32, float32, image.Point, image.Point) {
	m1, _ := imgo.ByteToImg(imgSource)
	m2, _ := imgo.ByteToImg(subImg)
	return FindImg(m2, m1)
}

// FindImgXY find image in the subImg return x, y
func FindImgXY(subImg, imgSource image.Image) (int, int) {
	_, _, _, maxLoc := FindImg(subImg, imgSource)
	return maxLoc.X, maxLoc.Y
}

// FindImgMatC find the image Mat in the temp Mat and close gocv.Mat
func FindImgMatC(imgSource, temp gocv.Mat) (float32, float32, image.Point, image.Point) {
	defer imgSource.Close()
	defer temp.Close()
	return FindImgMat(imgSource, temp)
}

// FindImgMat find the image Mat in the temp Mat
func FindImgMat(imgSource, temp gocv.Mat) (float32, float32, image.Point, image.Point) {
	res := gocv.NewMat()
	defer res.Close()
	msk := gocv.NewMat()
	defer msk.Close()

	gocv.MatchTemplate(imgSource, temp, &res, gocv.TmCcoeffNormed, msk)
	minVal, maxVal, minLoc, maxLoc := gocv.MinMaxLoc(res)

	return minVal, maxVal, minLoc, maxLoc
}

// ImgToMat trans image.Image to gocv.Mat
func ImgToMat(img image.Image) (gocv.Mat, error) {
	return gocv.ImageToMatRGB(img)
}

// MatToImg trans gocv.Mat to image.Image
func MatToImg(m1 gocv.Mat) (image.Image, error) {
	return m1.ToImage()
}

// Show show the gocv.Mat image
func Show(img gocv.Mat, name ...string) {
	wName := "show"
	if len(name) > 0 {
		wName = name[0]
	}
	window := gocv.NewWindow(wName)
	defer window.Close()

	window.ResizeWindow(1200, 800)
	window.IMShow(img)
	window.WaitKey(0)
}

// Result find template result structure
type Result struct {
	MiddlePoint []int
	TopLeft     []int
	Rectangles  [][]int

	Confidence []float32
	RangeVal   []int
}

// FindAllImg find the search image all template in the source image return []Result
func FindAllImg(imgSearch, imgSource image.Image, args ...interface{}) []Result {
	imSource, _ := ImgToMat(imgSource)
	imSearch, _ := ImgToMat(imgSearch)

	return FindAllTemplateC(imSource, imSearch, args...)
}

// FindAllImgFlie find the search image all template in the source image file
// return []Result
func FindAllImgFile(fileSearh, file string, args ...interface{}) []Result {
	return FindAllTemplateC(IMRead(file), IMRead(fileSearh), args...)
}

// FindMultiAllImg find the multi search image all template in the source image return []Result
func FindMultiAllImg(imgSearch []image.Image, imgSource image.Image, args ...interface{}) [][]Result {
	imSource, _ := ImgToMat(imgSource)
	var imSearch []gocv.Mat
	for i := 0; i < len(imgSearch); i++ {
		search, _ := ImgToMat(imgSearch[i])
		imSearch = append(imSearch, search)
	}

	return FindMultiAllTemplateC(imSource, imSearch, args...)
}

// FindMultiAllTemplate find the multi imgSearch all template in the imgSource return [][]Result
// and close gocv.Mat
func FindMultiAllTemplateC(imgSource gocv.Mat, imgSearch []gocv.Mat, args ...interface{}) (r [][]Result) {
	for i := 0; i < len(imgSearch); i++ {
		r = append(r, FindAllTemplateC(imgSource, imgSearch[i], args...))
	}

	return
}

// FindMultiAllTemplate find the multi imgSearch all template in the imgSource return [][]Result
func FindMultiAllTemplate(imgSource gocv.Mat, imgSearch []gocv.Mat, args ...interface{}) (r [][]Result) {
	for i := 0; i < len(imgSearch); i++ {
		r = append(r, FindAllTemplate(imgSource, imgSearch[i], args...))
	}

	return
}

// FindAllTemplate find the imgSearch all template in the imgSource return []Result
// and close gocv.Mat
func FindAllTemplateC(imgSource, imgSearch gocv.Mat, args ...interface{}) []Result {
	defer imgSource.Close()
	defer imgSearch.Close()
	return FindAllTemplate(imgSource, imgSearch, args...)
}

// FindAllTemplate find the imgSearch all template in the imgSource return []Result
func FindAllTemplate(imgSource, imgSearch gocv.Mat, args ...interface{}) []Result {
	threshold := float32(0.8)
	if len(args) > 0 {
		threshold = float32(args[0].(float64))
	}

	maxCount := 10
	if len(args) > 1 {
		maxCount = args[1].(int)
	}
	// rgb := false
	// if len(args) > 2 {
	// 	rgb = args[2].(bool)
	// }

	method := false
	if len(args) > 3 {
		method = args[3].(bool)
	}

	iGray := gocv.NewMat()
	defer iGray.Close()
	sGray := gocv.NewMat()
	defer sGray.Close()

	// if !rgb {
	gocv.CvtColor(imgSource, &iGray, gocv.ColorRGBToGray)
	gocv.CvtColor(imgSearch, &sGray, gocv.ColorRGBToGray)
	// }

	results := make([]Result, 0)
	for {
		_, maxVal, minLoc, maxLoc := FindImgMat(iGray, sGray)
		h, w := imgSearch.Rows(), imgSearch.Cols()
		if maxVal < threshold || len(results) > maxCount {
			break
		}

		if method {
			maxLoc = minLoc
		}

		rs := getVal(maxLoc, maxVal, w, h)
		results = append(results, rs)
		if len(results) <= 0 {
			return nil
		}

		Fill(iGray, rs.Rectangles)
		// Rectangle(iGray, maxLoc, w, h)
		// Show(iGray)
	}

	return results
}

func getVal(maxLoc image.Point, maxVal float32, w, h int) Result {
	rectangle := [][]int{
		{maxLoc.X, maxLoc.Y},
		{maxLoc.X, maxLoc.Y + h},
		{maxLoc.X + w, maxLoc.Y + h},
		{maxLoc.X + w, maxLoc.Y},
	}

	middle := image.Pt(maxLoc.X+w/2, maxLoc.Y+h/2)
	middlePoint := []int{middle.X, middle.Y}

	topLeft := []int{maxLoc.X, maxLoc.Y}
	maxVals := []float32{maxVal}
	rangeVal := []int{w, h}

	return Result{
		MiddlePoint: middlePoint,
		TopLeft:     topLeft,
		Rectangles:  rectangle,
		Confidence:  maxVals,
		RangeVal:    rangeVal,
	}
}

// Rectangle rectangle the iGray image
func Rectangle(iGray gocv.Mat, maxLoc image.Point, w, h int) {
	r := image.Rect(
		int(maxLoc.X-w/2),
		int(maxLoc.Y-h/2),
		int(maxLoc.X+w/2),
		int(maxLoc.Y+h/2),
	)

	// white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 0}
	gocv.Rectangle(&iGray, r, black, -1)
}

// Fill fillpoly the iGray image
func Fill(iGray gocv.Mat, rectangle [][]int) {
	pts := [][]image.Point{{
		image.Pt(rectangle[0][0], rectangle[0][1]),
		image.Pt(rectangle[1][0], rectangle[1][1]),
		image.Pt(rectangle[2][0], rectangle[2][1]),
		image.Pt(rectangle[3][0], rectangle[3][1]),
	}}

	pts1 := gocv.NewPointsVectorFromPoints(pts)
	defer pts1.Close()

	blue := color.RGBA{0, 0, 255, 0}
	gocv.FillPoly(&iGray, pts1, blue)
}
