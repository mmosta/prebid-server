package openrtb_ext

// ExtImpDmx defines the contract for bidrequest.imp[i].ext.dmx
type ExtImpDmx struct {
	DmxID    int `json:"dmxid"`
	MemberID int `json:"memberid"`
}
