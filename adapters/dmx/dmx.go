package dmx

import (
	"encoding/json"
	"fmt"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"net/http"
	"reflect"
	"strconv"
)

type DmxAdapter struct {
	URI string
}

func (a *DmxAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {

	publisherIds := make(map[string]bool)
	errs := make([]error, 0, len(request.Imp))

	if len(request.Imp) == 0 {
		err := &errortypes.BadInput{
			Message: "Empty BidRequest.Imp[]",
		}
		errs = append(errs, err)
		return nil, errs
	}

	for i := 0; i < len(request.Imp); i++ {

		pubId, err := preprocess(&request.Imp[i])

		if pubId != "" {
			publisherIds[pubId] = true
		}

		// If the preprocessing failed, the server won't be able to bid on this Imp. Delete it, and note the error.
		if err != nil {
			errs = append(errs, err)
			request.Imp = append(request.Imp[:i], request.Imp[i+1:]...)
			i--
		}
	}

	if len(publisherIds) != 1 {
		errs = append(errs, fmt.Errorf("All request.imp[i].ext.dmx.memberid params must both exist and match. Request contained: %v", len(publisherIds)))
		return nil, errs
	}

	if len(request.Imp) == 0 {
		err := &errortypes.BadInput{
			Message: fmt.Sprintf("No valid impression in the bid request"),
		}
		errs = append(errs, err)
		return nil, errs
	}

	// Set auction type to first price
	request.AT = 1

	request.Site.Publisher = &openrtb.Publisher{ID: reflect.ValueOf(publisherIds).MapKeys()[0].String()}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	headers.Add("x-openrtb-version", "2.5")

	if request.Device != nil {
		addHeaderIfNonEmpty(headers, "User-Agent", request.Device.UA)
		addHeaderIfNonEmpty(headers, "X-Forwarded-For", request.Device.IP)
		addHeaderIfNonEmpty(headers, "Accept-Language", request.Device.Language)
		addHeaderIfNonEmpty(headers, "DNT", strconv.Itoa(int(request.Device.DNT)))
	}

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.URI,
		Body:    reqJSON,
		Headers: headers,
	}}, errs
}

func (a *DmxAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Please ensure input parameters are correct for your account", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("bad server response: unable to parse body"),
		}}
	}

	impMap := make(map[string]*openrtb.Banner)

	for i := 0; i < len(internalRequest.Imp); i++ {
		impMap[internalRequest.Imp[i].ID] = internalRequest.Imp[i].Banner
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(len(bidResp.SeatBid))

	var errs []error
	for _, sb := range bidResp.SeatBid {
		for i := 0; i < len(sb.Bid); i++ {
			bid := sb.Bid[i]
			handleDefaultSize(impMap, &bid)
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &bid,
				BidType: openrtb_ext.BidTypeBanner,
			})
		}
	}
	return bidResponse, errs
}

func handleDefaultSize(m map[string]*openrtb.Banner, bid *openrtb.Bid) {
	if !(bid.W > 0) || !(bid.H > 0) {
		bid.W = *m[bid.ImpID].W
		bid.H = *m[bid.ImpID].H
	}
}

func preprocess(imp *openrtb.Imp) (string, error) {

	if imp.Banner == nil {
		return "", &errortypes.BadInput{
			Message: "DMX only supports banners at this stage",
		}
	}
	var extension adapters.ExtImpBidder
	err := json.Unmarshal(imp.Ext, &extension)

	if err != nil {
		return "", &errortypes.BadInput{
			Message: "ext.bidder not provided",
		}
	}

	var dmxExt openrtb_ext.ExtImpDmx

	err = json.Unmarshal(extension.Bidder, &dmxExt)
	if err != nil {
		return "", &errortypes.BadInput{
			Message: "ext.bidder (dmx) not provided",
		}
	}

	if dmxExt.MemberID == 0 {
		return "", &errortypes.BadInput{
			Message: "ext.dmx.memberid is missing",
		}
	}

	// TagID is optional by the client, no need for an error
	if dmxExt.DmxID != 0 {
		imp.TagID = strconv.Itoa(dmxExt.DmxID)
	}

	bannerCopy := *imp.Banner

	if bannerCopy.W == nil && bannerCopy.H == nil && len(bannerCopy.Format) > 0 {
		firstFormat := bannerCopy.Format[0]
		bannerCopy.W = &(firstFormat.W)
		bannerCopy.H = &(firstFormat.H)
	}
	imp.Banner = &bannerCopy
	imp.Ext = nil

	return strconv.Itoa(dmxExt.MemberID), nil

}

//Adding header fields to request header
func addHeaderIfNonEmpty(headers http.Header, headerName string, headerValue string) {
	if len(headerValue) > 0 {
		headers.Add(headerName, headerValue)
	}
}

func NewDmxBidder(endpoint string) *DmxAdapter {
	return &DmxAdapter{
		URI: endpoint,
	}
}
