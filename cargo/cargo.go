package cargo

import (
	"errors"
	"strings"
	"time"

	"github.com/pborman/uuid"

	"github.com/Promacanthus/vigour/location"
)

// ErrUnknown 表示货物不存在
var ErrUnknown = errors.New("unknown cargo")

// TrackingID 唯一标识特定货物
type TrackingID string

// RouteSpecification 包含一条路线的信息：起点、终点和到达的截止时间
// 表示的是客户要求的运送路线
type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

// IsSatisfiedBy 用于检查提供的行程是否能满足客户的需求
func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return itinerary.Legs != nil &&
		s.Origin == itinerary.InitDepartureLocation() &&
		s.Destination == itinerary.FinalArrivalLocation()
}

// Cargo 表示货物是领域模型中的一个核心类（聚合根？）
type Cargo struct {
	TrackingID         TrackingID
	Origin             location.UNLocode
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery           Delivery
}

// SpecifyNewRoute 给货物指定一条新路线
func (c *Cargo) SpecifyNewRoute(rs RouteSpecification) {
	c.RouteSpecification = rs
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// AssignToRoute 给货物附加新行程
func (c *Cargo) AssignToRoute(itinerary Itinerary) {
	c.Itinerary = itinerary
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// DeriveDeliveryProgress 根据当前的路线规划、行程和货物处理情况来更新货物各方面的汇总状态
func (c *Cargo) DeriveDeliveryProgress(history HandingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
}

// New 创建一个新的没有路线信息的货物
func New(id TrackingID, rs RouteSpecification) *Cargo {
	itinerary := Itinerary{}
	history := HandingHistory{make([]HandingEvent, 0)}
	return &Cargo{
		TrackingID:         id,
		Origin:             rs.Origin,
		RouteSpecification: rs,
		Delivery:           DeriveDeliveryFrom(rs, itinerary, history),
	}
}

// NextTrackingID 生产一个新的 Tracking ID
// TODO: Move to infrastructure(?)
func NextTrackingID() TrackingID {
	return TrackingID(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}

// RoutingStatus 表示货物路线的状态
type RoutingStatus int

// 有效的路线状态
const (
	NotRouted RoutingStatus = iota
	Misrouted
	Routed
)

func (s RoutingStatus) String() string {
	switch s {
	case NotRouted:
		return "Not routed"
	case Misrouted:
		return "Misrouted"
	case Routed:
		return "Routed"
	default:
		return ""
	}
}

// TransportStatus 表示货物运输的状态
type TransportStatus int

// 有效的运送状态
const (
	NotReceived TransportStatus = iota
	InPort
	OnboardCarrier
	Claimed
	Unknown
)

func (s TransportStatus) String() string {
	switch s {
	case NotReceived:
		return "Not received"
	case InPort:
		return "In port"
	case OnboardCarrier:
		return "Onboard carrier"
	case Claimed:
		return "Claimed"
	case Unknown:
		return "Unknown"
	default:
		return ""
	}
}

// Repository 提供对货物存储的访问
type Repository interface {
	Store(cargo *Cargo) error
	Find(id TrackingID) (*Cargo, error)
	FindAll() []*Cargo
}
