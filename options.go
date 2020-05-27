package simplecache

import "time"

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 15:06:26
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 15:12:06
 */

type Options struct {

	// 每隔多少时间主动扫一遍map删除过期Key
	IntervalOfProactivelyDeleteExpiredKey time.Duration
}

var (
	defaultOptions = Options{
		IntervalOfProactivelyDeleteExpiredKey: time.Second * 1,
	}
)
