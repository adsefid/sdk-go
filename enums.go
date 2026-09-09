package adsefid

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LineSelector selects which line-accounting mode a send should use. See the
// adsefid.com API documentation (doc v1.12.0, section 3.1).
type LineSelector int

const (
	LineSelectorPromotionalSendBased            LineSelector = 0
	LineSelectorPromotionalDeliverBased         LineSelector = 1
	LineSelectorBulkServiceSendBased            LineSelector = 2
	LineSelectorBulkServiceDeliverBased         LineSelector = 3
	LineSelectorCustomerClubServiceSendBased    LineSelector = 4
	LineSelectorCustomerClubServiceDeliverBased LineSelector = 5
)

func (v LineSelector) String() string {
	switch v {
	case LineSelectorPromotionalSendBased:
		return "PromotionalSendBased"
	case LineSelectorPromotionalDeliverBased:
		return "PromotionalDeliverBased"
	case LineSelectorBulkServiceSendBased:
		return "BulkServiceSendBased"
	case LineSelectorBulkServiceDeliverBased:
		return "BulkServiceDeliverBased"
	case LineSelectorCustomerClubServiceSendBased:
		return "CustomerClubServiceSendBased"
	case LineSelectorCustomerClubServiceDeliverBased:
		return "CustomerClubServiceDeliverBased"
	default:
		return fmt.Sprintf("LineSelector(%d)", int(v))
	}
}

// WebServiceMessageStatus is the status of a single SMS or Messenger message,
// as reported by the get-status endpoints, the send endpoints, and webhook
// status-update payloads. Values 1000-1999. See doc section 3.2.
type WebServiceMessageStatus int

const (
	WebServiceMessageStatusScheduled          WebServiceMessageStatus = 1000
	WebServiceMessageStatusSending            WebServiceMessageStatus = 1001
	WebServiceMessageStatusDelivered          WebServiceMessageStatus = 1002
	WebServiceMessageStatusUndelivered        WebServiceMessageStatus = 1003
	WebServiceMessageStatusCanceled           WebServiceMessageStatus = 1004
	WebServiceMessageStatusSentToOperator     WebServiceMessageStatus = 1005
	WebServiceMessageStatusBlacklisted        WebServiceMessageStatus = 1006
	WebServiceMessageStatusProviderError      WebServiceMessageStatus = 1007
	WebServiceMessageStatusPendingApproval    WebServiceMessageStatus = 1008
	WebServiceMessageStatusRejected           WebServiceMessageStatus = 1009
	WebServiceMessageStatusInvalidSender      WebServiceMessageStatus = 1010
	WebServiceMessageStatusInvalidAttachment  WebServiceMessageStatus = 1011
	WebServiceMessageStatusForbiddenWord      WebServiceMessageStatus = 1012
	WebServiceMessageStatusLinkNotAllowed     WebServiceMessageStatus = 1013
	WebServiceMessageStatusInvalidReceiver    WebServiceMessageStatus = 1014
	WebServiceMessageStatusUndeliverable      WebServiceMessageStatus = 1015
	WebServiceMessageStatusSenderLimitReached WebServiceMessageStatus = 1016
	WebServiceMessageStatusUnknown            WebServiceMessageStatus = 1999
)

func (v WebServiceMessageStatus) String() string {
	switch v {
	case WebServiceMessageStatusScheduled:
		return "Scheduled"
	case WebServiceMessageStatusSending:
		return "Sending"
	case WebServiceMessageStatusDelivered:
		return "Delivered"
	case WebServiceMessageStatusUndelivered:
		return "Undelivered"
	case WebServiceMessageStatusCanceled:
		return "Canceled"
	case WebServiceMessageStatusSentToOperator:
		return "SentToOperator"
	case WebServiceMessageStatusBlacklisted:
		return "Blacklisted"
	case WebServiceMessageStatusProviderError:
		return "ProviderError"
	case WebServiceMessageStatusPendingApproval:
		return "PendingApproval"
	case WebServiceMessageStatusRejected:
		return "Rejected"
	case WebServiceMessageStatusInvalidSender:
		return "InvalidSender"
	case WebServiceMessageStatusInvalidAttachment:
		return "InvalidAttachment"
	case WebServiceMessageStatusForbiddenWord:
		return "ForbiddenWord"
	case WebServiceMessageStatusLinkNotAllowed:
		return "LinkNotAllowed"
	case WebServiceMessageStatusInvalidReceiver:
		return "InvalidReceiver"
	case WebServiceMessageStatusUndeliverable:
		return "Undeliverable"
	case WebServiceMessageStatusSenderLimitReached:
		return "SenderLimitReached"
	case WebServiceMessageStatusUnknown:
		return "Unknown"
	default:
		return fmt.Sprintf("WebServiceMessageStatus(%d)", int(v))
	}
}

