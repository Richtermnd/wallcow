package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TODO: make it args
var (
	fontFile string
	output   string
	height   int
	width    int
	bgColor  color.Color
	fgColor  color.Color

	//go:embed CaskaydiaMonoNerdFontMono-Regular.ttf
	defaultFont []byte
)

func init() {
	var resolution string
	var fgHex, bgHex string
	flag.StringVar(&output, "o", "wallcow_output.png", "output file")
	flag.StringVar(&fontFile, "font", "default", "font file (now supports only ttf)")
	flag.StringVar(&resolution, "resolution", "1920x1080", "output image resolution WIDTHxHEIGHT")
	flag.StringVar(&fgHex, "fg", "e1e1e1ff", "hex font color")
	flag.StringVar(&bgHex, "bg", "1e1e1eff", "hex background color")
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

func fatalf(s string, args ...any) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(1)
}

func fill(im draw.Image) {
	draw.Draw(im, im.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
}

func newImage() draw.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

func readFont() []byte {
	if fontFile == "default" {
		return defaultFont
	}
	data, err := os.ReadFile(fontFile)
	if err != nil {
		fatalf("failed to read font file %s\n", fontFile)
	}
	return data
}

func getCowsayOutput() string {
	fortune := exec.Command("fortune")
	cowsay := exec.Command("cowsay")

	buf := new(bytes.Buffer)
	pipe, err := fortune.StdoutPipe()
	if err != nil {
		fatalf("failed to get pipe from fortune: %v\n", err)
	}

	cowsay.Stdin = pipe // pass 'fortune' stdout to 'cowsay' stdin
	cowsay.Stdout = buf // get output of cowsay into buffer
	cowsay.Start()      // start cowsay, it wait for input
	fortune.Run()       // run fortune, content of stdout will be redirected as stdin for cowsay
	cowsay.Wait()       // wait until cowsay finished his process

	return buf.String()
}

func renderText(im draw.Image, text string) {
	splittedText := strings.Split(text, "\n")
	textHeight := len(splittedText)
	textWidth := len(splittedText[0])

	var fontSize int
	// vertical gap - 100px
	// Looks weird
	// TODO: make it customizable
	if (height-200)/textHeight*10/12 < (width-300)/textWidth {
		fontSize = (height - 200) / textHeight * 10 / 12
	} else {
		fontSize = (width - 300) / textWidth
	}

	fmt.Printf("fontSize:    %d\n", fontSize)

	f, err := truetype.Parse(readFont())
	if err != nil {
		fatalf("failed to parse font: %v\n", err)
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: float64(fontSize),
	})
	d := font.Drawer{
		Dst:  im,
		Face: face,
		Src:  &image.Uniform{fgColor},
	}

	x := fixed.I((width - textWidth*fontSize*3/5) / 2)
	yOffset := (height - textHeight*(fontSize*12/10)) / 2
	for y, v := range splittedText {
		d.Dot = fixed.Point26_6{
			X: x,
			Y: fixed.I(yOffset + y*int(float64(fontSize)*1.2)),
		}
		d.DrawString(v)
	}
}

func saveImage(im image.Image) {
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
	im := newImage()
	fill(im)
	renderText(im, getCowsayOutput())
	saveImage(im)
}
