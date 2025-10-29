package snowflake

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

/*
@author: shg
@since: 2024/1/24 11:34 PM
@mail: shgang97@163.com
*/

/*
0-41位毫秒级时间戳-10位节点id-12位序列号
*/

const (
	timestampBits        = 41
	workerIdBits         = 10
	sequenceBits         = 12
	timestampShift       = workerIdBits + sequenceBits
	workIdShift          = sequenceBits
	sequenceMask         = 1<<sequenceBits - 1
	workIdMax      int64 = -1 ^ (-1 << workerIdBits)
)

var (
	/*
	   start time： 2024-01-24
	*/
	twepoch = time.Date(2024, 1, 24, 0, 0, 0, 0, time.UTC).UnixMilli()
)

type Node struct {
	lastTimestamp int64
	workId        int64
	sequence      int64

	lock sync.Mutex
}

// Init returns a new snowflake node that can be used to generate snowflake
func Init(workId int64) (*Node, error) {
	n := Node{}
	n.workId = workId
	if n.workId < 0 || n.workId > workIdMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(workIdMax, 10))
	}
	return &n, nil
}

func (n *Node) NextId() (id int64, err error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	now := time.Now().UnixMilli()
	// 发生时钟回拨
	if now < n.lastTimestamp {
		return id, fmt.Errorf("clock callback occurs")
	}

	if now == n.lastTimestamp {
		n.sequence = (n.sequence + 1) & sequenceMask
		if n.sequence == 0 {
			// 并发量太高，序列号溢出
			now = wait(n.lastTimestamp)
		}
	} else {
		n.sequence = 0
	}
	n.lastTimestamp = now
	ts := getNewestTimestamp(n.lastTimestamp)
	id = (ts << timestampShift) | (n.workId << workIdShift) | (n.sequence)
	return id, nil
}

func wait(lastTimestamp int64) int64 {
	now := time.Now().UnixMilli()
	if now <= lastTimestamp {
		time.Sleep(time.Millisecond * time.Duration(lastTimestamp-now+1))
	}
	return now
}

/*
get the newest timestamp relative to twepoch
*/
func getNewestTimestamp(lastTimestamp int64) int64 {
	return lastTimestamp - twepoch
}
