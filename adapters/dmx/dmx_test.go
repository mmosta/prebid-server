package dmx

import (
	"testing"

	"github.com/prebid/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "dmxtest", NewDmxBidder("https://mocktest.districtm.io/openrtb/not-a-real-endpont"))
}