//go:build linux
// +build linux

package utils

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	// "github.com/0xAX/notificator"
	// "github.com/gen2brain/beeep"
	// "gopkg.in/toast.v1"
)

// 打开文件夹 https://unix.stackexchange.com/questions/364997/open-a-directory-in-the-default-file-manager-and-select-a-file
func OpenFolder(file string) {
	Exec("dbus-send", "--session", "--print-reply", "--dest=org.freedesktop.FileManager1", "--type=method_call", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1.ShowItems", `array:string:"`+file+`"`, `string:""`)
}

// 检查是否开机启动
func CheckStartup() bool {
	return FileExists("/home/" + UserName + "/.config/autostart/clashweb.desktop")
}

// func CheckStartup() bool {
// 	// return FileExists("/home/" + UserName + "/.config/systemd/user/default.target.wants/clashweb.service")
// 	if _, e := exec.LookPath("crontab"); e != nil {
// 		Msg(e.Error())
// 		return false
// 	}

// 	o, _ := Exec("crontab", "-l")
// 	for _, entry := range strings.Split(o, "\n") {
// 		if strings.HasPrefix(entry, "@reboot") {
// 			continue
// 		}
// 		if strings.Contains(entry, "clashweb") {
// 			return true
// 		}
// 	}
// 	return false
// }

// 设置 Clash 开机启动 https://blog.lovefan.club/2020/09/18/22.html
func SetClashStartup(enable bool) bool {
	if enable {
		str := `[Desktop Entry]
Name=ClashWeb
Exec=` + CurrentWorkDir + `/clashweb
Type=Application`

		if e := os.MkdirAll("/home/"+UserName+"/.config/autostart", os.ModePerm); e != nil {
			Msg(e.Error())
			return false
		}

		if e := ioutil.WriteFile("/home/"+UserName+"/.config/autostart/clashweb.desktop", []byte(str), os.ModePerm); e != nil {
			return false
		}
	} else {
		if _, e := Exec("rm", "-rf", "/home/"+UserName+"/.config/autostart/clashweb.desktop"); e != nil {
			return false
		}
	}
	return true
}

// 设置 Clash 开机启动 https://stackoverflow.com/questions/878600/how-to-create-a-cron-job-using-bash-automatically-without-the-interactive-editor
// func SetClashStartup(enable bool) bool { // 返回bool表示操作是否成功，不表示开机启动的状态
// 	str := `@reboot sleep 5 && ` + CurrentWorkDir + "/clashweb"

// 	cronTmpFile := CurrentWorkDir + "/.cron"
// 	defer Exec("rm", "-rf", cronTmpFile)
// 	if o, e := Exec("crontab", "-l"); e == nil {
// 		if e := ioutil.WriteFile(cronTmpFile, []byte(o), os.ModePerm); e != nil {
// 			Msg(e.Error())
// 			return false
// 		}
// 	} else {
// 		return false
// 	}

// 	if enable {
// 		if _, e := Exec("sed", "-i", "1i "+str, cronTmpFile); e != nil { // 在第1行插入
// 			return false
// 		}
// 	} else {
// 		str = strings.ReplaceAll(str, "/", "\\/") // \也需要转义
// 		if _, e := Exec("sed", "-i", "/"+str+"/d", cronTmpFile); e != nil {
// 			return false
// 		}
// 	}

// 	// _, e := Exec("/bin/sh", "-c", "sudo crontab "+cronTmpFile)
// 	_, e := Exec("crontab", cronTmpFile)
// 	return e == nil
// }

// linux and mac 设置系统代理 https://mrjun.cn/Dy4ALDFDO/
func SetSystemProxy(enable bool) bool {
	proxies := []string{"http", "https", "socks"}
	port := strconv.Itoa(ClashConfig.MixedPort)

	var e error
	for _, proxy := range proxies {
		if enable {
			e = exec.Command("gsettings", "set", "org.gnome.system.proxy."+proxy, "host", "127.0.0.1").Run()
			e = exec.Command("gsettings", "set", "org.gnome.system.proxy."+proxy, "port", port).Run()
		} else {
			e = exec.Command("gsettings", "set", "org.gnome.system.proxy."+proxy, "host", "").Run()
			e = exec.Command("gsettings", "set", "org.gnome.system.proxy."+proxy, "port", "").Run()
		}
	}

	if enable {
		e = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	} else {
		e = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	}

	if e != nil {
		Msg("Set System Proxy Error: " + e.Error())
		return false
	}
	return true
}

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// githbub.com beeep
// https://wiki.archlinux.org/title/Desktop_notifications
func Msg(msg string) {
	c := exec.Command("gdbus", "call", "--session", "--dest=org.freedesktop.Notifications", "--object-path=/org/freedesktop/Notifications", "--method=org.freedesktop.Notifications.Notify", "", "0", CurrentWorkDir+"/icon/enable.png", "ClashWeb", msg, "[]", `{"urgency":<1>,"sound-name":<"default">}`, "0")
	c.SysProcAttr = GetSysProcAttr()
	c.Run()
}
