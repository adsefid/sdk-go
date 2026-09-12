package adsefid

// errorEnvelope is the outer shape of every failed API response:
// {"status":"error","error":{"code":...,"name":"...","details":{...}}}.
type errorEnvelope struct {
	Status string        `json:"status"`
	Error  *errorPayload `json:"error"`
}

// errorPayload is the inner "error" object of errorEnvelope.
type errorPayload struct {
	Code    int              `json:"code"`
	Name    string           `json:"name"`
	Details *APIErrorDetails `json:"details"`
}
