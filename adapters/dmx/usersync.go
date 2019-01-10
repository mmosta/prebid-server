package dmx

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

const IAB_DISTRICTM = 144

func NewDmxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("dmx", IAB_DISTRICTM, temp, adapters.SyncTypeIframe)
}