// WebServiceResponseCode is the API's error code, carried in the error
// envelope's error.code field. Values 2000-2045. See doc section 3.4.
type WebServiceResponseCode int

const (
	WebServiceResponseCodeInternalError                 WebServiceResponseCode = 2000
	WebServiceResponseCodeInvalidPlan                   WebServiceResponseCode = 2001
	WebServiceResponseCodeLineNotFound                  WebServiceResponseCode = 2002
	WebServiceResponseCodeTooManyReceptors              WebServiceResponseCode = 2003
	WebServiceResponseCodeInvalidLine                   WebServiceResponseCode = 2004
	WebServiceResponseCodeInvalidAPIKey                 WebServiceResponseCode = 2005
	WebServiceResponseCodeIPNotAllowed                  WebServiceResponseCode = 2006
	WebServiceResponseCodeDuplicateLocalID              WebServiceResponseCode = 2007
	WebServiceResponseCodeUserInformationNotFound       WebServiceResponseCode = 2008
	WebServiceResponseCodeEmptyReceptors                WebServiceResponseCode = 2009
	WebServiceResponseCodeInvalidReceptors              WebServiceResponseCode = 2010
	WebServiceResponseCodeEmptyBody                     WebServiceResponseCode = 2011
	WebServiceResponseCodeEmptyLine                     WebServiceResponseCode = 2012
	WebServiceResponseCodeEmptyMessage                  WebServiceResponseCode = 2013
	WebServiceResponseCodeInvalidReceptor               WebServiceResponseCode = 2014
	WebServiceResponseCodeEmptyReceptor                 WebServiceResponseCode = 2015
	WebServiceResponseCodeMessageTooLarge               WebServiceResponseCode = 2016
	WebServiceResponseCodeInvalidLineSelector           WebServiceResponseCode = 2017
	WebServiceResponseCodeUnauthorized                  WebServiceResponseCode = 2018
	WebServiceResponseCodeInvalidSendRange              WebServiceResponseCode = 2019
	WebServiceResponseCodeAllReceptorsBlacklisted       WebServiceResponseCode = 2020
	WebServiceResponseCodeMessageContainsForbiddenWords WebServiceResponseCode = 2021
	WebServiceResponseCodeNotEnoughCredit               WebServiceResponseCode = 2022
	WebServiceResponseCodeDuplicateTag                  WebServiceResponseCode = 2023
	WebServiceResponseCodeInvalidParameter              WebServiceResponseCode = 2024
	WebServiceResponseCodeReceptorBlacklisted           WebServiceResponseCode = 2025
	WebServiceResponseCodeInvalidLinkInMessage          WebServiceResponseCode = 2026
	WebServiceResponseCodeTemplateNotApproved           WebServiceResponseCode = 2027
	WebServiceResponseCodeInvalidTemplateParameter      WebServiceResponseCode = 2028
	WebServiceResponseCodeInvalidLocalIDs               WebServiceResponseCode = 2029
	WebServiceResponseCodeEmptyLocalIDs                 WebServiceResponseCode = 2030
	WebServiceResponseCodeEmptyMessageIDs               WebServiceResponseCode = 2031
	WebServiceResponseCodeInvalidSmsType                WebServiceResponseCode = 2032
	WebServiceResponseCodeLineNotActive                 WebServiceResponseCode = 2033
	WebServiceResponseCodeLineExpired                   WebServiceResponseCode = 2034
	WebServiceResponseCodeMessageLimitReached           WebServiceResponseCode = 2035
	WebServiceResponseCodeRequestLimitReached           WebServiceResponseCode = 2036
	WebServiceResponseCodeInvalidSendTime               WebServiceResponseCode = 2037
	WebServiceResponseCodeInvalidExpiry                 WebServiceResponseCode = 2038
	WebServiceResponseCodeInvalidTemplateID             WebServiceResponseCode = 2039
	WebServiceResponseCodeProfileNotFound               WebServiceResponseCode = 2040
	WebServiceResponseCodeProfileExpired                WebServiceResponseCode = 2041
	WebServiceResponseCodeFileNotFound                  WebServiceResponseCode = 2042
	WebServiceResponseCodeInvalidFile                   WebServiceResponseCode = 2043
	WebServiceResponseCodeAccessDenied                  WebServiceResponseCode = 2044
	WebServiceResponseCodeRejected                      WebServiceResponseCode = 2045

	// webServiceResponseCodeUnknown is used internally when an error HTTP
	// response could not be decoded as an error envelope at all (e.g. a bare
	// 429 or 5xx with no JSON body).
	webServiceResponseCodeUnknown WebServiceResponseCode = -1
)

