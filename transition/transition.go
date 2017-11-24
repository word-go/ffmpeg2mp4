package transition

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Transition struct {
	IntFile     *os.File  //写入文件
	OutFile     *os.File  //输出文件
	Duration    int       //视频总时长
	CurrentTime int       //当前转码时长
	Status      chan bool //转码状态
	DeBug       bool      //是否开启debug
}

func NewFunc(intFile *os.File) *Transition {
	t := new(Transition)
	t.IntFile = intFile
	t.Status = make(chan bool)
	t.DeBug = true
	return t
}

/**
Mp4转码
*/
func (t *Transition) Mp4(outFile *os.File) error {
	//开始执行命令
	cmd := exec.Command("ffmpeg", "-i", t.IntFile.Name(), "-vcodec", "libx264", "-y", outFile.Name())
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderrPipe.Close()
	if err := cmd.Start(); err != nil {
		return err
	}
	reader := bufio.NewReader(stderrPipe)
	if t.DeBug {
		fmt.Printf("转码开始：\n")
	}
	for {
		line, err := reader.ReadBytes('\r')
		if err != nil || err == io.EOF {
			break
		}
		//匹配视频时长
		reg1 := regexp.MustCompile(`Duration:(.*?),`)
		snatch1 := reg1.FindStringSubmatch(string(line))
		if len(snatch1) > 1 {
			t.Duration = timeEncode(snatch1[1])
			if t.DeBug {
				fmt.Printf("视频Duration: %d 秒\n", t.Duration)
			}
		}
		//匹配视频转码进度时间
		reg2 := regexp.MustCompile(`frame=(.*?)fps=(.*?)q=(.*?)size=(.*?)time=(.*?)bitrate=`)
		snatch2 := reg2.FindStringSubmatch(string(line))
		if len(snatch2) > 5 {
			t.CurrentTime = timeEncode(snatch2[5])
			if t.DeBug {
				fmt.Printf("转码进度：%.2f%% \r", float32(t.CurrentTime)/float32(t.Duration)*100)
			}
			if t.DeBug && t.Duration == t.CurrentTime {
				fmt.Printf("\n")
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	t.Status <- true
	if t.DeBug {
		fmt.Printf("转码完成\n")
	}

	return nil
}

/**
*获取输出文件名
 */
func (t *Transition) GetOutFileName(suf string) (string, error) {
	h := sha1.New()
	_, err := io.Copy(h, t.IntFile)
	if err != nil {
		return fmt.Sprintf("123.%s", suf), err
	}
	return fmt.Sprintf("%x.%s", h.Sum(nil), suf), nil
}

/**
开发调试信息
*/
func (t *Transition) SetDeBug(b bool) *Transition {
	t.DeBug = b
	return t
}

/**
时间解析
*/
func timeEncode(t string) int {
	time := strings.Trim(t, " ")
	hour, _ := strconv.Atoi(time[:2])
	minute, _ := strconv.Atoi(time[3:5])
	second, _ := strconv.Atoi(time[6:8])
	return second + minute*60 + hour*3600
}
