package types

import (
	"database/sql/driver"
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

// 实现 Scanner 接口
func (t *LocalTime) Scan(src interface{}) error {
	t2, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("cannot convert %v to LocalTime", src)
	}
	*t = LocalTime(t2)
	return nil
}

// 实现 Valuer 接口
func (t LocalTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}
