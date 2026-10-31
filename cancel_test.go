package adsefid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestCancellationCountsDistinctIDsAndPreservesBody(t *testing.T) {
	for _, area := range []string{"sms", "messenger"} {
		for _, tc := range []struct {
			messageCount int
			localCount   int
			repetitions  int
			rejected     bool
		}{
			{0, 0, 1, true},
			{2000, 0, 1, false},
			{2001, 0, 1, true},
			{0, 2000, 1, false},
			{0, 2001, 1, true},
			{1000, 1000, 1, false},
			{1000, 1001, 1, true},
			{1000, 1000, 3, false},
		} {
			t.Run(fmt.Sprintf("%s/%d+%d/repeated%d", area, tc.messageCount, tc.localCount, tc.repetitions), func(t *testing.T) {
				ids := func(count int) []string {
					var values []string
					for repeat := 0; repeat < tc.repetitions; repeat++ {
						for i := 0; i < count; i++ {
							values = append(values, fmt.Sprintf("00000000-0000-0000-0000-%012d", i))
						}
					}
					return values
				}
				messageIDs, localIDs := ids(tc.messageCount), ids(tc.localCount)
				client, requests := newFixtureClient(t, "envelopes/"+area+".cancel.success.json")
				var err error
				if area == "sms" {
					_, err = client.SMS.Cancel(context.Background(), &CancelSmsRequest{MessageIDs: messageIDs, LocalIDs: localIDs})
				} else {
					_, err = client.Messenger.Cancel(context.Background(), &CancelMessengerRequest{MessageIDs: messageIDs, LocalIDs: localIDs})
				}
				if tc.rejected {
					var validation *ValidationError
					if !errors.As(err, &validation) || len(*requests) != 0 {
						t.Fatalf("expected local validation with no request, got %v, requests=%d", err, len(*requests))
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				request := onlyRequest(t, requests)
				var body CancelSmsRequest
				if err := json.Unmarshal(request.Body, &body); err != nil {
					t.Fatal(err)
				}
				if request.Method != "POST" || request.Path != "/v1/"+area+"/cancel" ||
					!reflect.DeepEqual(body.MessageIDs, messageIDs) || !reflect.DeepEqual(body.LocalIDs, localIDs) {
					t.Fatal("cancellation changed the request IDs or endpoint")
				}
			})
		}
	}
}

func TestCancellationPreservesNullableIDsAndItemErrors(t *testing.T) {
	for _, area := range []string{"sms", "messenger"} {
		t.Run(area, func(t *testing.T) {
			client, requests := newFixtureClient(t, "envelopes/cancel.partial_success.json")
			var result any
			var err error
			if area == "sms" {
				var response *CancelSmsResponse
				response, err = client.SMS.Cancel(context.Background(), &CancelSmsRequest{LocalIDs: []string{"missing-local"}})
				if err != nil {
					t.Fatal(err)
				}
				code, ok := response.FailedToCancel[2].ErrorCode()
				if !ok || code != WebServiceResponseCodeInvalidLocalIDs {
					t.Fatal("missing typed local-ID error")
				}
				if status, ok := response.CancelledMessages[0].MessageStatus(); !ok || status != WebServiceMessageStatusCanceled {
					t.Fatal("missing cancelled status")
				}
				if _, ok := response.FailedToCancel[3].ErrorCode(); ok {
					t.Fatal("unknown error must remain raw")
				}
				result = response
			} else {
				var response *CancelMessengerResponse
				response, err = client.Messenger.Cancel(context.Background(), &CancelMessengerRequest{LocalIDs: []string{"missing-local"}})
				if err != nil {
					t.Fatal(err)
				}
				code, ok := response.FailedToCancel[2].ErrorCode()
				if !ok || code != WebServiceResponseCodeInvalidLocalIDs {
					t.Fatal("missing typed local-ID error")
				}
				if status, ok := response.CancelledMessages[0].MessageStatus(); !ok || status != WebServiceMessageStatusCanceled {
					t.Fatal("missing cancelled status")
				}
				if _, ok := response.FailedToCancel[3].ErrorCode(); ok {
					t.Fatal("unknown error must remain raw")
				}
				result = response
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var got any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			var fixture struct {
				Data any `json:"data"`
			}
			if err := json.Unmarshal(mustFixture(t, "envelopes/cancel.partial_success.json"), &fixture); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, fixture.Data) {
				t.Fatal("cancel outcomes must round-trip, including null IDs and unknown codes")
			}
			onlyRequest(t, requests)
		})
	}
}
