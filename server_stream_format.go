package gortsplib

import (
	"github.com/westpoint-robotics/gortsplib/pkg/format"
	"github.com/westpoint-robotics/gortsplib/pkg/rtcpsender"
)

type serverStreamFormat struct {
	format     format.Format
	rtcpSender *rtcpsender.RTCPSender
}
