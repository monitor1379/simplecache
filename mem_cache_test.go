package simplecache_test

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 14:25:58
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 14:35:05
 */

import (
	"testing"

	"github.com/monitor1379/simplecache"
)

func BenchmarkMemCache(b *testing.B) {
	mc := simplecache.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Set("0123456789", i, 0)
	}
}
