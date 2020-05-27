package simplecache_test

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 14:25:58
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-28 00:35:14
 */

import (
	"testing"
	"time"

	"github.com/monitor1379/simplecache"
)

func BenchmarkMemCache_Set(b *testing.B) {
	var err error
	mc := simplecache.NewWithOptions(simplecache.Options{
		IntervalOfProactivelyDeleteExpiredKey: 1 * time.Second,
	})

	mc.SetMaxMemory("1GB")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = mc.Set("0123456789", i, 0)
		if err != nil {
			b.Errorf("unexpected err: %s\n", err)
		}
	}
}

func BenchmarkMap_Set(b *testing.B) {
	m := make(map[string]interface{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m["0123456789"] = i
	}
}
