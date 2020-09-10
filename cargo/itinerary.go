package cargo

import (
	"time"

	"github.com/Promacanthus/vigour/location"
	"github.com/Promacanthus/vigour/voyage"
)

// Leg 描述了两个地点之间的一条运输路线
type Leg struct {
	VoyageNumber   voyage.Number     `json:"voyage_number"`
	LoadLocation   location.UNLocode `json:"from"`
	UnloadLocation location.UNLocode `json:"to"`
	LoadTime       time.Time         `json:"load_time"`
	UnloadTime     time.Time         `json:"unload_time"`
}

// NewLeg 创建一条新的运输路线
func NewLeg(voyageNumber voyage.Number, loadLocation, unloadLocation location.UNLocode, loadTime, unloadTime time.Time) Leg {
	return Leg{
		VoyageNumber:   voyageNumber,
		LoadLocation:   loadLocation,
		UnloadLocation: unloadLocation,
		LoadTime:       loadTime,
		UnloadTime:     unloadTime,
	}
}

// Itinerary 行程中规定了将货物从始发地运输到目的地所经过的路线
// 表示的是原计划行程中的路线
type Itinerary struct {
	Legs []Leg `json:"legs"`
}

// InitDepartureLocation 返回行程的起点
func (i Itinerary) InitDepartureLocation() location.UNLocode {
	if i.IsEmpty() {
		return location.UNLocode("")
	}
	return i.Legs[0].LoadLocation
}

// FinalArrivalLocation 返回行程的终点
func (i Itinerary) FinalArrivalLocation() location.UNLocode {
	if i.IsEmpty() {
		return location.UNLocode("")
	}
	return i.Legs[len(i.Legs)-1].UnloadLocation
}

// FindArrivalTime 返回预期到达终点的时间
func (i Itinerary) FindArrivalTime() time.Time {
	return i.Legs[len(i.Legs)-1].UnloadTime
}

// IsEmpty 判断行程中是否至少有一条运输路线
func (i Itinerary) IsEmpty() bool {
	return i.Legs == nil || len(i.Legs) == 0
}

// IsExpected 用于检查在执行这个行程时给定的处理事件是否符合预期
func (i Itinerary) IsExpected(event HandingEvent) bool {
	if i.IsEmpty() {
		return true
	}
	switch event.Activity.Type {
	case Receive:
		return i.InitDepartureLocation() == event.Activity.Location
	case Load:
		for _, leg := range i.Legs {
			if leg.LoadLocation == event.Activity.Location && leg.VoyageNumber == event.Activity.VoyageNumber {
				return true
			}
		}
		return false
	case UnLoad:
		for _, leg := range i.Legs {
			if leg.UnloadLocation == event.Activity.Location && leg.VoyageNumber == event.Activity.VoyageNumber {
				return true
			}
		}
		return false
	case Claim:
		return i.FinalArrivalLocation() == event.Activity.Location
	}
	return true
}
