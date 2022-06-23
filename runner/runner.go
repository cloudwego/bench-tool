/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runner

import (
	"log"
	"sync"
	"time"

	"github.com/cloudwego/bench-tool/stats"
)

// 单次测试
type RunOnce func() error

type Runner struct {
	counter *stats.Counter // 计数器
	timer   *stats.Timer   // 计时器
}

func NewRunner() *Runner {
	r := &Runner{
		counter: stats.NewCounter(),
		timer:   stats.NewTimer(time.Microsecond),
	}
	return r
}

func (r *Runner) benching(runOnce RunOnce, concurrent int, total int64) stats.Report {
	r.counter.Reset(total)
	start := r.timer.Now()

	var wg sync.WaitGroup
	wg.Add(concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()
			for {
				idx := r.counter.Idx()
				if idx >= total {
					return
				}
				begin := r.timer.Now()
				err := runOnce()
				end := r.timer.Now()
				if err != nil {
					log.Printf("No.%d request failed: %v", idx, err)
				}
				cost := end - begin
				r.counter.AddRecord(idx, err, cost)
			}
		}()
	}
	wg.Wait()

	duration := r.timer.Now() - start
	return r.counter.Report(duration)
}

func (r *Runner) Run(onceFn RunOnce, concurrent int, total int64) stats.Report {
	return r.benching(onceFn, concurrent, total)
}
