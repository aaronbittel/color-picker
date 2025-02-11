package main

import (
	"fmt"
	"image/color"
	"math"
	"slices"
	"strings"
)

var (
	RED   RGB = RGB{255, 0, 0}
	GREEN RGB = RGB{0, 255, 0}
	BLUE  RGB = RGB{0, 0, 255}
)

type ColorSet map[RGB]struct{}

type RGBCountPair struct {
	rgb   RGB
	count int
}

func NewColorCount(rgbs []RGB) RGBCountPair {
	var sumRed, sumGreen, sumBlue float64
	for _, g := range rgbs {
		sumRed += float64(g.red)
		sumGreen += float64(g.green)
		sumBlue += float64(g.blue)
	}

	avgColor := RGB{
		red:   uint8(sumRed / float64(len(rgbs))),
		green: uint8(sumGreen / float64(len(rgbs))),
		blue:  uint8(sumBlue / float64(len(rgbs))),
	}

	return RGBCountPair{
		rgb:   avgColor,
		count: len(rgbs),
	}
}

func (cp RGBCountPair) redDiff() int {
	return 2*int(cp.rgb.red) - int(cp.rgb.green) - int(cp.rgb.blue)
}

func (cp RGBCountPair) greenDiff() int {
	return 2*int(cp.rgb.green) - int(cp.rgb.red) - int(cp.rgb.blue)
}

func (cp RGBCountPair) blueDiff() int {
	return 2*int(cp.rgb.blue) - int(cp.rgb.red) - int(cp.rgb.green)
}

type RGBColorPairSlice []RGBCountPair

func (cps *RGBColorPairSlice) sortByCount() {
	slices.SortFunc(*cps, func(a, b RGBCountPair) int {
		if a.count > b.count {
			return -1
		} else if a.count < b.count {
			return 1
		} else {
			return 0
		}
	})
}

func (cp *RGBColorPairSlice) sortByRed() {
	slices.SortFunc(*cp, func(a, b RGBCountPair) int {
		aRDiff := a.redDiff()
		bRDiff := b.redDiff()
		if aRDiff > bRDiff {
			return -1
		} else if aRDiff < bRDiff {
			return 1
		} else {
			return 0
		}
	})
}

func (cp *RGBColorPairSlice) sortByGreen() {
	slices.SortFunc(*cp, func(a, b RGBCountPair) int {
		aGDiff := a.greenDiff()
		bGDiff := b.greenDiff()
		if aGDiff > bGDiff {
			return -1
		} else if aGDiff < bGDiff {
			return 1
		} else {
			return 0
		}
	})
}

func (cp *RGBColorPairSlice) sortByBlue() {
	slices.SortFunc(*cp, func(a, b RGBCountPair) int {
		aBDiff := a.blueDiff()
		bBDiff := b.blueDiff()
		if aBDiff > bBDiff {
			return -1
		} else if aBDiff < bBDiff {
			return 1
		} else {
			return 0
		}
	})
}

type RGB struct {
	red   uint8
	green uint8
	blue  uint8
}

func fromColor(color color.Color) RGB {
	red, green, blue, _ := color.RGBA()
	return RGB{
		red:   uint8(red / 257),
		green: uint8(green / 257),
		blue:  uint8(blue / 257),
	}
}

func (rgb RGB) asHex() string {
	return fmt.Sprintf("#%X%X%X", rgb.red, rgb.green, rgb.blue)
}

func (rgb RGB) asFormattedRGB() string {
	return fmt.Sprintf("%3d-%3d-%3d", rgb.red, rgb.green, rgb.blue)
}

func (rgb RGB) printColor() {
	colorBlock := fmt.Sprintf("%s%s%s", colored(rgb), strings.Repeat(FullBlock, 5), Reset)
	fmt.Printf("%s -> %s (%s)\n", colorBlock, rgb.asFormattedRGB(), rgb.asHex())
}

func colored(rgb RGB) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", rgb.red, rgb.green, rgb.blue)
}

type ColorType int

const (
	Red ColorType = iota
	Green
	Blue
	Gray
)

const MAX_DIST = 217.0

func (rgb RGB) identify() ColorType {
	if dist(rgb, RED) <= MAX_DIST {
		return Red
	}

	if dist(rgb, GREEN) <= MAX_DIST {
		return Green
	}

	if dist(rgb, BLUE) <= MAX_DIST {
		return Blue
	}

	return Gray
}

func dist(rgb1, rgb2 RGB) float64 {
	return math.Sqrt(
		math.Pow(float64(rgb1.red)-float64(rgb2.red), 2.0) +
			math.Pow(float64(rgb1.green)-float64(rgb2.green), 2.0) +
			math.Pow(float64(rgb1.blue)-float64(rgb2.blue), 2.0),
	)
}
