package idgen

import "errors"

var (
	ErrInvalidNodeID = errors.New("invalid node ID")
	ErrClockBackward = errors.New("clock moved backwards")
)

// Generator ID生成器接口
type Generator interface {
	NextId() (uint64, error)
	// String 返回生成器类型
	String() string
}
