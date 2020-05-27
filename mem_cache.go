package simplecache

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 13:56:36
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 17:29:10
 */

import (
	"fmt"
	"sync"
	"time"

	"github.com/monitor1379/simplecache/utils"
)

type Entry struct {
	value       interface{}
	expiredNano int64
}

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

func NewMemCache() Cache {
	return NewMemCacheWithOptions(defaultOptions)
}

func NewMemCacheWithOptions(options Options) Cache {
	mc := new(MemCache)
	mc.options = options
	mc.table = make(map[string]*Entry)
	mc.expiredTable = make(map[string]*Entry)

	// 后台协程主动定期清理过期key
	go mc.backgroundCleanupExpiredKeys()

	return mc
}

func (mc *MemCache) backgroundCleanupExpiredKeys() {
	// 每隔一段时间，扫描数据库中expiredTable中key，判断是否过期并清理掉
	// 在扫描的过程中，数据库会发生阻塞
	ticker := time.NewTicker(mc.options.IntervalOfProactivelyDeleteExpiredKey)
	for {
		select {
		case <-ticker.C:
			mc.doCleanupExpiredKeysImmediately()
		}
	}
}

func (mc *MemCache) doCleanupExpiredKeysImmediately() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.expiredMu.Lock()
	defer mc.expiredMu.Unlock()

	// map的delete在遍历时是安全的
	// 将过期的key从MemCache.table以及MemCache.expiredTable中删除
	for key := range mc.expiredTable {
		entry := mc.expiredTable[key]
		if entry.expiredNano < time.Now().UnixNano() {
			delete(mc.table, key)
			delete(mc.expiredTable, key)
		}
	}

}

func (mc *MemCache) SetMaxMemory(size string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// TODO(monitor1379): 该函数返回的bool的true/false表示什么意思
	maxMemory, err := utils.ParseSizeString(size)
	if err != nil {
		fmt.Println("[WARNING] simplecache/mem_cache.go#93: parse size string faild:", err)
		return false
	}

	sysTotalMemory, err := utils.GetSystemTotalMemory()
	if err != nil {
		fmt.Println("[WARNING] simplecache/mem_cache.go#99: get system total memory faild:", err)
		return false
	}

	if maxMemory > sysTotalMemory {
		fmt.Println("[WARNING] simplecache/mem_cache.go#104: using system max memory as max memory size")
		maxMemory = sysTotalMemory
		return false
	}

	mc.maxMemory = maxMemory
	return true
}

func (mc *MemCache) Set(key string, value interface{}, expire time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 计算该key的过期时间
	expiredNano := time.Now().UnixNano() + expire.Nanoseconds()

	entry := &Entry{
		value:       value,
		expiredNano: expiredNano,
	}
	mc.table[key] = entry

	// 如果该key设置了过期时间，才将其加入到expiredTable中
	if expire > 0 {
		mc.expiredMu.Lock()
		defer mc.expiredMu.Unlock()

		mc.expiredTable[key] = entry
	}

	return
}

func (mc *MemCache) Get(key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, ok := mc.table[key]
	if !ok {
		return nil, false
	}

	// 在Get操作中lazy删除
	if entry.expiredNano < time.Now().UnixNano() {
		mc.expiredMu.Lock()
		defer mc.expiredMu.Unlock()

		delete(mc.table, key)
		delete(mc.expiredTable, key)

		return nil, false
	}
	return entry.value, true
}

func (mc *MemCache) Del(key string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	_, ok := mc.table[key]
	if ok {
		delete(mc.table, key)
	}

	return ok
}

func (mc *MemCache) Exists(key string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	_, ok := mc.table[key]

	return ok
}

func (mc *MemCache) Flush() bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.expiredMu.Lock()
	defer mc.expiredMu.Unlock()

	mc.table = make(map[string]*Entry)
	mc.expiredTable = make(map[string]*Entry)

	// TODO(monitor1379): 这里返回值的意义是什么，什么时候返回true/false
	return true
}

func (mc *MemCache) Keys() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return int64(len(mc.table))
}
