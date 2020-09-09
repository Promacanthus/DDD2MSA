package voyage

import (
	"errors"
	"time"

	"github.com/Promacanthus/vigour/location"
)

// ErrUnknown 表示航程不存在
var ErrUnknown = errors.New("unknown voyage")

// Number 唯一标识特定航程
type Number string

// CarrierMovement 表示船只从一个地方到另一个地方的行程
type CarrierMovement struct {
	DepartureLocation location.UNLocode
	ArrivalLocation   location.UNLocode
	DepartureTime     time.Time
	ArrivalTime       time.Time
}

// Schedule 表示航行时间表
type Schedule struct {
	CarrierMovements []CarrierMovement
}

// Voyage 是一系列航程的唯一标识
type Voyage struct {
	Number   Number
	Schedule Schedule
}

// New 根据航程编号和提供航程时间表新建航程
func New(n Number, s Schedule) *Voyage {
	return &Voyage{Number: n, Schedule: s}
}

// Repository 提供对航程存储的访问
type Repository interface {
	Find(Number) (*Voyage, error)
}
