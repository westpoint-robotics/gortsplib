package main

import (
	"log"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/westpoint-robotics/gortsplib"
	"github.com/westpoint-robotics/gortsplib/pkg/format"
	"github.com/westpoint-robotics/gortsplib/pkg/media"
	"github.com/westpoint-robotics/gortsplib/pkg/url"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. read all media streams on a path.

func main() {
	c := gortsplib.Client{}

	// parse URL
	u, err := url.Parse("rtsp://localhost:8554/mystream")
	if err != nil {
		panic(err)
	}

	// connect to the server
	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// find published medias
	medias, baseURL, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	// setup all medias
	err = c.SetupAll(medias, baseURL)
	if err != nil {
		panic(err)
	}

	// called when a RTP packet arrives
	c.OnPacketRTPAny(func(medi *media.Media, forma format.Format, pkt *rtp.Packet) {
		log.Printf("RTP packet from media %v\n", medi)
	})

	// called when a RTCP packet arrives
	c.OnPacketRTCPAny(func(medi *media.Media, pkt rtcp.Packet) {
		log.Printf("RTCP packet from media %v, type %T\n", medi, pkt)
	})

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	// wait until a fatal error
	panic(c.Wait())
}
