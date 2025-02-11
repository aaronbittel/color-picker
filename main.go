package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	FullBlock = "█"
	Reset     = "\033[m"
)

var SORT_BY = []string{"count", "red", "green", "blue"}

func main() {
	var (
		filepath string
		sortBy   string
		limit    int
		verbose  bool
	)

	sortUsage := fmt.Sprintf("sort pixels by {%s}", strings.Join(SORT_BY, ", "))

	flag.StringVar(&filepath, "path", "", "png file to analyze [optional]")
	flag.StringVar(&sortBy, "sort", "", sortUsage)
	flag.IntVar(&limit, "limit", 0, "limit output count")
	flag.BoolVar(&verbose, "verbose", false, "extra sort values")
	flag.Parse()

	var err error

	if filepath == "" {
		filepath, err = flameshot()
		if err != nil {
			log.Fatal(err)
		}
	}

	colors, err := getPixels(filepath)
	if err != nil {
		log.Fatal(err)
	}

	groupedColors := groupSimilarColors(colors)
	groupedColors = sort(groupedColors, sortBy)
	groupedColors = limitSlice(groupedColors, limit)

	for _, cc := range groupedColors {
		if verbose {
			printSortInfo(cc, sortBy)
		}
		cc.rgb.printColor()
	}
}

func colorPrint(color RGB, msg string) {
	fmt.Println(colored(color), msg, Reset)
}

func flameshot() (string, error) {
	now := time.Now().Format("2006-01-02_15-04-05") + ".png"
	filepath := path.Join(os.TempDir(), now)
	flameshot := exec.Command("flameshot", "gui", "-s", "-p", filepath)
	if err := flameshot.Run(); err != nil {
		return "", err
	}

	return filepath, nil
}

func getPixels(filepath string) (map[RGB]struct{}, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	rect := img.Bounds()
	startX, startY, endX, endY := rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y

	colorSet := make(map[RGB]struct{})

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			color := fromColor(img.At(x, y))
			colorSet[color] = struct{}{}
		}
	}

	return colorSet, nil
}

func groupSimilarColors(colorSet map[RGB]struct{}) RGBColorPairSlice {
	groupedColors := make([]RGBCountPair, 0, len(colorSet)/2)
	visited := make([]bool, len(colorSet), len(colorSet))
	colors := []RGB{}
	for c := range colorSet {
		colors = append(colors, c)
	}

	for i := range colors {
		if visited[i] {
			continue
		}
		grouped := []RGB{}
		for j := range colors {
			if i == j {
				continue
			}

			if visited[j] {
				continue
			}

			d := dist(colors[i], colors[j])
			if d <= 15 {
				grouped = append(grouped, colors[j])
				visited[j] = true
			}
		}

		if len(grouped) > 0 {
			rgbCount := NewColorCount(grouped)
			groupedColors = append(groupedColors, rgbCount)
		}
	}

	return RGBColorPairSlice(groupedColors)
}

func sort(groupedColors RGBColorPairSlice, sortBy string) RGBColorPairSlice {
	switch sortBy {
	case "count":
		groupedColors.sortByCount()
	case "red":
		groupedColors.sortByRed()
	case "green":
		groupedColors.sortByGreen()
	case "blue":
		groupedColors.sortByBlue()
	case "":
	default:
		fmt.Printf("unknown sort by %s\n", sortBy)
	}

	return groupedColors
}

func limitSlice(groupedColors RGBColorPairSlice, limit int) RGBColorPairSlice {
	if limit > 0 {
		limit = min(limit, len(groupedColors))
		groupedColors = groupedColors[:limit]
	}
	return groupedColors
}

func printSortInfo(cc RGBCountPair, sortBy string) {
	switch sortBy {
	case "count":
		fmt.Printf("count: %d\t", cc.count)
	case "red":
		fmt.Printf("redDiff: %d\t", cc.redDiff())
	case "green":
		fmt.Printf("greenDiff: %d\t", cc.greenDiff())
	case "blue":
		fmt.Printf("blueDiff: %d\t", cc.blueDiff())
	}
}
