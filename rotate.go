package rotate_writer

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

type (
	Policy interface {
		NeedRotate(filename string, writeLen int) bool
		GenerateFilename(filename string, customTime ...time.Time) string
		DoClean(filename string) (notNeedLoop bool)
		GetOpenFileFlag() int
	}

	Writer struct {
		filename string
		realName string
		fd       *os.File
		locker   sync.Mutex
		policy   Policy
		once     sync.Once
	}
)

func NewWriter(filename string, policy Policy) *Writer {
	w := &Writer{
		filename: path.Clean(filename),
		policy:   policy,
	}
	if w.policy == nil {
		w.policy = defaultPolicy{}
	}

	return w
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		ticker := time.NewTicker(time.Second)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					for {
						<-ticker.C
						if w.policy.DoClean(w.filename) {
							break
						}
					}
				}
			}()
			for {
				<-ticker.C
				if w.policy.DoClean(w.filename) {
					break
				}
			}
		}()
	})
	w.locker.Lock()
	defer w.locker.Unlock()

	var openFileFlag = w.policy.GetOpenFileFlag()

	if w.fd == nil {
		fileInfo, err := os.Stat(w.filename)
		if err == nil {
			createTime := time.Unix(pickFileCreateTime(fileInfo), 0)
			_ = os.Rename(w.filename, w.policy.GenerateFilename(w.filename, createTime))
		}

		// 打开文件
		fd, err := w.openFile(w.filename, openFileFlag)
		if err != nil {
			return 0, err
		}
		w.fd = fd
		// 生成新文件名
		w.realName = w.policy.GenerateFilename(w.filename)
	}

	if !w.policy.NeedRotate(w.filename, len(p)) {
		return w.fd.Write(p)
	}

	_ = w.fd.Sync()
	// 关闭之前的fd
	err = w.fd.Close()
	if err != nil {
		return 0, err
	}
	// 文件重命名
	err = os.Rename(w.filename, w.realName)
	fmt.Println("rename:"+w.filename+"=>"+w.realName, err)
	if err != nil {
		return 0, err
	}
	// 创建新文件
	newFd, err := w.openFile(w.filename, os.O_CREATE|openFileFlag)
	if err != nil {
		return 0, err
	}
	w.fd = newFd
	// 生成新文件名
	w.realName = w.policy.GenerateFilename(w.filename)
	// 写入数据
	return w.fd.Write(p)
}

func (Writer) openFile(fn string, flag int) (*os.File, error) {
	dir := path.Dir(fn)
	_ = os.MkdirAll(dir, 0777)
	return os.OpenFile(fn, flag, 0666)
}

func (w Writer) Close() error {
	if w.fd == nil {
		return nil
	}
	_ = w.fd.Sync()
	return w.fd.Close()
}

type (
	defaultPolicy struct {
	}
)

func (d defaultPolicy) NeedRotate(filename string, writeLen int) bool {
	return false
}

func (d defaultPolicy) GenerateFilename(filename string, customTime ...time.Time) string {
	return filename
}

func (d defaultPolicy) DoClean(filename string) (notNeedLoop bool) {
	return true
}

func (d defaultPolicy) GetOpenFileFlag() int {
	return os.O_CREATE | os.O_APPEND | os.O_RDWR
}
