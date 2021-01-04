package rotate_writer

import (
	"os"
	"syscall"
)

func pickFileCreateTime(info os.FileInfo) (second int64) {
	stat_t := info.Sys().(*syscall.Stat_t)
	tCreate := int64(stat_t.Ctim.Sec)
	return tCreate
}
