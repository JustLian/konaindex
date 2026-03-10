package utils

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sort"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"golang.org/x/image/draw"
)

// -- Color temperature

func sRGBToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func CalculateTemperature(r, g, b float64) float64 {
	linR := sRGBToLinear(r / 255.0)
	linG := sRGBToLinear(g / 255.0)
	linB := sRGBToLinear(b / 255.0)

	X := linR*0.4124564 + linG*0.3575761 + linB*0.1804375
	Y := linR*0.2126729 + linG*0.7151522 + linB*0.0721750
	Z := linR*0.0193339 + linG*0.1191920 + linB*0.9503041

	if X+Y+Z == 0 {
		return 0
	}

	x := X / (X + Y + Z)
	y := Y / (X + Y + Z)

	n := (x - 0.3320) / (0.1858 - y)
	cct := 449*math.Pow(n, 3) + 3525*math.Pow(n, 2) + 6823.3*n + 5520.33

	return cct
}

// -- Colors clustering

type LAB struct {
	L, A, B float64
}

type ColorResult struct {
	Color  LAB
	Weight float64
}

type ImageInfo struct {
	Palette     []ColorResult
	Temperature float64
}

func GetImageInfo(img image.Image, paletteSize int) (*ImageInfo, error) {
	bounds := img.Bounds()
	thumbRect := image.Rect(0, 0, 150, 150)
	thumb := image.NewRGBA(thumbRect)

	draw.BiLinear.Scale(thumb, thumbRect, img, bounds, draw.Over, nil)

	var observations clusters.Observations
	var sumR, sumB, sumG float64
	pixelCount := float64(thumbRect.Dx() * thumbRect.Dy())

	for y := 0; y < thumbRect.Dy(); y++ {
		for x := 0; x < thumbRect.Dx(); x++ {

			r, g, b, _ := thumb.At(x, y).RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(b>>8)

			sumR += r8
			sumG += g8
			sumB += b8

			// converting to LAB
			c := colorful.Color{
				R: r8 / 255.0,
				G: g8 / 255.0,
				B: b8 / 255.0,
			}
			l, a, b_val := c.Lab()
			observations = append(observations, clusters.Coordinates{l, a, b_val})

		}
	}

	avgR, avgG, avgB := sumR/pixelCount, sumG/pixelCount, sumB/pixelCount
	temp := CalculateTemperature(avgR, avgG, avgB)

	km := kmeans.New()
	clst, err := km.Partition(observations, paletteSize)
	if err != nil {
		return nil, err
	}

	sort.Slice(clst, func(i, j int) bool {
		return len(clst[i].Observations) > len(clst[j].Observations)
	})

	var palette []ColorResult
	for _, cl := range clst {
		cent := cl.Center

		clusterWeight := float64(len(cl.Observations)) / pixelCount

		// boosting cluster weight if saturation is high
		chroma := math.Sqrt(math.Pow(cent[1], 2) + math.Pow(cent[2], 2))

		if chroma > 40.0 {
			clusterWeight = clusterWeight * 1.5
		}

		if clusterWeight > 1.0 {
			clusterWeight = 1.0
		}

		palette = append(palette, ColorResult{
			Color: LAB{
				L: float64(cent[0]),
				A: float64(cent[1]),
				B: float64(cent[2]),
			},
			Weight: clusterWeight,
		})
	}

	return &ImageInfo{
		Palette:     palette,
		Temperature: temp,
	}, nil
}