func (v WebServiceResponseCode) String() string {
	switch v {
	case WebServiceResponseCodeInternalError:
		return "InternalError"
	case WebServiceResponseCodeInvalidPlan:
		return "InvalidPlan"
	case WebServiceResponseCodeLineNotFound:
		return "LineNotFound"
	case WebServiceResponseCodeTooManyReceptors:
		return "TooManyReceptors"
	case WebServiceResponseCodeInvalidLine:
		return "InvalidLine"
	case WebServiceResponseCodeInvalidAPIKey:
		return "InvalidApiKey"
	case WebServiceResponseCodeIPNotAllowed:
		return "IpNotAllowed"
	case WebServiceResponseCodeDuplicateLocalID:
		return "DuplicateLocalId"
	case WebServiceResponseCodeUserInformationNotFound:
		return "UserInformationNotFound"
	case WebServiceResponseCodeEmptyReceptors:
		return "EmptyReceptors"
	case WebServiceResponseCodeInvalidReceptors:
		return "InvalidReceptors"
	case WebServiceResponseCodeEmptyBody:
		return "EmptyBody"
	case WebServiceResponseCodeEmptyLine:
		return "EmptyLine"
	case WebServiceResponseCodeEmptyMessage:
		return "EmptyMessage"
	case WebServiceResponseCodeInvalidReceptor:
		return "InvalidReceptor"
	case WebServiceResponseCodeEmptyReceptor:
		return "EmptyReceptor"
	case WebServiceResponseCodeMessageTooLarge:
		return "MessageTooLarge"
	case WebServiceResponseCodeInvalidLineSelector:
		return "InvalidLineSelector"
	case WebServiceResponseCodeUnauthorized:
		return "Unauthorized"
	case WebServiceResponseCodeInvalidSendRange:
		return "InvalidSendRange"
	case WebServiceResponseCodeAllReceptorsBlacklisted:
		return "AllReceptorsBlacklisted"
	case WebServiceResponseCodeMessageContainsForbiddenWords:
		return "MessageContainsForbiddenWords"
	case WebServiceResponseCodeNotEnoughCredit:
		return "NotEnoughCredit"
	case WebServiceResponseCodeDuplicateTag:
		return "DuplicateTag"
	case WebServiceResponseCodeInvalidParameter:
		return "InvalidParameter"
	case WebServiceResponseCodeReceptorBlacklisted:
		return "ReceptorBlacklisted"
	case WebServiceResponseCodeInvalidLinkInMessage:
		return "InvalidLinkInMessage"
	case WebServiceResponseCodeTemplateNotApproved:
		return "TemplateNotApproved"
	case WebServiceResponseCodeInvalidTemplateParameter:
		return "InvalidTemplateParameter"
	case WebServiceResponseCodeInvalidLocalIDs:
		return "InvalidLocalIds"
	case WebServiceResponseCodeEmptyLocalIDs:
		return "EmptyLocalIds"
	case WebServiceResponseCodeEmptyMessageIDs:
		return "EmptyMessageIds"
	case WebServiceResponseCodeInvalidSmsType:
		return "InvalidSmsType"
	case WebServiceResponseCodeLineNotActive:
		return "LineNotActive"
	case WebServiceResponseCodeLineExpired:
		return "LineExpired"
	case WebServiceResponseCodeMessageLimitReached:
		return "MessageLimitReached"
	case WebServiceResponseCodeRequestLimitReached:
		return "RequestLimitReached"
	case WebServiceResponseCodeInvalidSendTime:
		return "InvalidSendTime"
	case WebServiceResponseCodeInvalidExpiry:
		return "InvalidExpiry"
	case WebServiceResponseCodeInvalidTemplateID:
		return "InvalidTemplateId"
	case WebServiceResponseCodeProfileNotFound:
		return "ProfileNotFound"
	case WebServiceResponseCodeProfileExpired:
		return "ProfileExpired"
	case WebServiceResponseCodeFileNotFound:
		return "FileNotFound"
	case WebServiceResponseCodeInvalidFile:
		return "InvalidFile"
	case WebServiceResponseCodeAccessDenied:
		return "AccessDenied"
	case WebServiceResponseCodeRejected:
		return "Rejected"
	case webServiceResponseCodeUnknown:
		return "UnknownError"
	default:
		return fmt.Sprintf("WebServiceResponseCode(%d)", int(v))
	}
}

