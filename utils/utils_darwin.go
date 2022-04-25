//go:build darwin
// +build darwin

package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// 打开文件夹
func OpenFolder(file string) {
	e := exec.Command("open", "-R", file).Run()
	if e != nil {
		Msg(e.Error())
	}
}

// 检查是否开机启动
func CheckStartup() bool {
	return FileExists("/Users/" + UserName + "/Library/LaunchAgents/com.ctaoist.clashweb.plist")
}

// 设置 Clash 开机启动
func SetClashStartup(enable bool) bool {
	if enable {
		str := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>KeepAlive</key>
	<true/>
	<key>Label</key>
	<string>com.ctaoist.clashweb</string>
	<key>ProgramArguments</key>
	<array>
		<string>` + CurrentWorkDir + "/clashweb" + `</string>
	</array>
	<key>StandardErrorPath</key>
	<string>/tmp/com.ctaoist.clashweb.err</string>
	<key>StandardOutPath</key>
	<string>/tmp/com.ctaoist.clashweb.out</string>
</dict>
</plist>`

		if e := ioutil.WriteFile("/Users/"+UserName+"/Library/LaunchAgents/com.ctaoist.clashweb.plist", []byte(str), os.ModePerm); e != nil {
			Msg(e.Error())
			return false
		}
		// f, e := os.OpenFile("/Users/"+UserName+"/Library/LaunchAgents/com.ctaoist.clashweb.plist", os.O_CREATE|os.O_RDWR, os.ModePerm)
		// defer f.Close()
		// if e != nil {
		// 	Msg(e.Error())
		// 	return false
		// }
		// f.WriteString(str)
	} else {
		if e := exec.Command("rm", "-rf", "/Users/"+UserName+"/Library/LaunchAgents/com.ctaoist.clashweb.plist").Run(); e != nil {
			return false
		}
	}
	return true
}

// https://mrjun.cn/Dy4ALDFDO/
func SetSystemProxy(enable bool) bool {
	o, e := exec.Command("networksetup", "-listallnetworkservices").CombinedOutput()
	if e != nil {
		Msg(e.Error())
		return false
	}
	ignoreNetService := []string{"Serial", "Iphone", "Ipad", "Bluetooth", "*"}
	proxies := []string{"-setwebproxy", "-setsecurewebproxy", "-setsocksfirewallproxy"}
	port := strconv.Itoa(ClashConfig.MixedPort)

	for _, networkservice := range strings.Split(string(o), "\n") {
		if func() bool {
			for _, ignore := range ignoreNetService {
				if strings.Contains(networkservice, ignore) {
					return true
				}
			}
			return false
		}() {
			continue
		}
		if len(strings.TrimSpace(networkservice)) < 2 {
			continue
		}

		var e error
		for _, proxy := range proxies {
			if enable {
				e = exec.Command("networksetup", proxy, networkservice, "127.0.0.1", port).Run()
			} else {
				e = exec.Command("networksetup", proxy, networkservice, "", "").Run()
				e = exec.Command("networksetup", proxy+"state", networkservice, "off").Run()
			}
		}
		if e != nil {
			Msg("Set System Proxy for " + networkservice + " Error : " + e.Error())
			return false
		}
	}
	return true
}

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

//https://github.com/martinlindhe/notify
//https://blog.csdn.net/arbboter/article/details/40621967 msgbox
//https://code-maven.com/display-notification-from-the-mac-command-line
//github.com/gen2brain/beeep
// Msg 通知消息
func Msg(msg string) {
	c := exec.Command("osascript", "-e", `display notification "`+msg+`" with title "ClashWeb" sound name "Default"`)
	c.SysProcAttr = GetSysProcAttr()
	c.Run()
}
