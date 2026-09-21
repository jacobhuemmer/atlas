package domain

import (
	"os"
	"testing"
)

func fixtureCatalog() Catalog {
	return Catalog{
		Sites: []Site{
			{Alias: "sesamidevel", Hostname: "sesamidevel.atlassian.net", UUID: "dd528466-a332-4f63-9cee-0033fe01f875", Role: RoleLicensed},
			{Alias: "sesami-io", Hostname: "sesami-io.atlassian.net", UUID: "8986e79a-6c2a-4e0d-bb65-4a57e0f68db9", Role: RoleLicensed},
			{Alias: "garda", Hostname: "gardaworld.atlassian.net", UUID: "62907a08-fd61-4a15-b642-f79d03e70b92", Role: RoleJSMCustomer},
		},
		Projects: map[string]string{
			"SDO": "sesamidevel",
			"SDP": "sesamidevel",
			"SES": "sesamidevel",
			"CAB": "sesami-io",
		},
		Spaces: map[string]string{"CCAB": "sesami-io"},
		Defaults: Defaults{
			Workspace:         "sesamiio",
			AssigneeAccountID: "712020:092d1246-2a54-4f40-8317-ef3dc777bf3c",
			JSMSite:           "garda",
		},
	}
}

func TestMain(m *testing.M) {
	SetCatalog(fixtureCatalog())
	os.Exit(m.Run())
}
