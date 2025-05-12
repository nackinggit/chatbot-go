package util

import (
	"hash/crc32"
	"log"
	"math"
	"os"

	xlog "com.imilair/chatbot/bootstrap/log"
	"github.com/rs/xid"

	"github.com/bwmarrin/snowflake"
)

// The function "NewXID" generates a new unique identifier (XID) and returns it as a string.
func NewXID() string {
	return xid.New().String()
}

var node *snowflake.Node

func init() {
	snowflake.NodeBits = 6 // 2**6 nodes
	snowflake.StepBits = 6 // 2**6 every ms
	nodesCount := int64(math.Pow(2, float64(snowflake.NodeBits)))
	hostname, _ := os.Hostname()

	var hostIdx = int64(crc32.ChecksumIEEE([]byte(hostname)))
	var err error
	node, err = snowflake.NewNode(hostIdx % nodesCount)
	if err != nil {
		log.Fatalf("failed to create snowflake node: %s", err.Error())
		return
	}
	xlog.Infof("snowflake node created: %d/%d", hostIdx%nodesCount, nodesCount)
}

func NewSnowflakeID() int64 {
	return node.Generate().Int64()
}
