package util

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"

	xlog "com.imilair/chatbot/bootstrap/log"
)

// FindIndexOf searches an element in a slice based on a predicate and returns the index and true.
// It returns -1 and false if the element is not found.
func FindIndexOf[T any](collection []T, predicate func(item T) bool) int {
	for i := range collection {
		if predicate(collection[i]) {
			return i
		}
	}

	return -1
}
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func Md5Object(x any) string {
	jsonBytes, _ := json.Marshal(x)
	hash := md5.Sum(jsonBytes)
	return hex.EncodeToString(hash[:])
}

func Struct2Map(x any) map[string]any {
	b, err := json.Marshal(x)
	if err != nil {
		return nil
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil
	}
	return m
}

func Map2Struct[T any](m map[string]any, t *T) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, t)
	if err != nil {
		return err
	}
	return nil
}

func AsyncGo(ctx context.Context, fn func(), handler func(ctx context.Context, sig error)) {
	go func() {
		defer func() {
			if sig := recover(); sig != nil {
				xlog.Errorf("GOPOOL: panic: %s", debug.Stack())
				handler(ctx, fmt.Errorf("%v", sig))
			}
		}()
		fn()
	}()
}

func AsyncGoWithDefault(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if sig := recover(); sig != nil {
				xlog.Errorf("GOPOOL: panic: %s", debug.Stack())
			}
		}()
		fn()
	}()
}

func FormatDate(t string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", t, time.Local)
}

func Int64Array2String(ints []int64, spliter string) string {
	var sb strings.Builder
	for i, v := range ints {
		sb.WriteString(strconv.FormatInt(v, 10))
		if i < len(ints)-1 {
			sb.WriteString(spliter)
		}
	}
	return sb.String()
}

func String2Int64Array(str string, spliter string) []int64 {
	ids := []int64{}
	if str == "" {
		return ids
	}
	arr := strings.Split(str, spliter)
	for _, v := range arr {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func IntArray2String(ints []int, spliter string) string {
	var sb strings.Builder
	for i, v := range ints {
		sb.WriteString(strconv.Itoa(v))
		if i < len(ints)-1 {
			sb.WriteString(spliter)
		}
	}
	return sb.String()
}

func String2IntArray(str string, spliter string) []int {
	ids := []int{}
	if str == "" {
		return ids
	}
	arr := strings.Split(str, spliter)
	for _, v := range arr {
		id, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func Split(content string, regstr string) []string {
	re := regexp.MustCompile(regstr)
	parts := re.Split(content, -1)
	for i, v := range parts {
		parts[i] = strings.TrimSpace(v)
	}
	return parts
}

func StringToInt64(num string) int64 {
	r, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		xlog.Warnf("StringToInt64 failed: val: %v, err: %v", num, err.Error())
	}
	return r
}

func StringToInt(num string) int {
	r, err := strconv.Atoi(num)
	if err != nil {
		xlog.Warnf("StringToInt64 failed: val: %v, err: %v", num, err.Error())
	}
	return r
}

func RandCode() string {
	return fmt.Sprintf("%04v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
}

func PhoneMask(phone string) string {
	slen := len(phone)
	if slen < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[slen-4:]
}

func MonthFirstDay() time.Time {
	now := time.Now()
	firstDateTime := now.AddDate(0, 0, -now.Day()+1)
	return time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location())
}

func MonthLastDay() time.Time {
	now := time.Now()
	lastDateTime := now.AddDate(0, 1, -now.Day())
	return time.Date(lastDateTime.Year(), lastDateTime.Month(), lastDateTime.Day(), 23, 59, 59, 0, lastDateTime.Location())
}

func DayEndTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
}

func DayFirstTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func CopyStruct[T any](src T) T {
	var dst T
	data := JsonString(src)
	Unmarshal([]byte(data), &dst)
	return dst
}

func SubString(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)
	start = max(0, start)
	end = min(length, end)
	if start >= end {
		return ""
	}
	return string(rs[start:end])
}
