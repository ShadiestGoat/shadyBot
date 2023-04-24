package snownode

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Dextication/snowflake"
	"github.com/shadiestgoat/log"
)

var node *snowflake.Node

var base_id_time = time.Date(2019, time.March, 5, 0, 0, 0, 0, time.UTC)
var base_id_stamp = base_id_time.UnixMilli()

func init() {
	var err error
	node, err = snowflake.NewNode(0, base_id_time, 41, 11, 11)
	log.FatalIfErr(err, "Creating snowflake node")
}

func SnowToTime(id string) time.Time {
	i, _ := strconv.ParseInt(id, 10, 64)

	timestamp := (i >> 22) + base_id_stamp

	return time.UnixMilli(timestamp)
}

func TimeToSnow(time time.Time) string {
	stamp := time.UnixMilli()
	stamp -= base_id_stamp

	return fmt.Sprint(stamp << 22)
}

func Generate() string {
	return node.Generate().String()
}
