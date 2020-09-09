package cargo

import (
	"errors"
	"time"

	"github.com/Promacanthus/vigour/location"
)

// ErrUnknown 表示货物不存在
var ErrUnknown = errors.New("unknown cargo")

// TrackingID 唯一标识特定货物
type TrackingID string

// RouteSpecification 包含一条路线的信息：
// 起点、终点和到达的截止时间
type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

// Cargo 表示货物是领域模型中的一个核心类（聚合根？）
type Cargo struct {
	TrackingID         TrackingID
	Origin             location.UNLocode
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery           Delivery
}

func (c *Cargo) SpecifyNewRoute(rs RouteSpecification) {
	c.RouteSpecification = rs
}

// RoutingStatus 表示货物路线的状态
type RoutingStatus int

// TransportStatus 表示货物运输的状态
type TransportStatus int

// Repository 提供对货物存储的访问
type Repository interface {
	Store(cargo *Cargo) error
	Find(id TrackingID) (*Cargo, error)
	FindAll() []*Cargo
}
