package main

import (
	"log"

	"github.com/pion/rtp"
	"github.com/westpoint-robotics/gortsplib"
	"github.com/westpoint-robotics/gortsplib/pkg/format"
	"github.com/westpoint-robotics/gortsplib/pkg/formatdecenc/rtpvp9"
	"github.com/westpoint-robotics/gortsplib/pkg/url"
)

// This example shows how to
// 1. connect to a RTSP server
// 2. check if there's a VP9 media
// 3. get access units of that media

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

	// find the VP9 media and format
	var forma *format.VP9
	medi := medias.FindFormat(&forma)
	if medi == nil {
		panic("media not found")
	}

	// create decoder
	rtpDec := forma.CreateDecoder()

	// setup a single media
	_, err = c.Setup(medi, baseURL, 0, 0)
	if err != nil {
		panic(err)
	}

	// called when a RTP packet arrives
	c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		// extract VP9 frames from RTP packets
		vf, _, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtpvp9.ErrNonStartingPacketAndNoPrevious && err != rtpvp9.ErrMorePacketsNeeded {
				log.Printf("ERR: %v", err)
			}
			return
		}

		log.Printf("received frame of size %d\n", len(vf))
	})

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	// wait until a fatal error
	panic(c.Wait())
}
