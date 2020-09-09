package location

import "errors"

// ErrUnknown 表示一个地址不存在
var ErrUnknown = errors.New("unknown location")

// UNLocode 唯一标识特定位置的联合国位置代码
// http://www.unece.org/cefact/locode/
// http://www.unece.org/cefact/locode/DocColumnDescription.htm#LOCODE
type UNLocode string

// Location 表示模型在路途中停留的位置
// 例如：货物的起点或终点；承运人的移动终点
type Location struct {
	UNLocode UNLocode
	Name string
}

// Repository 提供对位置信息存储的访问
type Repository interface {
	Find(locode UNLocode)(*Location,error)
	FindAll()[]*Location
}