// A WebServiceCode (doc section 3.3) is the raw per-item status in a bulk or
// P2P send response: 1000-1999 is a WebServiceMessageStatus (the item was
// accepted), 2000 and above is a WebServiceResponseCode (that one item was
// rejected even though the response as a whole succeeded).
const (
	webServiceMessageStatusMin = 1000
	webServiceResponseCodeMin  = 2000
)

// MessageStatusOf splits a raw per-item WebServiceCode: it returns the
// WebServiceMessageStatus and true when code is in 1000-1999 and names a
// status this SDK knows, otherwise 0 and false.
func MessageStatusOf(code int) (WebServiceMessageStatus, bool) {
	if code < webServiceMessageStatusMin || code >= webServiceResponseCodeMin {
		return 0, false
	}
	status := WebServiceMessageStatus(code)
	if strings.HasPrefix(status.String(), "WebServiceMessageStatus(") {
		return 0, false
	}
	return status, true
}

// ErrorCodeOf splits a raw per-item WebServiceCode: it returns the
// WebServiceResponseCode and true when code is 2000 or above and names an
// error this SDK knows, otherwise 0 and false.
func ErrorCodeOf(code int) (WebServiceResponseCode, bool) {
	if code < webServiceResponseCodeMin {
		return 0, false
	}
	responseCode := WebServiceResponseCode(code)
	if strings.HasPrefix(responseCode.String(), "WebServiceResponseCode(") {
		return 0, false
	}
	return responseCode, true
}

// TemplateState is the moderation state of a user-defined message template.
// See doc section 3.5.
type TemplateState string

const (
	TemplateStatePendingApproval TemplateState = "pendingapproval"
	TemplateStateApproved        TemplateState = "approved"
	TemplateStateRejected        TemplateState = "rejected"
)

func (v TemplateState) String() string {
	if v == "" {
		return "TemplateState(\"\")"
	}
	return string(v)
}

// TemplateParameterType is the declared type of one template parameter, as
// reported by GetTemplates. See doc section 3.6.
//
// The documented, supported public set is exactly {string, number}. The live
// API has been observed to also emit a third, undocumented value for some
// templates; this SDK intentionally does not model or expose it. If you need
// to detect it, compare the raw wire value defensively rather than assuming
// it is always one of the two constants below.
type TemplateParameterType string

