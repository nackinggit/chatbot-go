package util

import (
	"testing"
	"time"
)

func TestNewSnowflakeID(t *testing.T) {
	for i := range 100 {
		snId := NewSnowflakeID()
		time.Sleep(time.Millisecond)
		t.Logf("%d -> snowflakeID:%s time:%d, node:%d, step:%d", i, snId.String(), snId.Time(), snId.Node(), snId.Step())
	}
}

type Test struct {
	Name string
}

func TestUtil(t *testing.T) {
	// t.Logf("%v", Split("小明", `,|，`))
	// t.Logf("%v", Split("小明,小花", `,|，`))
	// t.Logf("%v", RandCode())
	// t.Logf("%v", PhoneMask("12342123"))
	// t.Logf("%v", MonthFirstDay())
	// t.Logf("%v", MonthLastDay())
	m := Test{}
	err := TryParseJson(`学生宿舍为：{"name": "小明"}`, &m)
	t.Logf("%v, %v", m, err)
	marray := []Test{}
	err = TryParseJsonArray(`学生宿舍为：[{"name": "小明"}]`, &marray)
	t.Logf("%v, %v", marray, err)
}
