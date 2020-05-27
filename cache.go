package simplecache

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 13:53:35
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 15:12:59
 */

import "time"

type Cache interface {

	// size 是一个字符串。支持以下参数: 1KB, 100KB, 1MB, 2MB, 1GB 等
	SetMaxMemory(size string) bool

	// 设置一个缓存项，并且在expire时间之后过期
	Set(key string, value interface{}, expire time.Duration)

	// 获取一个值
	Get(key string) (interface{}, bool)

	// 删除一个值
	Del(key string) bool

	// 检测一个值是否存在
	Exists(key string) bool

	// 清空所有值
	Flush() bool

	// 返回key的数量
	Keys() int64
}

func New() Cache {
	return NewMemCache()
}

func NewWithOptions(options Options) Cache {
	return NewMemCacheWithOptions(options)
}
