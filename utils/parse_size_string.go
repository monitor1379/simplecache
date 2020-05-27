package utils

/*
 * @Author: ZhenpengDeng(monitor1379)
 * @Date: 2020-05-27 16:52:11
 * @Last Modified by: ZhenpengDeng(monitor1379)
 * @Last Modified time: 2020-05-27 17:26:03
 */

import (
	"errors"
	"strconv"
	"strings"
)

const (
	UnitB  = "B"
	UnitKB = "KB"
	UnitMB = "MB"
	UnitGB = "GB"
	UnitTB = "TB"
)

var (
	Units = []string{
		UnitTB,
		UnitKB,
		UnitMB,
		UnitGB,
		UnitB,
	}

	UnitToBytes = map[string]int{
		UnitB:  1,
		UnitKB: 1 << 10,
		UnitMB: 1 << 20,
		UnitGB: 1 << 30,
		UnitTB: 1 << 40,
	}
)

var (
	ErrInvalidUnit   = errors.New("simplecache/utils: invalid unit")
	ErrUnsupportUnit = errors.New("simplecache/utils: unsupport unit")
)

func ParseSizeString(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(sizeStr)

	unit := ""
	for i := range Units {
		if strings.HasSuffix(sizeStr, Units[i]) {
			unit = Units[i]
			break
		}
	}
	if unit == "" {
		return 0, ErrInvalidUnit
	}

	realSizeStr := strings.TrimSuffix(sizeStr, unit)
	realSize, err := strconv.ParseFloat(realSizeStr, 64)

	if err != nil {
		return 0, err
	}

	unitBytes, ok := UnitToBytes[unit]
	if !ok {
		return 0, ErrUnsupportUnit
	}

	return int64(realSize * float64(unitBytes)), nil
}
