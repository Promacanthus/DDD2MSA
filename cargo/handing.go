package cargo

import (
	"errors"
	"time"

	"github.com/Promacanthus/vigour/location"
	"github.com/Promacanthus/vigour/voyage"
)

// HandingEventType 表示处理事件的类型
type HandingEventType int

// 有效的处理事件的类型
const (
	NotHandled HandingEventType = iota
	Load
	UnLoad
	Receive
	Claim
	Customs
)

func (t HandingEventType) String() string {
	switch t {
	case NotHandled:
		return "Not Handled"
	case Load:
		return "Load"
	case UnLoad:
		return "UnLoad"
	case Receive:
		return "Receive"
	case Claim:
		return "Claim"
	case Customs:
		return "Customs"
	}
	return ""
}

// HandingActivity 表示如何以及在何处处理货物，
// 并且可以表达对货物未来期望发生事情的预测
type HandingActivity struct {
	Type         HandingEventType
	Location     location.UNLocode
	VoyageNumber voyage.Number
}

// HandingEvent 用于注册事件
// 例如，在给定时间从某个位置的承运人处卸下货物
type HandingEvent struct {
	TrackingID TrackingID
	Activity   HandingActivity
}

// HandingHistory 表示货物的处理历史
type HandingHistory struct {
	HandingEvents []HandingEvent
}

// MostRecentlyCompletedEvent 返回最近完成的处理事件
func (h HandingHistory) MostRecentlyCompletedEvent() (HandingEvent, error) {
	if len(h.HandingEvents) == 0 {
		return HandingEvent{}, errors.New("delivery history is empty")
	}
	return h.HandingEvents[len(h.HandingEvents)-1], nil
}

// HandingEventRepository 提供对处理事件存储的访问
type HandingEventRepository interface {
	Store(e HandingEvent)
	QueryHandingHistory(id TrackingID) HandingHistory
}

// HandingEventFactory 创建处理事件
type HandingEventFactory struct {
	CargoRepository    Repository
	VoyageRepository   voyage.Repository
	LocationRepository location.Repository
}

func (f *HandingEventFactory) CreateHandingEvent(registered time.Time, completed time.Time, id TrackingID,
	voyageNumber voyage.Number, unLocode location.UNLocode, eventType HandingEventType) (HandingEvent, error) {
	if _, err := f.CargoRepository.Find(id); err != nil {
		return HandingEvent{}, err
	}
	if _, err := f.VoyageRepository.Find(voyageNumber); err != nil {
		// TODO：这样处理有点丑陋，但是在创建一个接收事件的时候，航程的编号是不知道的
		if len(voyageNumber) > 0 {
			return HandingEvent{}, err
		}
	}

	if _, err := f.LocationRepository.Find(unLocode); err != nil {
		return HandingEvent{}, err
	}

	return HandingEvent{
		TrackingID: id,
		Activity:   HandingActivity{Type: eventType, Location: unLocode, VoyageNumber: voyageNumber}}, nil
}
