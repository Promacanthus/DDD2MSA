package booking

import (
	"time"

	"github.com/Promacanthus/vigour/cargo"
	"github.com/Promacanthus/vigour/location"
)

type Service interface {
	// 	BookNewCargo在追踪系统中注册尚未路由的新货物
	BookNewCargo(origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error)
}
