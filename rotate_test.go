package rotate_writer

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWriter(t *testing.T) {
	p := NewTimePolicy(time.Second*5, time.Second*5*6)

	w := NewWriter("E:/tmp/tmp/name.log", p)

	var wg sync.WaitGroup
	runnerCnt := 1

	wg.Add(runnerCnt)
	for i := 0; i < runnerCnt; i++ {
		go func(idx int) {
			defer wg.Done()
			tk := time.NewTicker(time.Second)
			tm := time.NewTimer(time.Second * 30)
			for {
				select {
				case <-tk.C:
					_, err := w.Write([]byte(fmt.Sprintf("%d:  %s\n", idx, time.Now().Format(time.RFC3339))))
					if err != nil {
						t.Error(err)
					}
					fmt.Printf("%d: write done\n", idx)
				case <-tm.C:
					return
				}
			}
		}(i)
	}

	wg.Wait()
}
