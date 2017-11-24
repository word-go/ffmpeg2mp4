package main

import (
	"os"
	"fmt"
	"time"
	"github.com/word-go/ffmpeg2mp4/transition"
)

func main() {

	intFile, err := os.Open("test.mp4")
	if err != nil {
		fmt.Print(err)
	}
	tt := transition.NewFunc(intFile)
	tt.SetDeBug(false)
	has1name, err := tt.GetOutFileName("mp4")
	if err != nil {
		fmt.Print(err)
	}
	outFile, err := os.Create(has1name)
	if err != nil {
		fmt.Print(err)
	}
	go func() {
		if err := tt.Mp4(outFile); err != nil {
			fmt.Sprintf("转码错误：%v\n", err)
		}
	}()

forSelect:
	for {
		select {
		case <-tt.Status:
			fmt.Println("完事")
			break forSelect
		default:
			time.Sleep(time.Millisecond * 100)
			fmt.Printf("当前进度：%.2f%%\r", float32(tt.CurrentTime)/float32(tt.Duration)*100)
		}
	}

}
