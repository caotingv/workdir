package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func Start() {
	listen, err := net.Listen("tcp", ":8087")
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return
	}
	defer listen.Close()

	for {
		// 等待客户端建立连接
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("accept failed, err:%v\n", err)
			continue
		}

		// 启动一个单独的 goroutine 去处理连接

		go connHandler(conn)

	}
}

func connHandler(c net.Conn) {
	pr, pw := io.Pipe()
	buf := make([]byte, 4096)

	// 获取图片的尺寸
	n, err := c.Read(buf)
	if err != nil {
		c.Close()
		return
	}
	fmt.Println("n-----------", n)
	imgSize := string(buf[:n])

	// 获取图片数据
	go func() {
		for {
			cnt, err := c.Read(buf)
			if cnt == 0 || err != nil {
				c.Close()
				break
			}

			// 数据写入管道
			_, err = pw.Write(buf)
			if err != nil {
				panic(fmt.Sprintf("write error: %s", err))
			}
		}
	}()
	nowTime := time.Now().Unix()
	outFile := fmt.Sprint("./file/out", nowTime, ".jpg")
	width, height := string2Int(imgSize)
	startFFmpegProcess(outFile, pr, width, height)
}

func startFFmpegProcess(outfileName string, buf io.Reader, width, height int) {
	log.Println("Starting ffmpeg process2")
	go func() {
		err := ffmpeg.Input("pipe:",
			ffmpeg.KwArgs{"format": "rawvideo",
				"pix_fmt": "rgb24", "s": fmt.Sprintf("%dx%d", width, height),
			}).
			Output(outfileName, ffmpeg.KwArgs{"pix_fmt": "yuv420p"}).
			OverWriteOutput().
			WithInput(buf).
			Run()
		log.Println("ffmpeg process2 done")
		if err != nil {
			log.Println("The image save failed :", err)
		}
	}()
}

func string2Int(size string) (width, height int) {
	arr := strings.Fields(size)
	width, err := strconv.Atoi(arr[0])
	if err != nil {
		log.Println(err)
	}
	height, err = strconv.Atoi(arr[1])

	if err != nil {
		log.Println(err)
	}
	return
}
