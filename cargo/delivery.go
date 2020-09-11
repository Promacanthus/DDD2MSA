package cargo

import (
	"time"

	"github.com/Promacanthus/vigour/location"
	"github.com/Promacanthus/vigour/voyage"
)

// Delivery 是货物实际运送的路线
// 而不是客户要求（RouteSpecification）和计划（Itinerary）的路线
type Delivery struct {
	Itinerary               Itinerary
	RouteSpecification      RouteSpecification
	RoutingStatus           RoutingStatus
	TransportStatus         TransportStatus
	NextExpectedActivity    HandingActivity
	LastEvent               HandingEvent
	LastKnownLocation       location.UNLocode
	CurrentVoyage           voyage.Number
	ETA                     time.Time
	IsMisdirected           bool
	IsUnloadedAtDestination bool
}

// UpdateOnRouting 新建交货的快照来反应路线的变化
// 例如，客户需求的路线或行程发生了更改，但没有对货物进行其他任何处理
func (d Delivery) UpdateOnRouting(rs RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, rs)
}

// IsOnTrack 检查交货是否按计划进行
func (d Delivery) IsOnTrack() bool {
	return d.RoutingStatus == Routed && !d.IsMisdirected
}

// DeriveDeliveryFrom 根据货物完整的处理历史记录、货物的线路规划和行程创建一个新的交货快照
func DeriveDeliveryFrom(rs RouteSpecification, itinerary Itinerary, history HandingHistory) Delivery {
	lastEvent, _ := history.MostRecentlyCompletedEvent()
	return newDelivery(lastEvent, itinerary, rs)

}

// newDelivery 根据处理事件、行程和路线创建最新的交货
func newDelivery(lastEvent HandingEvent, itinerary Itinerary, rs RouteSpecification) Delivery {
	var (
		routingStatus           = calculateRoutingStatus(itinerary, rs)
		transportStatus         = calculateTransportStatus(lastEvent)
		lastKnownLocation       = calculateLastKnownLocation(lastEvent)
		isMisdirected           = calculateMisdirectedStatus(lastEvent, itinerary)
		isUnloadedAtDestination = calculateUnloadedAtDestination(lastEvent, rs)
		currentVoyage           = calculateCurrentVoyage(transportStatus, lastEvent)
	)

	d := Delivery{
		Itinerary:               itinerary,
		RouteSpecification:      rs,
		RoutingStatus:           routingStatus,
		TransportStatus:         transportStatus,
		LastEvent:               lastEvent,
		LastKnownLocation:       lastKnownLocation,
		CurrentVoyage:           currentVoyage,
		IsMisdirected:           isMisdirected,
		IsUnloadedAtDestination: isUnloadedAtDestination,
	}

	d.NextExpectedActivity = calculateNextExpectedActivity(d)
	d.ETA = calculateETA(d)
	return d
}

// 内部函数，新建delivery时使用
func calculateRoutingStatus(itinerary Itinerary, rs RouteSpecification) RoutingStatus {
	if itinerary.Legs == nil {
		return NotRouted
	}
	if rs.IsSatisfiedBy(itinerary) {
		return Routed
	}
	return Misrouted
}

func calculateMisdirectedStatus(event HandingEvent, itinerary Itinerary) bool {
	if event.Activity.Type == NotHandled {
		return false
	}
	return !itinerary.IsExpected(event)
}

func calculateUnloadedAtDestination(event HandingEvent, rs RouteSpecification) bool {
	if event.Activity.Type == NotHandled {
		return false
	}
	return event.Activity.Type == UnLoad && rs.Destination == event.Activity.Location
}

func calculateTransportStatus(event HandingEvent) TransportStatus {
	switch event.Activity.Type {
	case NotHandled:
		return NotReceived
	case Load:
		return OnboardCarrier
	case UnLoad, Receive, Customs:
		return InPort
	case Claim:
		return Claimed
	default:
		return Unknown
	}
}

func calculateLastKnownLocation(event HandingEvent) location.UNLocode {
	return event.Activity.Location
}

func calculateNextExpectedActivity(d Delivery) HandingActivity {
	if !d.IsOnTrack() {
		return HandingActivity{}
	}

	switch d.LastEvent.Activity.Type {
	case NotHandled:
		return HandingActivity{Type: Receive, Location: d.RouteSpecification.Origin}
	case Receive:
		l := d.Itinerary.Legs[0]
		return HandingActivity{
			Type:         Load,
			Location:     l.LoadLocation,
			VoyageNumber: l.VoyageNumber,
		}
	case Load:
		for _, l := range d.Itinerary.Legs {
			if l.LoadLocation == d.LastEvent.Activity.Location {
				return HandingActivity{
					Type:         UnLoad,
					Location:     l.UnloadLocation,
					VoyageNumber: l.VoyageNumber,
				}
			}
		}
	case UnLoad:
		for i, l := range d.Itinerary.Legs {
			if l.UnloadLocation == d.LastEvent.Activity.Location {
				if i < len(d.Itinerary.Legs)-1 {
					return HandingActivity{
						Type:         Load,
						Location:     d.Itinerary.Legs[i+1].LoadLocation,
						VoyageNumber: d.Itinerary.Legs[i+1].VoyageNumber,
					}
				}
				return HandingActivity{
					Type:     Claim,
					Location: l.LoadLocation,
				}
			}
		}
	}
	return HandingActivity{}
}

func calculateCurrentVoyage(status TransportStatus, event HandingEvent) voyage.Number {
	if status == OnboardCarrier && event.Activity.Type != NotHandled {
		return event.Activity.VoyageNumber
	}
	return voyage.Number("")
}

func calculateETA(d Delivery) time.Time {
	if !d.IsOnTrack() {
		return time.Time{}
	}
	return d.Itinerary.FindArrivalTime()
}
