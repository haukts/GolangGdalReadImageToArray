package main

import (
	"fmt"
	"github.com/lukeroth/gdal"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
)

var wg sync.WaitGroup

func main() {
	ds, err := gdal.Open("test.img", gdal.ReadOnly)
	if err == nil {
		fmt.Println("img read success!")
		x := ds.RasterXSize()
		y := ds.RasterYSize()
		b := ds.RasterCount()
		fmt.Println("band numbers", b)
		fmt.Println("Image width", x)
		fmt.Println("Image height", y)

		img := make([][][]float64, b) //储存影像数据的三维数组
		for i := 0; i < b; i++ {
			img[i] = make([][]float64, x)
		}

		for i := 0; i < b; i++ {
			for j := 0; j < x; j++ {
				img[i][j] = make([]float64, y)
			}
		}

		minMax := make([][]float64, b) //储存每个波段DN值的最大值与最小值

		wg.Add(b)

		for i := 0; i < b; i++ {

			minMax[i] = make([]float64, 2)
			band := ds.RasterBand(i + 1)
			go ReadDataFromBand(band, x, y, minMax[i], img[i])
		}

		wg.Wait()

		//程序运行到这里的时候影像数据已经读入到三维数组中了

		CreateImage(x, y, img, minMax) //将读入的数据输出为png

	}

}

func ReadDataFromBand(band gdal.RasterBand, x, y int, minMax []float64, img [][]float64) {
	p := make([]float64, x*y, x*y)
	band.IO(gdal.Read, 0, 0, x, y, p, x, y, 0, 0) //从raster里取数据到缓冲区
	minMax[0], _ = band.GetMinimum()
	minMax[1], _ = band.GetMaximum()
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			img[i][j] = float64(p[i+j*x])
		}
	}
	wg.Done()
}

func CreateImage(x, y int, imageArry [][][]float64, minMax [][]float64) {
	width := x
	height := y

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			/*
				R:=int(float64(imageArry[0][x][y]-minMax[0][0])/float64(minMax[0][1]-minMax[0][0])*float64(255))
				G:=int(float64(imageArry[1][x][y]-minMax[1][0])/float64(minMax[1][1]-minMax[1][0])*float64(255))
				B:=int(float64(imageArry[2][x][y]-minMax[2][0])/float64(minMax[2][1]-minMax[2][0])*float64(255))
				//线性拉伸
			*/

			color := color.RGBA{uint8(imageArry[0][x][y]), uint8(imageArry[1][x][y]), uint8(imageArry[2][x][y]), 0xff}
			img.Set(x, y, color)
		}
	}
	// Encode as PNG.
	f, _ := os.Create("image.png")
	png.Encode(f, img)
}
