package util

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	xlog "com.imilair/chatbot/bootstrap/log"
)

func Marshal(src any) ([]byte, error) {
	byteBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(src)
	if err != nil {
		return nil, err
	}
	dst := byteBuf.Bytes()
	return dst, nil
}

func JsonString(d any) string {
	data, _ := Marshal(d)
	return string(data)
}

func Unmarshal(src []byte, dst any) error {
	return json.Unmarshal(src, dst)
}

func JsonToMap(input string) map[string]any {
	var data map[string]any
	err := Unmarshal([]byte(input), &data)
	if err != nil {
		xlog.Warnf("JsonToMap err， data:%v, err:%v", input, err)
	}
	return data
}

// 将json转map，所有数字换成整形
func JsonToMapWithNumber(input string) map[string]any {
	if input == "" {
		return map[string]any{}
	}
	reader := strings.NewReader(input)

	// 创建一个新的 Decoder，并启用 UseNumber
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	// 解析 JSON 到 map[string]interface{}
	var result map[string]any
	if err := decoder.Decode(&result); err != nil {
		xlog.Warnf("JsonToMapWithNumber err， data:%v, err:%v", input, err)
	}

	// 遍历 map 并将 json.Number 转换为 int
	for k, v := range result {
		if num, ok := v.(json.Number); ok {
			if i, err := strconv.ParseInt(num.String(), 10, 64); err == nil {
				result[k] = i // 将转换后的 int 存回 map
			} else {
				xlog.Warnf("Error converting number to int for key '%s': %v", k, err)
			}
		}
	}
	return result
}

func ConvertMapAny2Int(m map[string]string) map[string]any {
	newMap := make(map[string]any)
	for k, v := range m {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			newMap[k] = i
		} else {
			newMap[k] = v
		}
	}
	return newMap
}
