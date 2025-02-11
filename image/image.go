package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// To determine whether a color is dark or light, you can calculate the
// perceived brightness of the color using the following formula:
// Brightness=0.2126R+0.7152G+0.0722B where R, G, and B are the red, green, and
// blue values of the color, respectively. If the brightness value is greater
// than or equal to 128, the color is considered light. If the brightness value
// is less than 128, the color is considered dark.

// uint32 to RGB => / 257

var COMMANDS = []string{"create", "analyze", "color"}

const (
	FullBlock = "█"
	Reset     = "\033[m"
)

func main() {
	cli()
}

func cli() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "error: no command given\nUse one of %s\n",
			strings.Join(COMMANDS, ", "))
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: no output filename given\n")
			os.Exit(1)
		}
		out := os.Args[2]
		create_image(out)
	case "analyze":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "error: no input filename given\n")
			os.Exit(1)
		}
		filename := os.Args[2]
		if err := analyze_image(filename); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "color":
		if len(os.Args) < 5 {
			fmt.Fprintf(os.Stderr, "enter 3 r g b values\nUsage: color <R> <G> <B>\n")
			os.Exit(1)
		}

		red, err := strToByte(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		green, err := strToByte(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		blue, err := strToByte(os.Args[4])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		color := color.RGBA{R: red, G: green, B: blue, A: 255}
		brightness := 0.2126*float64(red) + 0.7152*float64(green) + 0.0722*float64(blue)
		var brightStr string
		if brightness >= 128.0 {
			brightStr = "light"
		} else {
			brightStr = "dark"
		}

		fmt.Printf("%d %d %d -> %.1f (%s)\n", red, green, blue, brightness, brightStr)

		fmt.Printf("%s%s%s\n", colored(color), strings.Repeat(FullBlock, 11), Reset)
		fmt.Printf("%s%s%s\n", colored(color), strings.Repeat(FullBlock, 11), Reset)
		fmt.Printf("%s%s%s\n", colored(color), strings.Repeat(FullBlock, 11), Reset)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %s\nUse one of %s\n",
			os.Args[1], strings.Join(COMMANDS, ", "))
		os.Exit(1)
	}
}

func strToByte(colorStr string) (byte, error) {
	red, err := strconv.ParseUint(colorStr, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("could not convert to a number %s: %v\n", os.Args[2], err)
	}
	return byte(red), nil
}

func analyze_image(filename string) error {
	f, err := OpenFile(filename)
	if err != nil {
		return err
	}
	img, err := decode_image(f)
	if err != nil {
		log.Fatal("error decoding image")
	}

	colors := make(map[color.RGBA]int)

	rect := img.Bounds()
	width := rect.Max.X - rect.Min.X
	height := rect.Max.Y - rect.Min.Y

	for y := range height {
		for x := range width {
			c := img.At(x, y)
			rgba, ok := c.(color.RGBA)
			if ok {
				colors[rgba]++
			}
		}
	}

	redish := []color.RGBA{}
	greenish := []color.RGBA{}
	blueish := []color.RGBA{}
	grayish := []color.RGBA{}

	for color := range colors {
		switch getColorType(color) {
		case Redish:
			redish = append(redish, color)
		case Greenish:
			greenish = append(greenish, color)
		case Blueish:
			blueish = append(blueish, color)
		case Grayish:
			grayish = append(grayish, color)
		default:
			panic("unreachable")
		}
	}

	printColor := func(color color.RGBA) {
		c := colored(color)
		fmt.Printf("%s%s%s -> %d %d %d\n",
			c, strings.Repeat(FullBlock, 3), Reset,
			color.R, color.G, color.B)
	}

	fmt.Println("\nREDISH")
	for _, color := range redish {
		printColor(color)
	}

	fmt.Println("\nGREEISH")
	for _, color := range greenish {
		printColor(color)
	}

	fmt.Println("\nBLUEISH")
	for _, color := range blueish {
		printColor(color)
	}

	fmt.Println("\nGRAYISH")
	for _, color := range grayish {
		printColor(color)
	}

	return nil
}

type ColorType int

const (
	Redish ColorType = iota
	Greenish
	Blueish
	Grayish
)

func getColorType(color color.RGBA) ColorType {
	if isRedish(color) {
		return Redish
	}

	if isGreenish(color) {
		return Greenish
	}

	if isBlueish(color) {
		return Blueish
	}

	return Grayish
}

func isRedish(color color.RGBA) bool {
	// if 255-color.G < 20 || 255-color.B < 20 {
	// 	return false
	// }
	// return color.R >= 100 && color.R >= color.G+14 && color.R >= color.B+14
	return color.R > color.G && color.R > color.B
}

func isGreenish(color color.RGBA) bool {
	// if 255-color.R < 20 || 255-color.B < 20 {
	// 	return false
	// }
	// return color.G >= 100 && color.G >= color.R+14 && color.G >= color.B+14
	return color.G > color.R && color.G > color.B
}

func isBlueish(color color.RGBA) bool {
	// if 255-color.R < 20 || 255-color.G < 20 {
	// 	return false
	// }
	return color.B > color.R && color.B > color.G
}

func colored(color color.RGBA) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.R, color.G, color.B)
}

func decode_image(r io.Reader) (image.Image, error) {
	image, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("error: decoding image: %v", err)
	}
	return image, nil
}

func create_image(out string) {
	width, height := 10, 6

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Define some colors
	colors := []color.RGBA{
		{255, 0, 0, 255},     // Red
		{0, 255, 0, 255},     // Green
		{0, 0, 255, 255},     // Blue
		{128, 128, 128, 255}, // Gray
		{0, 0, 0, 255},       // Black
		{255, 255, 255, 255}, // White
	}

	for y := range height {
		color := colors[0]
		for x := range width {
			img.SetRGBA(x, y, color)
		}
	}

	filename := fmt.Sprintf("pngs/%s", out)
	if !strings.HasSuffix(out, ".png") {
		filename += ".png"
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		panic(err)
	}

	fmt.Printf("✅ PNG file '%s' created successfully!", filename)
}

func OpenFile(filename string) (*os.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filename, err)
	}
	return f, nil
}
