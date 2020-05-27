# simplecache

Golang implementation of Memory Cache.

Features:
- 接口使用简单。
- 支持设置过期时间。采取Lazy Delete + 定期扫描的方式，高效删除过期键值。
- 支持设置最大可用内存。内存淘汰策略目前仅支持`no-eviction`，即当内存占用超过最大限制时，插入操作会返回失败。后续逐步支持:
    - `MaxMemoryPolicyTypeAllKeysRandom`: 从所有key中随机删除一个。
	- `MaxMemoryPolicyTypeVolatileRandom`: 从设置了过期时间的key中随机删除一个。
    - `MaxMemoryPolicyTypeAllKeysLRU`: 从所有key中按照LRU删除最近最少使用的一个。
	- `MaxMemoryPolicyTypeVolatileLRU`: 从设置了过期时间的key中按照按照LRU删除最近最少使用的一个。

## Installation 

```bash
go get -u -v github.com/monitor1379/simplecache
```


## Benchmark

```bash
cd ${GOPATH}/src/github.com/monitor1379/simplecache
go test -v -bench=. -run=none -benchmem mem_cache_test.go   
```

与内建`map`的插入操作比较:
```
goos: linux
goarch: amd64
BenchmarkMemCache_Set-8          9634270               113 ns/op              40 B/op          2 allocs/op
BenchmarkMap_Set-8              43507314                26.2 ns/op             8 B/op          1 allocs/op
```


## Examples


```go
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
	example01_SetGetDel()
	example02_Expire()
	example03_MaxMemory()
	example04_Others()
}

```



# Development

## 1. 数据结构

`Cache`接口的实现为`MemCache`:
```go

type MemCache struct {
	options Options

	// 单位: bytes
	maxMemory   int64
	memoryUsage int64

	// 存储所有key-entry pair
	mu    sync.RWMutex
	table map[string]*Entry

	// 存储所有设置了expire的key-entry pair，用于主动定期清理，所以用普通锁而不是读写锁
	expiredMu    sync.Mutex
	expiredTable map[string]*Entry
}
```


`MemCache.table`为一个`map[string]*Entry`，其中，`Entry`定义为:

```go

type Entry struct {
	value       interface{}
	expiredNano int64 // 过期时间戳。0表示永不过期
	valueSize   int64 // sizeof(value)
}

```

通过读写锁`sync.RWMutex`来实现对`MemCache`读写操作的并发安全。



## 2. 过期时间的实现

最简单的实现就是在`Set()`操作时，给需要检查过期时间的key加上一个`time.Ticker`，起一个goroutine去后台清理。大致的伪代码是:

```go

import "time"

func Set(key string, value interface{}, expire time.Duration) {

    go func(expire int64) {
        ticker := time.Ticker(expire)
        for {
            select {
                case <- ticker.C:
                    Del(key)
                    return 
            }
        }

    }(expire)
}
```

但这种方法的缺点很明显，就是当key-value数量特别多的情况下（例如上百万个），就得起上百万个goroutine来执行定时器，明显不可行。

所以`simplecache`中采取类似Redis的过期键值对删除方法，即Lazy Delete + 定期扫描。

- Lazy Delete: 在`Get(key)`操作中，如果发现key已过期，则执行删除操作，然后返回`nil`。
- 定期扫描: 维护另外一张表`expiredTable`，类型为`map[string]*Entry`。`Set()`操作中如果key设置了过期时间，则也插入到`expiredTable`中。后台协程每隔一定时间，锁上`table`和`expiredTable`，然后遍历`expiredTable`中的`Entry`并判断是否过期。如果过期，则从`table`和`expiredTable`中删除。


至于这个时间间隔可以根据实际情况进行配置，默认是10秒:

