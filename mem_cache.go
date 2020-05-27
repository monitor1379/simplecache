package simplecache

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 13:56:36
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 23:54:04
 */

import (
	"errors"
	"sync"
	"time"

	"github.com/monitor1379/simplecache/utils"
)

type Entry struct {
	value       interface{}
	expiredNano int64 // 过期时间戳。0表示永不过期
	valueSize   int64 // sizeof(value)
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
			mc.memoryUsage = mc.memoryUsage - entry.valueSize - int64(len(key))
		}
	}

}

func (mc *MemCache) SetMaxMemory(size string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	maxMemory, err := utils.ParseSizeString(size)
	if err != nil {
		return err
	}

	systemTotalMemory, err := utils.GetSystemTotalMemory()
	if err != nil {
		return err
	}

	if maxMemory > systemTotalMemory {
		return errors.New("invalid max memory size")
	}

	mc.maxMemory = maxMemory
	return nil
}

func (mc *MemCache) Set(key string, value interface{}, expire time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 计算key-value的内存大小占用
	keySize := int64(len(key))
	valueSize := utils.Sizeof(value)
	var incrSize int64

	// 计算写入该key-value会新增多少内存
	// 如果key不存在，则新增内存大小字节数为：sizeof(key) + sizeof(value)
	// 如果key已经存在，则新增内存大小为sizeof(value) - sizeof(originalValue)
	entry, exists := mc.table[key]
	if !exists {
		incrSize = keySize + valueSize
	} else {
		incrSize = valueSize - entry.valueSize
	}

	// 判断写入该key - value是否会超过最大内存大小，如果超过，先触发淘汰策略
	if mc.memoryUsage+incrSize > mc.maxMemory {
		switch mc.options.MaxMemoryPolicyType {
		case MaxMemoryPolicyTypeNoeviction:
			return ErrOutOfMaxMemory
		default:
			return ErrUnknownMaxMemoryPolicyType
		}
	}

	// 计算该key的过期时间
	var expiredNano int64
	if expire != 0 {
		expiredNano = time.Now().Add(expire).UnixNano()
	}

	entry = &Entry{
		value:       value,
		expiredNano: expiredNano,
		valueSize:   valueSize,
	}

	// 写入table
	mc.table[key] = entry
	mc.memoryUsage += incrSize

	// 如果该key设置了过期时间，才将其加入到expiredTable中
	// 如果该key之前已经存在，则此次Set()操作还会更新该key的过期时间
	// 如果原先设置了过期时间，但该次Set()操作的expire为0，将其从expiredTable中删除
	if expire > 0 {
		mc.expiredMu.Lock()
		mc.expiredTable[key] = entry
		mc.expiredMu.Unlock()
	} else if exists {
		mc.expiredMu.Lock()
		delete(mc.expiredTable, key)
		mc.expiredMu.Unlock()
	}

	return nil
}

func (mc *MemCache) Get(key string) (interface{}, bool) {
	mc.mu.RLock()
	entry, ok := mc.table[key]
	mc.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// Lazy delete: 如果该key设置了过期时间且已经过期
	if entry.expiredNano > 0 && entry.expiredNano < time.Now().UnixNano() {
		mc.mu.Lock()
		delete(mc.table, key)
		mc.mu.Unlock()

		mc.expiredMu.Lock()
		delete(mc.expiredTable, key)
		mc.expiredMu.Unlock()

		return nil, false
	}
	return entry.value, true
}

func (mc *MemCache) Del(key string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, ok := mc.table[key]
	if ok {
		delete(mc.table, key)
		mc.memoryUsage -= entry.valueSize
	}

	return ok
}

func (mc *MemCache) Exists(key string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	_, ok := mc.table[key]

	return ok
}

func (mc *MemCache) Flush() error {
	mc.mu.Lock()
	mc.table = make(map[string]*Entry)
	mc.memoryUsage = 0
	mc.mu.Unlock()

	mc.expiredMu.Lock()
	mc.expiredTable = make(map[string]*Entry)
	mc.expiredMu.Unlock()

	return nil
}

// Notes: 因为lazy delete，所以该方法返回的数量是不太精确的
func (mc *MemCache) Keys() int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return int64(len(mc.table))
}

func (mc *MemCache) MemoryUsage() int64 {
	return mc.memoryUsage
}
