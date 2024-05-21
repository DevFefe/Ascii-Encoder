package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"strings"
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
				R, G, B, A := frame.At(new_x, new_y).RGBA()
				if A != 0 {
					R = R * 255 / A
					G = G * 255 / A
					B = B * 255 / A
				}

				Y := 0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)

				buffer.WriteString("\x1b[38;2;" + fmt.Sprintf("%d", R) + ";" + fmt.Sprintf("%d", G) + ";" + fmt.Sprintf("%d", B) + "m" + string(asciiChars[int(Y)%len(asciiChars)]) + "\x1b[0m")
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

	video_filename := flag.String("video", "", "")
	flag.Parse()

	if string(*video_filename) == "" {
		fmt.Print("Specify video using -video {filename}\n")
		return
	}

	if file_info, err := os.Stat(*video_filename); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exists\n", *video_filename)
		return
	} else {
		if strings.Split(file_info.Name(), ".")[1] != "mp4" {
			fmt.Printf("File %s is not a mp4 video\n", *video_filename)
			return
		}
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

	video, err := vidio.NewVideo(string(*video_filename))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(img.Pix)

	index := 0
	for video.Read() {
		if index == 0 {
			index++
		} else {
			index = 0
			continue
		}

		startTime := time.Now()

		frame := video.FrameBuffer()
		png.Encode(bytes.NewBuffer(frame), img)
		fmt.Print(convertImage(img, asciiChars))

		elapsed := time.Since(startTime)
		sleepTime := time.Second/time.Duration(video.FPS()/2) - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}
}
