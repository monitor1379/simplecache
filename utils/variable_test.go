package utils_test

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 18:14:25
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 23:59:09
 */

import (
	"testing"

	"github.com/monitor1379/simplecache/utils"
)

func TestSizeof(t *testing.T) {

	testCases := []struct {
		Var  interface{}
		Size int64
	}{
		{int32(1), 4},
		{int64(1), 8},
		{float32(1), 4},
		{float64(1), 8},
		{"12345", 5},
		{"0123456789", 10},
		{[1024]byte{}, 1024},
		{
			map[string][]byte{
				"key": make([]byte, 1000),
			},
			1003,
		},
	}

	for _, testCase := range testCases {
		if s := utils.Sizeof(testCase.Var); s != testCase.Size {
			t.Errorf("expect %d but got %d\n", testCase.Size, s)
		}
	}

}
