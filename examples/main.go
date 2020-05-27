/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 14:11:05
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 23:35:27
 */
package main

import (
	"fmt"
	"time"

	"github.com/monitor1379/simplecache"
)

func example01_SetGetDel() {
	cache := simplecache.New()
	cache.Set("k1", "v1", 0)

	value, ok := cache.Get("k1")
	fmt.Println(value, ok)
	// print: v1 true

	ok = cache.Del("k1")
	fmt.Println(ok)
	// print: true

	ok = cache.Del("k1")
	fmt.Println(ok)
	// print: false
}

func example02_Expire() {
	cache := simplecache.New()
	cache.Set("k1", "v1", time.Millisecond*500) // expire 500ms

	value, ok := cache.Get("k1")
	fmt.Println(value, ok)
	// print: v1 true

	time.Sleep(time.Millisecond * 500)
	value, ok = cache.Get("k1")
	fmt.Println(value, ok)
	// print: <nil> false
}

func example03_MaxMemory() {
	var err error

	// default options: see github.com/monitor1379/simplecache/options.go::defaultOptions
	cache := simplecache.NewMemCacheWithOptions(simplecache.Options{
		IntervalOfProactivelyDeleteExpiredKey: time.Second * 1,
		MaxMemoryPolicyType:                   simplecache.MaxMemoryPolicyTypeNoeviction,
	})
	cache.SetMaxMemory("500B")

	value := make([]byte, 1024) // 1KB
	err = cache.Set("k1", value, 0)
	fmt.Println(err)
	// print: "out of max memory"

	cache.SetMaxMemory("1.2KB")
	err = cache.Set("k1", value, 1*time.Second)
	fmt.Println(err)
	// print: <nil>

	err = cache.Set("k2", value, 0)
	fmt.Println(err)
	// print: "out of max memory"

	time.Sleep(1 * time.Second)

	err = cache.Set("k2", value, 0)
	fmt.Println(err)
	// print: <nil>

}

func example04_Others() {
	cache := simplecache.New()
	cache.Set("k1", "v1", 0)
	cache.Set("k2", "v2", 0)

	fmt.Println(cache.Keys())
	// print: 2

	cache.Flush()

	fmt.Println(cache.Keys())
	// print: 0
}

func main() {
	// example01_SetGetDel()
	// example02_Expire()
	// example03_MaxMemory()
	example04_Others()
}
