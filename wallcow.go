package wallcow

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/CaskaydiaMonoNerdFontMono-Regular.ttf
var defaultFont []byte

func fatalf(s string, args ...any) {
	fmt.Fprintf(os.Stderr, s, args...)
	os.Exit(1)
}

func fill(im draw.Image, fillColor color.Color) {
	draw.Draw(im, im.Bounds(), &image.Uniform{fillColor}, image.Point{}, draw.Src)
}

func newImage(width, height int) draw.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

func renderText(im draw.Image, text string, textColor color.Color, rawFont []byte) {
	if rawFont == nil {
		rawFont = defaultFont
	}
	fmt.Printf("text: %v\n", text)
	splittedText := strings.Split(text, "\n")
	width, height := im.Bounds().Max.X, im.Bounds().Max.Y
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

	f, err := truetype.Parse(rawFont)
	if err != nil {
		fatalf("failed to parse font: %v\n", err)
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: float64(fontSize),
	})

	d := font.Drawer{
		Dst:  im,
		Face: face,
		Src:  &image.Uniform{textColor},
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

func pipeCmds(cmds []*exec.Cmd) {
	for i := 0; i < len(cmds)-1; i++ {
		cmd := cmds[i]
		nextCmd := cmds[i+1]

		pipe, err := cmd.StdoutPipe()
		if err != nil {
			fatalf("failed to get pipe: %v\n", err)
		}
		nextCmd.Stdin = pipe
	}
}

func getOutput(command string) string {
	pipeSeparated := strings.Split(command, "|")
	cmds := make([]*exec.Cmd, 0, len(pipeSeparated))
	for _, rawCmd := range pipeSeparated {
		splitted := strings.Split(strings.TrimSpace(rawCmd), " ")
		cmd := exec.Command(strings.TrimSpace(splitted[0]), splitted[1:]...)
		cmds = append(cmds, cmd)
	}

	pipeCmds(cmds)
	w := new(bytes.Buffer)
	cmds[len(cmds)-1].Stdout = w

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fatalf("failed to start %s: %v\n", cmd.Path, err)
		}
	}

	cmds[len(cmds)-1].Wait()

	return w.String()
}

// Generate generate image with rendered cmd output
// api is bullshit, but I don't care
func Generate(
	cmd string,
	width, height int,
	bgColor color.Color,
	fgColor color.Color,
	rawFont []byte,
) image.Image {
	var text string
	if cmd == "" {
		b, _ := io.ReadAll(os.Stdin)
		text = string(b)
	} else {
		text = getOutput(cmd)
	}
	im := newImage(width, height)
	fill(im, bgColor)
	renderText(im, text, fgColor, rawFont)
	return im
}
