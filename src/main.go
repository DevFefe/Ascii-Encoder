package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"time"

	vidio "github.com/AlexEidt/Vidio"
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

func calcBounds(value image.Image) (int, int, int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return 0, 0, 0, 0
	}

	imageWidth := value.Bounds().Max.X
	imageHeight := value.Bounds().Max.Y

	// Calculate the aspect ratio of the terminal
	terminalAspectRatio := float64(width) / float64(height)
	// Calculate the aspect ratio of the image
	imageAspectRatio := float64(imageWidth) / float64(imageHeight)

	var adjustedWidth, adjustedHeight int

	if imageAspectRatio > terminalAspectRatio {
		// Image is wider than the terminal, fit width
		adjustedWidth = width
		adjustedHeight = int(float64(width) / imageAspectRatio)
	} else {
		// Image is taller than the terminal, fit height
		adjustedHeight = height
		adjustedWidth = int(float64(height) * imageAspectRatio)
	}

	width_diff := imageWidth / adjustedWidth / 2
	height_diff := imageHeight / adjustedHeight

	return width, height, width_diff, height_diff
}

func convertImage(frame image.Image, asciiChars []rune) string {
	width, height, width_diff, height_diff := calcBounds(frame)

	var buffer bytes.Buffer

	buffer.WriteString("\033[H")

	for y := range height {
		for x := range width {
			new_x := x*width_diff - width_diff/2
			new_y := y*height_diff - height_diff/2

			if new_x > frame.Bounds().Max.X || new_y > frame.Bounds().Max.Y ||
				new_x < frame.Bounds().Min.X || new_y < frame.Bounds().Min.Y {
				buffer.WriteRune(asciiChars[len(asciiChars)-1])

			} else {
				R, G, B, _ := frame.At(new_x, new_y).RGBA()
				Y := 0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)

				buffer.WriteRune(asciiChars[int(Y)%len(asciiChars)])
			}
		}
		if y != height-1 {
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}

func main() {

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Not a terminal")
		return
	}

	asciiChars := []rune{'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\''}

	// Convert Image to Ascii

	// image, _, err := decodeImage("./gif.gif")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// gif, err := decodeGif("./gif.gif")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// index := 0
	// for {
	// 	// frame := video.FrameBuffer()
	// 	// img, _, err := image.Decode(bytes.NewReader(frame))
	// 	// if err != nil {
	// 	// 	log.Fatalln(err)
	// 	// }

	// 	fmt.Print(convertFrame(gif.Image[index], asciiChars))

	// 	time.Sleep(time.Duration(time.Second / 15))

	// 	index += 1
	// }

	video, err := vidio.NewVideo("video.mp4")
	if err != nil {
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(img.Pix)

	for video.Read() {
		startTime := time.Now()

		frame := video.FrameBuffer()
		png.Encode(bytes.NewBuffer(frame), img)

		fmt.Print(convertImage(img, asciiChars))

		elapsed := time.Since(startTime)
		sleepTime := time.Second/time.Duration(video.FPS()) - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}
}
