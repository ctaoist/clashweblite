package main

import (
	"github.com/ctaoist/clashweb/icon"
	"github.com/ctaoist/clashweb/utils"

	"os"
	"os/exec"

	// "fmt"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
	// "github.com/tidwall/gjson"
)

var (
	clash           *exec.Cmd
	isClashRunning  bool // clash 运行状态
	clashCurVersion string
	clashWebVersion = "1.0.0"

	systemProxyStatus bool // after restarting clash whether set SystemProxy
	// rkey registry.Key
)

func main() {
	// var err error
	// rkey, err = registry.OpenKey(registry.LOCAL_MACHINE, "", registry.ALL_ACCESS)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	startClash()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTooltip("ClashWeb")
	mStartup := systray.AddMenuItemCheckbox("Startup", "Startup ClashWebLite", false)
	mSystemProxy := systray.AddMenuItemCheckbox("System Proxy", "System Proxy", false)

	systray.AddSeparator()
	mOpenConfig := systray.AddMenuItem("Edit Clash config.yaml", "Open Clash config.yaml folder")
	mOpenUserConfig := systray.AddMenuItem("Edit Clash user.yaml", "Open Clash user.yaml folder")
	mStopClash := systray.AddMenuItem("Stop Clash", "Stop or Start ClashWeb")
	mUpdate := systray.AddMenuItem("Update", "Update Clash")

	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About", "About")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the ClashWeb")

	if utils.CheckStartup() {
		mStartup.Check()
	}
	if utils.SetSystemProxy(true) {
		mSystemProxy.Check()
		systemProxyStatus = true
	}

	go func() {
		for {
			select {
			case <-mStartup.ClickedCh: // 开机启动
				if mStartup.Checked() {
					if utils.SetClashStartup(false) {
						mStartup.Uncheck()
					} // disable startup
				} else {
					if utils.SetClashStartup(true) {
						mStartup.Check()
					}
				}
			case <-mSystemProxy.ClickedCh: // 系统代理
				if mSystemProxy.Checked() {
					if utils.SetSystemProxy(false) { // off system proxy
						mSystemProxy.Uncheck()
						systemProxyStatus = false
					}
				} else {
					if !isClashRunning {
						utils.Msg("Clash is not running, and setting system proxy failed!")
						break
					}
					if utils.SetSystemProxy(true) {
						mSystemProxy.Check()
						systemProxyStatus = true
					}
				}
			case <-mOpenConfig.ClickedCh:
				utils.OpenFolder(utils.ClashAppDir + "/config.yaml")
			case <-mOpenUserConfig.ClickedCh:
				utils.OpenFolder(utils.ClashAppDir + "/ruleset/user.yaml")
			case <-mStopClash.ClickedCh:
				if !isClashRunning {
					startClash()
					mStopClash.SetTitle("Stop Clash")
					if systemProxyStatus {
						if utils.SetSystemProxy(true) {
							mSystemProxy.Check()
						}
					}
					systray.SetIcon(icon.ReadIconFile(utils.CurrentWorkDir + "/icon/enable." + icon.IconSuffix))
				} else {
					stopClash()
					mStopClash.SetTitle("Start Clash")
					if utils.SetSystemProxy(false) {
						mSystemProxy.Uncheck()
					} // off system proxy
					systray.SetIcon(icon.ReadIconFile(utils.CurrentWorkDir + "/icon/disable." + icon.IconSuffix))
				}
			case <-mUpdate.ClickedCh:
				mUpdate.SetTitle("Updating...")
				mUpdate.Disable()
				utils.ClashUpdateCh = make(chan int) // 重新初始化，清空
				go checkClashUpdate()
				go func() {
					for {
						p := <-utils.ClashUpdateCh
						if p == 102 {
							mUpdate.SetTitle("Update")
							mUpdate.Enable()
							break
						}
						mUpdate.SetTitle("Updating... " + strconv.Itoa(p) + "%")
					}
				}()
			case <-mAbout.ClickedCh:
				utils.Msg("Clash Version: " + getCurClashVersion() + "\nClashWeb Version: " + clashWebVersion)
			case <-mQuitOrig.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	stopClash()
	utils.SetSystemProxy(false)
}

func execClash(args ...string) *exec.Cmd {
	c := exec.Command(utils.ClashAppDir+"/clash-"+runtime.GOOS+"-amd64", args...)
	c.SysProcAttr = utils.GetSysProcAttr()
	return c
}

func stopClash() {
	clash.Process.Kill()
	clash.Wait()
	isClashRunning = false
}

func startClash() {
	utils.ReadConfig()
	clash = execClash("-d", utils.ClashAppDir)
	clash.Start()
	isClashRunning = true
}

func getCurClashVersion() string {
	if clashCurVersion == "" {
		out, _ := execClash("-v").CombinedOutput()
		s := strings.TrimSpace(string(out))
		t, err := time.Parse("2 Jan 2006 15:4:5 PM", strings.TrimSpace(s[len(s)-27:len(s)-4])) // s = "Clash latest windows amd64 with go1.18 Tue 29 Mar 2022 07:47:38 AM UTC"
		if err != nil {
			utils.Msg(err.Error())
			return ""
		}
		clashCurVersion = t.Format("2006.01.02")
	}
	// if clashCurVersion == "" {
	// 	utils.Request("GET", "127.0.0.1:9090/version")
	// }
	return clashCurVersion
}

func checkClashUpdate() {
	var newVersion string
	defer func() { utils.ClashUpdateCh <- 102 }()

	getCurClashVersion()

	_, r := utils.Request("GET", "https://release.dreamacro.workers.dev/latest/", url.Values{},
		map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36",
		},
	)
	if len(r) == 0 {
		utils.Msg("获取最新版本失败")
		return
	}
	date := regexp.MustCompile("\\d{1,2}/\\d{1,2}/\\d{4}, \\d{1,2}:\\d{1,2}:\\d{1,2} [APM]{2}").FindStringSubmatch(string(r))[0]
	t, err := time.Parse("1/2/2006, 15:4:5", date[:len(date)-3])
	if err != nil {
		utils.Msg(err.Error())
		return
	}
	newVersion = t.Format("2006.01.02")

	if newVersion > clashCurVersion {
		clashZipName := utils.ClashAppDir + "/cache/clash-latest." + utils.DownZipSuffix
		clashExeName := utils.ClashAppDir + "/clash-" + runtime.GOOS + "-amd64"
		if downloadClash(clashZipName) {
			stopClash()
			if err := utils.DeCompress(utils.ClashAppDir, clashZipName, "clash-"+runtime.GOOS+"-amd64"); err != nil {
				utils.Msg(err.Error())
				return
			}
			// 设置权限 重命名 移除压缩包
			if runtime.GOOS != "windows" {
				// utils.Exec("mv", clashExeName+"-latest", clashExeName)
				utils.Exec("chmod", "a+x", clashExeName)
			}
			if err := os.Remove(clashZipName); err != nil {
				utils.Msg(err.Error())
			}
			startClash()
			clashCurVersion = newVersion
			utils.Msg("更新 Clash 成功")
		} else {
			utils.Msg("更新 Clash 失败")
		}
	} else {
		utils.Msg("Clash 已经是最新，不用更新")
	}
}

func downloadClash(clashFileName string) bool {
	if err := os.MkdirAll(utils.CurrentWorkDir+"/App/cache", os.ModePerm); err != nil {
		utils.Msg(err.Error())
		return false
	}

	if !utils.RequestToFile("GET", "https://release.dreamacro.workers.dev/latest/clash-"+runtime.GOOS+"-amd64-latest."+utils.DownZipSuffix, clashFileName, url.Values{},
		map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36",
		},
	) {
		return false
	}
	return true
}
