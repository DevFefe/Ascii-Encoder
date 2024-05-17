package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/png"
	"os"
	"time"

	"golang.org/x/term"
)

// Convert image file to image.Image obj
func decodeImage(filename string) (image.Image, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

// Convert image file to image.Image obj
func decodeGif(filename string) (*gif.GIF, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	images, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func calcBounds(value gif.GIF) (int, int, int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return 0, 0, 0, 0
	}

	maxWidth := 0
	maxHeight := 0
	// Iterate through all frames to find the maximum width and height
	for _, frame := range value.Image {
		bounds := frame.Bounds()
		width := bounds.Max.X
		height := bounds.Max.Y

		if width > maxWidth {
			maxWidth = width
		}
		if height > maxHeight {
			maxHeight = height
		}
	}

	// Calculate width_diff and height_diff using the maximum dimensions
	width_diff := maxWidth / height / 2
	height_diff := maxHeight / height

	return width, height, width_diff, height_diff
}

func main() {

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Not a terminal")
		return
	}

	// Convert Image to Ascii

	// image, _, err := decodeImage("./gif.gif")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	gif, err := decodeGif("./gif.gif")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	asciiChars := []rune{'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\''}

	var width, height, width_diff, height_diff int

	index := 0
	for {
		width, height, width_diff, height_diff = calcBounds(*gif)

		var buffer bytes.Buffer

		buffer.WriteString("\033[H")

		for y := range height {
			for x := range width {
				new_x := x*width_diff - width_diff/2
				new_y := y*height_diff - height_diff/2

				if new_x > gif.Image[index%len(gif.Image)].Bounds().Max.X || new_y > gif.Image[index%len(gif.Image)].Bounds().Max.Y ||
					new_x < gif.Image[index%len(gif.Image)].Bounds().Min.X || new_y < gif.Image[index%len(gif.Image)].Bounds().Min.Y {
					buffer.WriteRune(asciiChars[len(asciiChars)-1])

				} else {
					R, G, B, _ := gif.Image[index%len(gif.Image)].At(new_x, new_y).RGBA()
					y := 0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)
					buffer.WriteRune(asciiChars[int(y)%len(asciiChars)])
				}
			}
			if y != height-1 {
				buffer.WriteString("\n")
			}
		}

		fmt.Print(buffer.String())

		time.Sleep(time.Duration(gif.Delay[index%len(gif.Image)]) * 10 * time.Millisecond)

		index += 1
	}
}
