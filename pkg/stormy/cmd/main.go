package main

import (
	"context"
	"github.com/k0kubun/pp"
	"skylytics/pkg/stormy"
)

func main() {
	client := stormy.NewClient()
	defer client.Close()

	profiles, err := client.GetProfiles(context.Background(),
		"did:plc:tm23hobdjv2dbatpfwptntpo",
		"did:plc:7hx5jgngalmn6vrpvdmlkgtn")
	if err != nil {
		panic(err)
	}

	pp.Printf("%+v", profiles)
}
