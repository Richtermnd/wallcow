package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/Richtermnd/wallcow"
)

var (
	fontFile string
	output   string
	height   int
	width    int
	bgColor  color.Color
	fgColor  color.Color
	cmd      string
)

func fatalf(s string, args ...any) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(1)
}

func init() {
	var resolution string
	var fgHex, bgHex string
	flag.StringVar(&output, "o", "wallcow_output.png", "output file")
	flag.StringVar(&fontFile, "font", "default", "font file (now supports only ttf)")
	flag.StringVar(&resolution, "resolution", "1920x1080", "output image resolution WIDTHxHEIGHT")
	flag.StringVar(&fgHex, "fg", "e1e1e1ff", "hex font color")
	flag.StringVar(&bgHex, "bg", "1e1e1eff", "hex background color")
	flag.StringVar(&cmd, "cmd", "", "command to render on image (default: read stdin)")
	flag.Parse()

	sw, sh, found := strings.Cut(resolution, "x")
	if !found {
		fatalf("invalid resolution %s\n", resolution)
	}
	width, _ = strconv.Atoi(sw)
	height, _ = strconv.Atoi(sh)
	fgColor = parseHexColor(fgHex)
	bgColor = parseHexColor(bgHex)
	fmt.Printf("output file: %s\n", output)
	fmt.Printf("font:        %s\n", fontFile)
	fmt.Printf("resolution:  %dx%d\n", width, height)
	fmt.Printf("fg color:    %s\n", fgHex)
	fmt.Printf("bg color:    %s\n", bgHex)
	fmt.Printf("cmd:         %s\n", cmd)
}

func parseHexColor(hexRepr string) color.Color {
	var c color.RGBA
	t, err := strconv.ParseUint(hexRepr, 16, 64)
	if err != nil {
		fatalf("failed to parse color %s: %v\n", hexRepr, err)
	}
	c.R = uint8((t >> 24) & 0xff)
	c.G = uint8((t >> 16) & 0xff)
	c.B = uint8((t >> 8) & 0xff)
	c.A = uint8((t >> 0) & 0xff)
	return c
}

func readFont(fontFile string) []byte {
	data, err := os.ReadFile(fontFile)
	if err != nil {
		fatalf("failed to read font file %s\n", fontFile)
	}
	return data
}

func saveImage(im image.Image, output string) {
	f, err := os.Create(output)
	if err != nil {
		fatalf("failed to create output file: %v\n", err)
	}
	if err := png.Encode(f, im); err != nil {
		fatalf("failed to save image as png: %v\n", err)
	}
	defer f.Close()
}

func main() {
	var rawFont []byte
	if fontFile == "default" {
		rawFont = nil
	} else {
		rawFont = readFont(fontFile)
	}
	im := wallcow.Generate(cmd, width, height, bgColor, fgColor, rawFont)
	saveImage(im, output)
}
