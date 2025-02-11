# Color Picker

This is a color picker written in Go. Currently, you can either use [Flameshot](https://flameshot.org/) to take a screenshot or provide a PNG file.
The output consists of all RGB colors found in the PNG, grouped by proximity to avoid overwhelming the results. 
Each color is displayed along with its RGB values and hex representation.

## Usage

```console
go install .
color-picker
```

```console
go run .
```

## Flags

- `--path`:    Provide a path to a PNG file to skip using Flameshot (Flameshot is not required in this case).
- `--sort`:    Sort the output by count, red, green, or blue.
- `--limit`:   Limit the number of colors displayed.
- `--verbose`: Print additional sorting information.
- `--help`:    Display help information.
  
## Exploration

This project also includes an exploration of the [png specification format](http://libpng.org/pub/png/spec/1.2/PNG-Contents.html).
