package util

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
)

func TestNewSnowflakeID(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := NewSnowflakeID()
		time.Sleep(time.Millisecond)
		snId := snowflake.ParseInt64(id)
		t.Logf("%d -> snowflakeID:%d time:%d, node:%d, step:%d", i, id, snId.Time(), snId.Node(), snId.Step())
	}
}

func TestUtil(t *testing.T) {
	// t.Logf("%v", Split("小明", `,|，`))
	// t.Logf("%v", Split("小明,小花", `,|，`))
	// t.Logf("%v", RandCode())
	// t.Logf("%v", PhoneMask("12342123"))
	t.Logf("%v", MonthFirstDay())
	t.Logf("%v", MonthLastDay())
}
