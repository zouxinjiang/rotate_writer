package rotate_writer

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type (
	TimePolicy struct {
		ticker         *time.Ticker
		RotateInterval time.Duration
		FilenameFormat string
		MaxAge         time.Duration

		clock time.Time
	}
)

func NewTimePolicy(interval, maxAge time.Duration) *TimePolicy {
	p := &TimePolicy{
		RotateInterval: interval,
		FilenameFormat: "2006-01-02-15-04-05",
		MaxAge:         maxAge,
	}
	p.setDefault()

	return p
}

func (t *TimePolicy) setDefault() {
	if t.RotateInterval < time.Second {
		// 小于调度单位，则设为默认值
		t.RotateInterval = time.Hour
	}
	if t.MaxAge < 0 {
		t.MaxAge = time.Hour * 24
	}
	if t.FilenameFormat == "" {
		t.FilenameFormat = "2006-01-02-15-04-05"
	}
	if t.ticker == nil {
		t.ticker = time.NewTicker(t.RotateInterval)
	}

}

func (t *TimePolicy) NeedRotate(filename string, writeLen int) bool {
	t.setDefault()
	select {
	case <-t.ticker.C:
		return true
	default:
		return false
	}
}

func (t *TimePolicy) GenerateFilename(filename string, customTime ...time.Time) string {
	t.setDefault()

	name := filename + "_" + time.Now().Format(t.FilenameFormat)
	return name
}

func (t *TimePolicy) DoClean(filename string) (notNeedLoop bool) {
	t.setDefault()
	if t.MaxAge == 0 {
		// 不清理
		return true
	}

	dir := path.Dir(filename)
	name := path.Base(filename)
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	var matchedFiles = []struct {
		name     string
		createAt time.Time
	}{}

	// 找到目录下匹配的所有文件
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}
		if matched, _ := path.Match(name+"_*", info.Name()); matched {

			tmp := strings.Split(info.Name(), name+"_")
			if len(tmp) <= 1 {
				continue
			}

			var (
				createdAt time.Time
				err       error
			)
			createdAt, err = time.ParseInLocation(t.FilenameFormat, tmp[1], time.Local)
			if err != nil {
				continue
			}

			matchedFiles = append(matchedFiles, struct {
				name     string
				createAt time.Time
			}{name: info.Name(), createAt: createdAt})
		}
	}

	now := time.Now()

	// 删除过期的文件
	for _, item := range matchedFiles {
		if item.createAt.Add(t.MaxAge).Before(now) {
			// 删除
			_ = os.Remove(path.Join(dir, item.name))
		}
	}
	return
}

func (t *TimePolicy) GetOpenFileFlag() int {
	return os.O_RDWR | os.O_CREATE | os.O_APPEND
}
