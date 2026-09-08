// Command templates lists the account's approved templates and sends one.
//
// The interesting part is TemplateParameterValue. A parameter the template
// declares as "number" may be sent either as a JSON number or as a JSON
// string, and the service substitutes a numeric string verbatim. That is the
// only way to keep a leading zero ("001234") or a trailing decimal zero
// ("1.50") — pass those as strings.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	adsefid "github.com/adsefid/sdk-go"
)

func main() {
	apiKey := os.Getenv("ADSEFID_API_KEY")
	lineNumber := os.Getenv("ADSEFID_LINE_NUMBER")
	if apiKey == "" || lineNumber == "" {
		log.Fatal("set ADSEFID_API_KEY and ADSEFID_LINE_NUMBER")
	}

	client, err := adsefid.NewClient(apiKey, adsefid.WithTimeout(30*time.Second))
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	template := firstApprovedTemplate(ctx, client)
	fmt.Printf("\nusing template %q\n  content: %s\n", template.TemplateID, template.Content)
	for name, kind := range template.Parameters {
		fmt.Printf("  parameter %-12s %s\n", name, kind)
	}

	// Build a value for each declared parameter. Note the two number-typed
	// parameters: one travels as a real JSON number, the other as a string so
	// its exact digits survive.
	parameters := map[string]adsefid.TemplateParameterValue{}
	for name, kind := range template.Parameters {
		switch kind {
		case adsefid.TemplateParameterTypeString:
			parameters[name] = adsefid.StringParam("Ali")
		case adsefid.TemplateParameterTypeNumber:
			parameters[name] = adsefid.IntParam(2)
		}
	}

	expiry := time.Now().Add(10 * time.Minute)
	resp, err := client.SMS.SendTemplate(ctx, &adsefid.SendTemplateSmsRequest{
		TemplateID: template.TemplateID,
		Parameters: parameters,
		Receptor:   "09120000000",
		LineNumber: lineNumber,
		ExpiryDate: &expiry,
	})
	if err != nil {
		log.Fatalf("send template: %v", err)
	}

	fmt.Printf("\nsent %s: %s\n  rendered: %s\n", resp.MessageID, resp.Status, resp.Message)
	fmt.Println("  parameters echoed back:")
	for name, value := range resp.Parameters {
		if raw, ok := value.RawNumber(); ok {
			// RawNumber is the lossless view — "1.50" comes back as "1.50",
			// not as the 1.5 a float64 would give you.
			fmt.Printf("    %-12s number %s\n", name, raw)
			continue
		}
		text, _ := value.StringValue()
		fmt.Printf("    %-12s string %q\n", name, text)
	}

	demonstrateExactValues()
}

func firstApprovedTemplate(ctx context.Context, client *adsefid.Client) adsefid.UserTemplate {
	approved := adsefid.TemplateStateApproved
	take := 100

	page, err := client.User.GetTemplates(ctx, &approved, nil, &take)
	if err != nil {
		log.Fatalf("list templates: %v", err)
	}
	if len(page.Items) == 0 {
		log.Fatal("no approved templates on this account — create one in the adsefid.com panel first")
	}

	fmt.Printf("%d approved template(s):\n", page.Total)
	for _, item := range page.Items {
		fmt.Printf("  %-24s %s\n", item.TemplateID, item.State)
	}
	return page.Items[0]
}

// demonstrateExactValues shows, without sending anything, which constructor to
// reach for when the exact digits matter.
func demonstrateExactValues() {
	fmt.Println("\nchoosing a parameter value:")
	for _, example := range []struct {
		why   string
		value adsefid.TemplateParameterValue
	}{
		{"an ordinary count", adsefid.IntParam(2)},
		{"a price where float rounding is fine", adsefid.NumberParam(19.99)},
		{"an invoice number whose leading zeros matter", adsefid.StringParam("001234")},
		{"an amount that must render as exactly 1.50", adsefid.StringParam("1.50")},
	} {
		encoded, err := example.value.MarshalJSON()
		if err != nil {
			log.Fatalf("marshal: %v", err)
		}
		fmt.Printf("  %-48s -> %s\n", example.why, encoded)
	}
}
