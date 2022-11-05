package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var inFile string = "./file/img1.jpg"

func Satrt() {
	conn, err := net.Dial("tcp", "127.0.0.1:8087")
	if err != nil {
		fmt.Printf("conn server failed, err:%v\n", err)
		return
	}
	defer conn.Close()

	pr, pw := io.Pipe()
	done := startFFmpegProcess(inFile, pw)

	connHandler(pr, conn)
	err = <-done
	if err != nil {
		panic(err)
	}

}

func connHandler(reader io.ReadCloser, conn net.Conn) {
	w, h := getVideoSize(inFile)
	log.Println(w, h)
	_, err := conn.Write([]byte(fmt.Sprintf("%d %d", w, h)))
	if err != nil {
		panic(fmt.Sprintf("write size error: %s", err))
	}

	for {
		buf := make([]byte, w*h)
		n, err := io.ReadFull(reader, buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("read error: %d, %s", n, err))

		}

		n, err = conn.Write(buf[:n])
		if err != nil {
			panic(fmt.Sprintf("write error: %d, %s", n, err))
		}
	}
	log.Println("数据传输完毕！")

}

func getVideoSize(fileName string) (int, int) {
	log.Println("Getting video size for", fileName)
	data, err := ffmpeg.Probe(fileName)
	if err != nil {
		panic(err)
	}
	// log.Println("got video info", data)
	type VideoInfo struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int
			Height    int
		} `json:"streams"`
	}
	vInfo := &VideoInfo{}
	err = json.Unmarshal([]byte(data), vInfo)
	if err != nil {
		panic(err)
	}
	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			return s.Width, s.Height
		}
	}
	return 0, 0
}

func startFFmpegProcess(infileName string, writer io.WriteCloser) <-chan error {
	log.Println("Starting ffmpeg process1")
	done := make(chan error)
	go func() {
		err := ffmpeg.Input(infileName).
			Output("pipe:",
				ffmpeg.KwArgs{
					"format": "rawvideo", "pix_fmt": "rgb24",
				}).
			WithOutput(writer).
			Run()
		log.Println("ffmpeg process1 done")
		_ = writer.Close()
		done <- err
		close(done)
	}()
	return done
}
