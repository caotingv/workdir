package main

import (
	"fmt"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func main() {

	// server.Start()
	// client.Satrt()
	// example.ExampleStream("", "", false)

	data, err := ffmpeg.Probe("./file/img1.jpg")
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}
