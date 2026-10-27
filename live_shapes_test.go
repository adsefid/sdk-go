package adsefid

import (
	"context"
	"testing"
)

// The fixtures below are shaped like real gateway responses: 7-digit fractional seconds,
// an empty-string description, null account fields, long decimal money and "+98" receptors.
func TestLiveShapedResponses(t *testing.T) {
	ctx := context.Background()

	t.Run("account info keeps null fields and the exact credit text", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/user.get_info.live.json")
		info, err := client.User.GetInfo(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if info.CompanyName != nil || info.Email != nil || info.CreditLeft.String() != "5877471300.00000" {
			t.Errorf("unexpected info: %+v", info)
		}
	})

	t.Run("templates parse 7-digit timestamps and an empty description", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/user.get_templates.live.json")
		got, err := client.User.GetTemplates(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		item := got.Items[0]
		if item.Description == nil || *item.Description != "" || item.CreatedAt.Nanosecond() != 403029200 {
			t.Errorf("unexpected template: %+v", item)
		}
	})

	t.Run("template send keeps long decimal cost and a normalized receptor", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.send_template.live.json")
		got, err := client.SMS.SendTemplate(ctx, &SendTemplateSmsRequest{TemplateID: "otp", Receptor: "989120000000", LineNumber: "983000"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Cost != "1129.0000000000000000" || got.Receptor != "+989120000000" || got.ExpiryDate != nil || got.SendTime == nil || got.SendTime.Nanosecond() != 921182500 {
			t.Errorf("unexpected response: %+v", got)
		}
	})

	t.Run("status parses a 7-digit delivery time", func(t *testing.T) {
		client, _ := newFixtureClient(t, "envelopes/sms.get_status.live.json")
		got, err := client.SMS.GetStatus(ctx, &GetSmsStatusRequest{LocalIDs: []string{"order-10001"}})
		if err != nil {
			t.Fatal(err)
		}
		item := got.Receptors[0]
		if item.Status != WebServiceMessageStatusUndelivered || item.DeliveryTime == nil || item.DeliveryTime.Nanosecond() != 953820400 {
			t.Errorf("unexpected status item: %+v", item)
		}
	})
}
