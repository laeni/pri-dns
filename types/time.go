package types

import (
	"fmt"
	"time"
)

type LocalTime time.Time

// MarshalJSON on LocalTime format Time field with %Y-%m-%d %H:%M:%S
func (t *LocalTime) MarshalJSON() ([]byte, error) {
	format := time.Time(*t).Format("2006-01-02 15:04:05")
	// 由于序列化之后为字符串，所以必须加引号
	return []byte(fmt.Sprintf("\"%s\"", format)), nil
}
