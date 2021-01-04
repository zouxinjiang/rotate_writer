package rotate_writer

import (
	"fmt"
	"os"
	"time"
)

type (
	SizePolicy struct {
		MaxSizeByte  int64
		MaxSaveCount int64
		counter      uint64
	}
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

func NewSizePolicy(maxSizeByte, maxSaveCount int64) *SizePolicy {
	p := &SizePolicy{
		MaxSizeByte:  maxSizeByte,
		MaxSaveCount: maxSaveCount,
	}
	p.setDefault()

	return p
}

func (p *SizePolicy) setDefault() {
	if p.MaxSizeByte <= 0 {
		p.MaxSizeByte = 10 * MB
	}
	if p.MaxSaveCount <= 0 {
		p.MaxSaveCount = 5
	}

	if p.counter > uint64(p.MaxSaveCount) {
		p.counter = p.counter % uint64(p.MaxSaveCount)
	}

}

func (p *SizePolicy) NeedRotate(filename string, writeLen int) bool {
	p.setDefault()
	fileInfo, err := os.Stat(filename)
	if err != nil {
		// 获取信息失败
		return true
	}
	if fileInfo.IsDir() {
		return true
	}
	if fileInfo.Size()+int64(writeLen) < p.MaxSizeByte {
		return false
	}
	return true
}

func (p *SizePolicy) GenerateFilename(filename string, customTime ...time.Time) string {
	cnt := p.counter
	p.counter = (p.counter + 1) % uint64(p.MaxSaveCount)
	return fmt.Sprintf("%s.%d", filename, cnt)
}

func (p SizePolicy) DoClean(filename string) (notNeedLoop bool) {
	return true
}

func (p SizePolicy) GetOpenFileFlag() int {
	return os.O_CREATE | os.O_RDWR | os.O_TRUNC
}
