package cfdns_test

import "github.com/calmh/cfdns"

func ExampleClient() {
	// Create a new Client
	client := cfdns.NewClient("test@example.com", "dc771129ae047c934adba2d1dc4fe184")

	// List all available zones
	zones, err := client.ListZones()
	if err != nil {
		panic(err)
	}

	// Find the "example.com" zone
	var exampleZone cfdns.Zone
	for _, zone := range zones {
		if zone.Name == "example.com" {
			exampleZone = zone
			break
		}
	}

	// If we found it ...
	if exampleZone.ID != "" {
		// ... create a new record "test.example.com A 192.168.0.2"
		if err := client.CreateDNSRecord(exampleZone.ID, "test.example.com", "A", "192.168.0.2"); err != nil {
			panic(err)
		}
	}
}
