package main

import (
	"flag"

	"fefe.ascii_encoder/lib"
)

func main() {
	video_filename := flag.String("video", "", "")
	flag.Parse()

	lib.AsciiEncodeFromVideoFile(string(*video_filename))
}
