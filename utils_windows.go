package rotate_writer

import (
	"os"
	"syscall"
)

func pickFileCreateTime(info os.FileInfo) (second int64) {
	wFileSys := info.Sys().(*syscall.Win32FileAttributeData)
	tNanSeconds := wFileSys.CreationTime.Nanoseconds() // 返回的是纳秒
	tSec := tNanSeconds / 1e9                          ///秒
	return tSec
}
