package main

import (
	"log"

	"github.com/pion/rtp"
	"github.com/westpoint-robotics/gortsplib"
	"github.com/westpoint-robotics/gortsplib/pkg/format"
	"github.com/westpoint-robotics/gortsplib/pkg/media"
	"github.com/westpoint-robotics/gortsplib/pkg/url"
)

// This example shows how to
// 1. connect to a RTSP server and read all medias on a path
// 2. re-publish all medias on another path.

func main() {
	reader := gortsplib.Client{}

	// parse source URL
	sourceURL, err := url.Parse("rtsp://localhost:8554/mystream")
	if err != nil {
		panic(err)
	}

	// connect to the server
	err = reader.Start(sourceURL.Scheme, sourceURL.Host)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// find published medias
	medias, baseURL, _, err := reader.Describe(sourceURL)
	if err != nil {
		panic(err)
	}

	log.Printf("republishing %d medias", len(medias))

	// setup all medias
	// this must be called before StartRecording(), since it overrides the control attribute.
	err = reader.SetupAll(medias, baseURL)
	if err != nil {
		panic(err)
	}

	// connect to the server and start recording the same medias
	publisher := gortsplib.Client{}
	err = publisher.StartRecording("rtsp://localhost:8554/mystream2", medias)
	if err != nil {
		panic(err)
	}
	defer publisher.Close()

	// read RTP packets from the reader and route them to the publisher
	reader.OnPacketRTPAny(func(medi *media.Media, forma format.Format, pkt *rtp.Packet) {
		publisher.WritePacketRTP(medi, pkt)
	})

	// start playing
	_, err = reader.Play(nil)
	if err != nil {
		panic(err)
	}

	// wait until a fatal error
	panic(reader.Wait())
}
