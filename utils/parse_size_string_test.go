package utils_test

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 17:11:27
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 17:19:11
 */

import (
	"testing"

	"github.com/monitor1379/simplecache/utils"
)

func TestParseSizeString(t *testing.T) {
	testCases := []struct {
		SizeStr string
		Size    int64
	}{
		{"1KB", 1024},
		{"1024B", 1024},
		{"1KB", 1024},
		{"-1KB", -1024},
		{"2048kb", 2097152},
		{"1GB", 1073741824},
		{"1.2gb", 1288490188}, // int64(1.2* 1024 * 1024 * 1024)
	}

	for _, testCase := range testCases {
		size, err := utils.ParseSizeString(testCase.SizeStr)
		if err != nil {
			t.Error("error:", err)
			continue
		}

		if size != testCase.Size {
			t.Errorf("expect %d but got %d", testCase.Size, size)
		}
	}
	return
}
