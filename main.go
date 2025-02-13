package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// TODO: make it args
var (
	fontFile = "CaskaydiaMonoNerdFontMono-Regular.ttf"
	output   = "output.png"
	height   = 1080
	width    = 1920
	bgColor  = color.RGBA{30, 30, 30, 255}
	fgColor  = color.RGBA{230, 230, 230, 255}
)

func fatalf(s string, args ...any) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(1)
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

	// vertical gap - 100px
	// TODO: make it customizable
	var fontSize int
	if (height-200)/textHeight*10/12 < (width-300)/textWidth {
		fontSize = (height - 200) / textHeight * 10 / 12
	} else {
		fontSize = (width - 300) / textWidth
	}
	fmt.Printf("fontSize: %v\n", fontSize)

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

func readFont() []byte {
	data, err := os.ReadFile(fontFile)
	if err != nil {
		fatalf("failed to read font: %v\n", err)
	}
	return data
}

func newImage() draw.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

func fill(im draw.Image) {
	draw.Draw(im, im.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
}

func main() {
	im := newImage()
	fill(im)
	renderText(im, getCowsayOutput())
	saveImage(im)
}
