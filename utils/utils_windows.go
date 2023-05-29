//go:build windows
// +build windows

package utils

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/Trisia/gosysproxy"
	// "golang.org/x/sys/windows"
	// "golang.org/x/sys/windows/registry"
)

var (
	// reg_addr = "HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run"
	reg_addr = "HKCU\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion"
)

// 打开文件夹
func OpenFolder(file string) {
	fpath := strings.ReplaceAll(file, "/", "\\")
	exec.Command("explorer", "/select,", fpath).CombinedOutput()
}

// 检查是否开机启动
func CheckStartup() bool {
	out, _ := Exec("reg", "query", reg_addr+"\\Run")
	return strings.Contains(out, "clashweb")
}

// 设置 Clash 开机启动
func SetClashStartup(enable bool) bool {
	var e error
	if enable {
		exe, _ := os.Executable()
		Exec("reg", "add", reg_addr+"\\Run", "/v", "clashweb", "/t", "REG_SZ", "/d", `"`+exe+`"`, "/f")
		// e = exec.Command("reg", "add", reg_addr, "/v", "clashweb", "/t", "REG_SZ", "/d", `"`+exe+`"`, "/f").Run()
		// if runElevated("reg", "add "+reg_addr+" /t REG_SZ /d \""+exe+"\" /f") {
		//     return true
		// }
	} else {
		Exec("reg", "delete", reg_addr+"\\Run", "/v", "clashweb", "/f")
		// e = exec.Command("reg", "delete", reg_addr, "/v", "clashweb", "/f").Run()
		// if runElevated("reg", "delete "+reg_addr+" /v clashweb /f") {
		//     return true
		// }
	}
	if e != nil {
		Msg(e.Error())
		return false
	}
	return true
}

// 设置系统代理 https://mrjun.cn/gSrhcL14a/ https://github.com/Trisia/gosysproxy
func SetSystemProxy(enable bool) bool {
	var err error
	if enable {
		err = gosysproxy.SetGlobalProxy("127.0.0.1:" + strconv.Itoa(ClashPort))
	} else {
		err = gosysproxy.Off()
	}

	if err != nil {
		Msg(err.Error())
		return false
	}
	return true
}

// 获取系统代理
func GetSystemProxy() {
	out, _ := Exec("reg", "query", reg_addr+"\\Internet Settings")
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "ProxyEnable") {
			continue
		}
		if strings.HasPrefix(l, "ProxyEnable") && (l[len(l)-1] == '1') && (len(ClashScheme) > 0) {
			if u, e := url.Parse(ClashScheme + "://127.0.0.1:" + strconv.Itoa(ClashPort)); e == nil {
				HttpClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}}
			} else {
				Msg(e.Error())
			}
		} else {
			HttpClient = &http.Client{}
		}
		break
	}
}

// https://stackoverflow.com/questions/31558066/how-to-ask-for-administer-privileges-on-windows-with-go
// 提权运行Command
// func RunElevated(exe, args string) bool {
// 	verb := "runas"
// 	// cwd, _ := os.Getwd()
// 	// args := strings.Join(os.Args[1:], " ")

// 	verbPtr, _ := syscall.UTF16PtrFromString(verb)
// 	exePtr, _ := syscall.UTF16PtrFromString(exe)
// 	cwdPtr, _ := syscall.UTF16PtrFromString(CurrentWorkDir)
// 	argPtr, _ := syscall.UTF16PtrFromString(args)

// 	var showCmd int32 = 1 //SW_NORMAL

// 	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
// 	if err != nil {
// 		Msg(err.Error())
// 		return false
// 	}
// 	return true
// }

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}

func Msg(msg string) {
	c := exec.Command("msg", UserName, msg)
	c.SysProcAttr = GetSysProcAttr()
	c.Run()
}
