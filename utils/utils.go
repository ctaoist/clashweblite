package utils

import (
	"runtime"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"

	// "fmt"
	// "os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"
	"net/url"
	// "syscall"

	"gopkg.in/yaml.v2"
)

type Downloader struct {
	io.Reader
	Total   int64
	Current int64
}

type ClashConf struct {
	MixedPort   int    `yaml:"mixed-port"`
	Port        int    `yaml:"port"`
	SocksPort   int    `yaml:"socks-port"`
	RedirPort   int    `yaml:"redir-port"`
	TproxyPort  int    `yaml:"tproxy-port"`
	BindAddress string `yaml:"bind-address"`
}

var (
	CurUser       *user.User
	UserName      string // system username
	HttpClient    = &http.Client{}
	DownZipSuffix string // "zip" or "gz"

	ClashUpdateCh  = make(chan int)
	ClashAppDir    string        // Clash Premium Dir
	ClashConfig    = ClashConf{} // Clash Premium config struct
	ClashPort      int           // 系统代理使用的端口
	CurrentWorkDir string        //
	ClashWebExe    string        // path_to_clashweb
	// rkey registry.Key
)

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	ClashUpdateCh <- int(d.Current * 100 / d.Total)
	// fmt.Println(d.Current*100/d.Total)
	// if d.Current == d.Total {
	//     ClashUpdateCh <- 100
	// }
	return
}

func init() {
	ClashWebExe, _ := os.Executable()
	CurrentWorkDir = filepath.Dir(ClashWebExe)
	ClashAppDir = CurrentWorkDir + "/App"

	CurUser, _ = user.Current()
	UserName = func(u string) string {
		u_lst := strings.Split(u, "\\")
		return u_lst[len(u_lst)-1]
	}(CurUser.Username)

	if runtime.GOOS == "windows" {
		DownZipSuffix = "zip"
	} else {
		DownZipSuffix = "gz"
	}
}

func ReadConfig() {
	b, e := ioutil.ReadFile(ClashAppDir + "/config.yaml")
	if e != nil {
		Msg(e.Error())
		panic(e)
	}
	e = yaml.Unmarshal(b, &ClashConfig)
	if e != nil {
		Msg(e.Error())
		panic(e)
	}
	if ClashConfig.MixedPort > 0 {
		ClashPort = ClashConfig.MixedPort
	} else if ClashConfig.Port > 0 {
		ClashPort = ClashConfig.Port
	} else if ClashConfig.SocksPort > 0 {
		ClashPort = ClashConfig.SocksPort
	} else if ClashConfig.RedirPort > 0 {
		ClashPort = ClashConfig.RedirPort
	} else if ClashConfig.TproxyPort > 0 {
		ClashPort = ClashConfig.TproxyPort
	} else {
		ClashPort = 80
	}
}

// 执行命令，如果出错并弹出程序输出
func Exec(exe string, args ...string) (string, error) {
	c := exec.Command(exe, args...)
	c.SysProcAttr = GetSysProcAttr()
	b, e := c.CombinedOutput()
	var s string
	if runtime.GOOS == "windows" {
		s = string(GbkToUtf8(b))
	} else {
		s = string(b)
	}
	if e != nil {
		Msg(s + "\n" + e.Error())
		return s, e
	}
	return s, nil
}

func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil || os.IsExist(err)
}

// https://www.jianshu.com/p/9ebb1c152000
// GBK 转 UTF-8
func GbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		Msg(e.Error())
		return nil
	}
	return d
}

func Request(method, url string, query url.Values, headers map[string]string) (int64, []byte) {
	req, _ := http.NewRequest(method, url, nil)
	req.URL.RawQuery = query.Encode()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if r, e := HttpClient.Do(req); e == nil {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Msg(err.Error())
		}
		return r.ContentLength, b
	} else {
		Msg(e.Error())
		return 0, []byte{}
	}

}

func RequestToFile(method, url, filename string, query url.Values, headers map[string]string) bool {
	req, _ := http.NewRequest(method, url, nil)
	req.URL.RawQuery = query.Encode()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// https://www.flysnow.org/2020/07/25/golang-file-size.html
	var s int64
	file, err := os.Stat(filename)
	if err == nil {
		s = file.Size()
	} else if os.IsNotExist(err) {
		s = 0
	} else {
		Msg(err.Error())
		return false
	}
	req.Header.Set("Range", "bytes="+strconv.FormatInt(s, 10)+"-")

	if r, e := HttpClient.Do(req); e == nil {
		defer r.Body.Close()

		var file *os.File
		var err error
		if _, ok := r.Header["Accept-Ranges"]; ok {
			file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND, os.ModePerm) //os.Create(filename)
		} else {
			s = 0
			file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, os.ModePerm) //os.Create(filename)
		}
		if err != nil {
			Msg(e.Error())
			return false
		}
		defer file.Close()

		d := &Downloader{
			Reader:  r.Body,
			Current: s,
			Total:   r.ContentLength,
		}
		if _, err := io.Copy(file, d); err != nil {
			Msg(err.Error())
			return false
		}
		return true
	} else {
		Msg(e.Error())
		return false
	}

}
