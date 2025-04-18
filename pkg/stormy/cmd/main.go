package main

import (
	"context"
	"github.com/k0kubun/pp"
	"log"
	"net/url"
	"resty.dev/v3"
	"skylytics/pkg/stormy"
)

func main() {
	client := stormy.NewClient(&stormy.ClientConfig{
		TransportSettings: stormy.DefaultConfig.TransportSettings,

		ResponseMiddlewares: []resty.ResponseMiddleware{func(client *resty.Client, response *resty.Response) error {
			reqURL, err := url.Parse(response.Request.URL)
			if err != nil {
				return err
			}

			log.Printf("%s %s: %s [%s]", response.Request.Method, reqURL.Path, response.Status(), response.Duration())
			return nil
		}},
	})
	defer client.Close()

	profiles, err := client.GetProfiles(context.Background(),
		"did:plc:tm23hobdjv2dbatpfwptntpo",
		"did:plc:7hx5jgngalmn6vrpvdmlkgtn")
	if err != nil {
		panic(err)
	}

	pp.Printf("%+v", profiles)
}
