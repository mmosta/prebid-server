package dmx

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestDmxSyncer(t *testing.T) {
	temp := template.Must(template.New("sync-template").Parse("https://cdn.districtm.io/ids/?sellerid=5&gdpr={{.GDPR}}&gdpr_consent={{.GDPRConsent}}"))
	syncer := NewDmxSyncer(temp)
	u, err := syncer.GetUsersyncInfo("1", "CONSENT")
	assert.NoError(t, err)
	assert.Equal(t, "https://cdn.districtm.io/ids/?sellerid=5&gdpr=1&gdpr_consent=CONSENT", u.URL)
	assert.Equal(t, "iframe", u.Type)
	assert.Equal(t, uint16(144), syncer.GDPRVendorID())
	assert.Equal(t, false, u.SupportCORS)
}
