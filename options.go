package simplecache

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 15:06:26
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 22:04:02
 */

import (
	"errors"
	"time"
)

type MaxMemoryPolicyType int

const (
	// 超出最大使用内存时的淘汰策略

	// 不删除，插入失败并返回error
	MaxMemoryPolicyTypeNoeviction = iota
	// MaxMemoryPolicyTypeAllKeysRandom
	// MaxMemoryPolicyTypeVolatileRandom
	// MaxMemoryPolicyTypeAllKeysLRU
	// MaxMemoryPolicyTypeVolatileLRU
)

var (
	ErrUnknownMaxMemoryPolicyType = errors.New("unknown max memory policy type")
	ErrOutOfMaxMemory             = errors.New("out of max memory")
)

type Options struct {

	// 每隔多少时间主动扫一遍map删除过期Key
	IntervalOfProactivelyDeleteExpiredKey time.Duration

	// 超出最大使用内存时的淘汰策略
	MaxMemoryPolicyType MaxMemoryPolicyType
}

var (
	defaultOptions = Options{
		IntervalOfProactivelyDeleteExpiredKey: time.Second * 10,
		MaxMemoryPolicyType:                   MaxMemoryPolicyTypeNoeviction,
	}
)
