package booking

import (
	"errors"
	"time"

	"github.com/Promacanthus/vigour/cargo"
	"github.com/Promacanthus/vigour/location"
	"github.com/Promacanthus/vigour/routing"
)

// ErrInvalidArgument 当一个或多个参数无效时返回
var ErrInvalidArgument = errors.New("invalid argument")

type Service interface {
	// BookNewCargo 在追踪系统中注册尚未运输的新货物
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error)
	// LoadCargo  返回货物的可读模型
	LoadCargo(id cargo.TrackingID) (Cargo, error)
	// RequestPossibleRoutesForCargo 提供描述该货物可能路线的路线清单
	RequestPossibleRoutesForCargo(id cargo.TrackingID, itinerary cargo.Itinerary) []cargo.Itinerary
	// AssignCargoToRoute 将货物分配到路线
	AssignCargoToRoute(id cargo.TrackingID, itinerary cargo.Itinerary) error
	// ChangeDestination 修改货物的目的地
	ChangeDestination(id cargo.TrackingID, destination location.UNLocode) error
	// Cargos 返回所有被预订的货物列表
	Cargos() []Cargo
	// Locations 返回注册地址的列表
	Locations() []Location
}

type service struct {
	cargos         cargo.Repository
	locations      location.Repository
	handlingEvent  cargo.HandingEventRepository
	routingService routing.Service
}

func (s *service) BookNewCargo(origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error) {
	if origin == "" || destination == "" || deadline.IsZero() {
		return "", ErrInvalidArgument
	}
	id := cargo.NextTrackingID()
	rs := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: deadline,
	}
	c := cargo.New(id, rs)
	err := s.cargos.Store(c)
	if err != nil {
		return "", err
	}
	return c.TrackingID, nil
}

func (s *service) LoadCargo(id cargo.TrackingID) (Cargo, error) {
	if id == "" {
		return Cargo{}, ErrInvalidArgument
	}
	c, err := s.cargos.Find(id)
	if err != nil {
		return Cargo{}, err
	}
	return assemble(c, s.handlingEvent), nil
}

func (s *service) RequestPossibleRoutesForCargo(id cargo.TrackingID, itinerary cargo.Itinerary) []cargo.Itinerary {
	if id == "" {
		return nil
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return []cargo.Itinerary{}
	}

	// TODO
	return s.routingService.FetchRoutesForSpecification(c.RouteSpecification)
}

func (s *service) AssignCargoToRoute(id cargo.TrackingID, itinerary cargo.Itinerary) error {
	if id == "" || len(itinerary.Legs) == 0 {
		return ErrInvalidArgument
	}
	c, err := s.cargos.Find(id)
	if err != nil {
		return err
	}
	c.AssignToRoute(itinerary)
	return s.cargos.Store(c)
}

func (s *service) ChangeDestination(id cargo.TrackingID, destination location.UNLocode) error {
	if id == "" || destination == "" {
		return ErrInvalidArgument
	}

	c, err := s.cargos.Find(id)
	if err != nil {
		return err
	}

	l, err := s.locations.Find(destination)
	if err != nil {
		return err
	}

	c.SpecifyNewRoute(cargo.RouteSpecification{
		Origin:          c.Origin,
		Destination:     l.UNLocode,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	})

	err = s.cargos.Store(c)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) Cargos() []Cargo {
	var result []Cargo
	for _, c := range s.cargos.FindAll() {
		result = append(result, assemble(c, s.handlingEvent))
	}
	return result
}

func (s *service) Locations() []Location {
	var res []Location
	for _, v := range s.locations.FindAll() {
		res = append(res, Location{
			UNLocode: string(v.UNLocode),
			Name:     v.Name,
		})
	}
	return res
}

// NewService 根据必要的依赖创建一个预订服务
func NewService(cargos cargo.Repository, locations location.Repository, events cargo.HandingEventRepository, rs routing.Service) Service {
	return &service{
		cargos:         cargos,
		locations:      locations,
		handlingEvent:  events,
		routingService: rs,
	}
}

// Location 是booking视图的读取模型
type Location struct {
	UNLocode string `json:"locode"`
	Name     string `json:"name"`
}

// Cargo 是booking视图的读取模型
type Cargo struct {
	ArrivalDeadline time.Time   `json:"arrival_deadline"`
	Destination     string      `json:"destination"`
	Legs            []cargo.Leg `json:"legs,omitempty"`
	Misrouted       bool        `json:"misrouted"`
	Origin          string      `json:"origin"`
	Routed          bool        `json:"routed"`
	TrackingID      string      `json:"tracking_id"`
}

func assemble(c *cargo.Cargo, events cargo.HandingEventRepository) Cargo {
	return Cargo{
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
		Destination:     string(c.RouteSpecification.Destination),
		Legs:            c.Itinerary.Legs,
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Origin:          string(c.Origin),
		Routed:          !c.Itinerary.IsEmpty(),
		TrackingID:      string(c.TrackingID),
	}
}