```go

var (
	defaultOptions = Options{
		IntervalOfProactivelyDeleteExpiredKey: time.Second * 10,
		MaxMemoryPolicyType:                   MaxMemoryPolicyTypeNoeviction,
	}
)


func NewMemCache() *MemCache {
	return NewMemCacheWithOptions(defaultOptions)
}

func NewMemCacheWithOptions(options Options) *MemCache {
	mc := new(MemCache)
	mc.options = options
	mc.table = make(map[string]*Entry)
	mc.expiredTable = make(map[string]*Entry)

	mc.SetMaxMemory("1MB")

	// 后台协程主动定期清理过期key
	go mc.backgroundCleanupExpiredKeys()

	return mc
}

```

当然这种方法也是有缺点的。不做分桶每次写入操作都得对整个`map`加写锁，当数据量大的时候瓶颈显而易见。一种改进的思路是hash+bucketing。就是实际上维护多个bucket，一个bucket一个map，通过对key进行hash来判断应该写入哪个bucket的map里，然后只对该bucket加写锁，这样可以提升整个`Cache`的写入性能。


## 3. 最大内存占用限制与变量大小计算

为了给Cache支持**最大内存占用限制**这个功能，我们需要能够获取变量的内存占用大小。

Golang中没有直接提供获取变量占用内存大小的方法，但可以间接地计算一个大概的值:
- 方法1: 将变量序列化成字节数组，以字节数组的长度作为变量的大致占用内存的大小。
- 方法2: 递归地利用反射机制获取变量以及其成员变量（如果有的话）的大小。

第一种方法可以用`gob`包实现，缺点是序列化耗时，且计算并不是非常准确。
第二种方法对于层次结构较复杂的变量可能会递归耗时，但是计算相对准确。

**Notes**: 此处的计算不考虑Golang内建的数据结构的内存占用。例如，对于一个`[1024]byte{}`，我们将他的内存占用大小看成是1024, 不考虑`slice`的底层实现。同理，对于`map`，我们只考虑累加所有key和value的字面意义上的大小，不考虑`map`内部底层实现所申请的内存。


以下`Size()`函数返回一个变量的理论内存占用大小，单位为Byte。

```go

// file: github.com/monitor1379/simplecache/utils/variable.go

func Sizeof(v interface{}) int64 {
	vv := reflect.ValueOf(v)
	return sizeof(vv)
}

func sizeof(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Invalid:
		return 0
	case reflect.Ptr:
		return sizeof(v.Elem())
	case reflect.Bool, reflect.Int8, reflect.Uint8:
		return 1
	case reflect.Int16, reflect.Uint16:
		return 2
	case reflect.Int, reflect.Uint, reflect.Int32, reflect.Uint32, reflect.Float32:
		return 4
	case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Complex64, reflect.Uintptr:
		return 8
	case reflect.Complex128:
		return 16
	case reflect.String:
		return int64(v.Len())
	case reflect.Slice, reflect.Array:
		if v.Len() > 0 {
			return int64(v.Len()) * sizeof(v.Index(0))
		}
		return 0
	case reflect.Map:
		var s int64
		for _, key := range v.MapKeys() {
			s += sizeof(key) + sizeof(v.MapIndex(key))
		}
		return s

	case reflect.Struct:
		var s int64
		for i := 0; i < v.NumField(); i++ {
			s += sizeof(v.Field(i))
		}
		return s

	default:
		panic("unsupport variable type for calculating memory usage")
	}
	return 0
}

```



## 内存淘汰策略


当内存超过实现限制的最大内存数时，需要通过淘汰策略来保证`Cache`的继续可用。


由于时间关系，目前实现先只支持`no-eviction`，即当内存占用超过最大限制时，插入操作会返回失败。

后续逐步支持:
- `MaxMemoryPolicyTypeAllKeysRandom`: 从所有key中随机删除一个。
- `MaxMemoryPolicyTypeVolatileRandom`: 从设置了过期时间的key中随机删除一个。
- `MaxMemoryPolicyTypeAllKeysLRU`: 从所有key中按照LRU删除最近最少使用的一个。
- `MaxMemoryPolicyTypeVolatileLRU`: 从设置了过期时间的key中按照按照LRU删除最近最少使用的一个。



LRU的方法实现不难，使用双向链表来维护key，每当`Get(key)`时，将该`key`对应的结点移动到双向链表的最末端。然后淘汰删除的时候，删除头结点即可。