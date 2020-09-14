package routing

import "github.com/Promacanthus/vigour/cargo"

// Service 提供对外部路由服务的访问
type Service interface {
	// FetchRouteFromSpecification 查找满足给定规范的全部线路
	FetchRoutesForSpecification(rs cargo.RouteSpecification) []cargo.Itinerary
}
