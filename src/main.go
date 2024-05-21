package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"fefe.ascii_encoder/lib"
)

func isValidYouTubeURL(videoURL string) bool {
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return false
	}
	if parsedURL.Host != "www.youtube.com" && parsedURL.Host != "youtu.be" {
		return false
	}
	return true
}

func main() {
	filename := flag.String("video", "", "Path to the video file")
	url := flag.String("url", "", "YouTube video URL")
	flag.Parse()

	if *filename == "" && *url == "" {
		fmt.Print("Specify video using -video {filename} or -url {youtube_url}\n")
		return
	}

	if *filename != "" && *url != "" {
		fmt.Print("Specify only one of -video or -url, not both\n")
		return
	}

	if *filename != "" {
		if fileInfo, err := os.Stat(*filename); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("File %s does not exist\n", *filename)
			return
		} else {
			if strings.Split(fileInfo.Name(), ".")[1] != "mp4" {
				fmt.Printf("File %s is not an mp4 video\n", *filename)
				return
			}
		}
	}

	if *url != "" {
		if !isValidYouTubeURL(*url) {
			fmt.Printf("URL %s is not a valid YouTube URL\n", *url)
			return
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\nBYE BYE") // TODO: rm temp video stored
	}()

	lib.AsciiEncodeFromVideoFile(*filename)
}