const (
	TemplateParameterTypeString TemplateParameterType = "string"
	TemplateParameterTypeNumber TemplateParameterType = "number"
)

func (v TemplateParameterType) String() string {
	switch v {
	case TemplateParameterTypeString:
		return "String"
	case TemplateParameterTypeNumber:
		return "Number"
	default:
		return fmt.Sprintf("TemplateParameterType(%q)", string(v))
	}
}

// TemplateParameterValue is the value bound to a single named template
// parameter, sent when sending a templated SMS or Messenger message and
// received back on the response. The API accepts either a JSON string or a
// JSON number for a template parameter, so this type marshals and unmarshals
// as whichever of the two it currently holds.
//
// For a parameter the template declares as "number", pass a string whenever
// the exact lexical form matters — StringParam("001234") keeps the leading
// zeros, and StringParam("1.50") keeps the trailing zero. The server accepts a
// numeric string for a number parameter and substitutes it verbatim. Numbers
// themselves may be decimal.
type TemplateParameterValue struct {
	stringValue string
	numberValue json.Number
	isNumber    bool
}

// StringParam builds a TemplateParameterValue holding a string.
func StringParam(value string) TemplateParameterValue {
	return TemplateParameterValue{stringValue: value}
}

// NumberParam builds a TemplateParameterValue holding a floating-point number.
// Note that a float64 cannot represent every decimal exactly; use StringParam
// when the exact decimal form matters.
func NumberParam(value float64) TemplateParameterValue {
	return TemplateParameterValue{
		numberValue: json.Number(strconv.FormatFloat(value, 'f', -1, 64)),
		isNumber:    true,
	}
}

// IntParam builds a TemplateParameterValue holding an integer.
func IntParam(value int64) TemplateParameterValue {
	return TemplateParameterValue{
		numberValue: json.Number(strconv.FormatInt(value, 10)),
		isNumber:    true,
	}
}

// IsNumber reports whether this value holds a number (as opposed to a string).
func (v TemplateParameterValue) IsNumber() bool {
	return v.isNumber
}

// StringValue returns the held string and true, or "" and false if this value
// holds a number instead.
func (v TemplateParameterValue) StringValue() (string, bool) {
	if v.isNumber {
		return "", false
	}
	return v.stringValue, true
}

// NumberValue returns the held number as a float64 and true, or 0 and false if
// this value holds a string instead. Use RawNumber to read the number without
// going through binary floating point.
func (v TemplateParameterValue) NumberValue() (float64, bool) {
	if !v.isNumber {
		return 0, false
	}
	parsed, err := v.numberValue.Float64()
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// RawNumber returns the held number as its exact decimal text and true, or ""
// and false if this value holds a string instead. This is the lossless view: a
// number read off the wire round-trips through it unchanged.
func (v TemplateParameterValue) RawNumber() (json.Number, bool) {
	if !v.isNumber {
		return "", false
	}
	return v.numberValue, true
}

// MarshalJSON implements json.Marshaler.
func (v TemplateParameterValue) MarshalJSON() ([]byte, error) {
	if v.isNumber {
		return []byte(v.numberValue), nil
	}
	return json.Marshal(v.stringValue)
}

// UnmarshalJSON implements json.Unmarshaler. It accepts a JSON string or a
// JSON number and rejects anything else. A number is retained as its exact
// wire text rather than being converted to float64.
func (v *TemplateParameterValue) UnmarshalJSON(data []byte) error {
	// Deliberately not the conventional no-op on null: a null parameter value
	// is not something the API accepts, and silently decoding it to an empty
	// string would send a wrong value rather than surface the problem.
	if string(data) == "null" {
		return fmt.Errorf("adsefid: a template parameter value must be a JSON string or number, got null")
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*v = TemplateParameterValue{stringValue: asString}
		return nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(data, &asNumber); err == nil {
		*v = TemplateParameterValue{numberValue: asNumber, isNumber: true}
		return nil
	}

	return fmt.Errorf("adsefid: a template parameter value must be a JSON string or number, got %q", string(data))
}
