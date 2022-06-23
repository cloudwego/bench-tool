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

package stats

import (
	"sync/atomic"
	"time"

	"github.com/montanaflynn/stats"
)

type Report struct {
	QPS         int64
	LatencyP99  float64 // ms
	LatencyP999 float64 // ms
}

// 计数器
type Counter struct {
	Total  int64     // 总调用次数(limiter)
	Failed int64     // 失败次数
	costs  []float64 // 耗时统计
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Reset(total int64) {
	c.Total = total
	c.Failed = 0
	c.costs = make([]float64, total)
}

func (c *Counter) AddRecord(idx int64, err error, cost int64) {
	c.costs[idx] = float64(cost)
	if err != nil {
		atomic.AddInt64(&c.Failed, 1)
	}
}

func (c *Counter) Idx() (idx int64) {
	idx = atomic.AddInt64(&c.Total, 1) - 1
	if idx < 0 {
		panic("counter index overflow")
	}
	return idx
}

func (c *Counter) Report(duration int64) Report {
	qps := float64(c.Total) / float64(duration) * float64(time.Second)

	costs := make([]float64, len(c.costs))
	for i := range c.costs {
		costs[i] = float64(c.costs[i])
	}
	tp99, _ := stats.Percentile(costs, 99)
	tp999, _ := stats.Percentile(costs, 99.9)

	return Report{
		QPS:         int64(qps),
		LatencyP99:  tp99 / float64(time.Millisecond),
		LatencyP999: tp999 / float64(time.Millisecond),
	}
}
