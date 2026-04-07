//go:build !linux
// +build !linux

package device

import (
	"github.com/bingbaga/AnyWhereEtherNet/conn"
	"github.com/bingbaga/AnyWhereEtherNet/rwcancel"
)

func (device *Device) startRouteListener(bind conn.Bind) (*rwcancel.RWCancel, error) {
	return nil, nil
}
