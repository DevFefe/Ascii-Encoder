package lib

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"golang.org/x/term"
)

func AsciiEncodeFromVideoFile(filename string) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Not a terminal")
		return
	}

	if string(filename) == "" {
		fmt.Print("Specify video using -video {filename}\n")
		return
	}

	if file_info, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exists\n", filename)
		return
	} else {
		if strings.Split(file_info.Name(), ".")[1] != "mp4" {
			fmt.Printf("File %s is not a mp4 video\n", filename)
			return
		}
	}

	asciiChars := []rune{'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\''}

	video, err := vidio.NewVideo(string(filename))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, video.Width(), video.Height()))
	video.SetFrameBuffer(img.Pix)

	frame_count := 0
	for video.Read() {
		startTime := time.Now()

		frame := video.FrameBuffer()
		png.Encode(bytes.NewBuffer(frame), img)
		fmt.Print(convertImage(img, asciiChars))

		elapsed_time := float64(frame_count) / float64(video.Frames()) * video.Duration()
		remaining_time := video.Duration() - elapsed_time
		fmt.Print("\033[2K\r")
		fmt.Printf("%s\t%d/%d Frames", formatTime(int(remaining_time)), frame_count, video.Frames())

		elapsed := time.Since(startTime)
		sleepTime := time.Second/time.Duration(video.FPS()) - elapsed
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
		frame_count++
	}
}

func calcBounds(value image.Image) (int, int, int, int) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		return 0, 0, 0, 0
	}

	imageWidth := value.Bounds().Max.X
	imageHeight := value.Bounds().Max.Y

	terminalAspectRatio := float64(width) / float64(height)
	imageAspectRatio := float64(imageWidth) / float64(imageHeight)

	var adjustedWidth, adjustedHeight int

	if imageAspectRatio > terminalAspectRatio {
		adjustedWidth = width
		adjustedHeight = int(float64(width) / imageAspectRatio)
	} else {
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

	for y := range height - 1 {
		for x := range width {
			new_x := x*width_diff - width_diff/2
			new_y := y*height_diff - height_diff/2

			if new_x > frame.Bounds().Max.X || new_y > frame.Bounds().Max.Y ||
				new_x < frame.Bounds().Min.X || new_y < frame.Bounds().Min.Y {
				buffer.WriteRune(' ')
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

func formatTime(seconds int) string {
	minutes := seconds / 60
	seconds = seconds % 60
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
