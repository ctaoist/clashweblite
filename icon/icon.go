package icon

import (
	"runtime"

	"github.com/ctaoist/clashweb/utils"

	"io/ioutil"
)

var (
	Data       []byte
	IconSuffix string // "ico" or "png"
)

func init() {
	if runtime.GOOS == "linux" {
		IconSuffix = "png"
	} else {
		IconSuffix = "ico"
	}
	Data = ReadIconFile(utils.CurrentWorkDir + "/icon/enable." + IconSuffix)
}

func ReadIconFile(fname string) []byte {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		utils.Msg("err = " + err.Error())
	}
	return data
}
