package utils

import "time"

// GetKardsNow 生成类似 2023-11-06T11:00:00.000Z 的时间字符串
func GetKardsNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000") + "Z"
}
