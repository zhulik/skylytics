package stormy

import (
	"context"
	"net/url"
	"time"
)

const (
	getProfiles = "/xrpc/app.bsky.actor.getProfiles"
)

// https://docs.bsky.app/docs/api/app-bsky-actor-get-profile
type Profile struct {
	DID    string `json:"did"`
	Handle string `json:"handle"`

	DisplayName string `json:"displayName"`
	Description string `json:"description"`

	Avatar string `json:"avatar"`
	Banner string `json:"banner"`

	IndexedAt time.Time `json:"indexedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// https://docs.bsky.app/docs/api/app-bsky-actor-get-profiles
// max=25
func (c *Client) GetProfiles(ctx context.Context, actors ...string) ([]*Profile, error) {
	type Profiles struct {
		Profiles []*Profile `json:"profiles"`
	}

	res, err := c.r(ctx).
		SetQueryParamsFromValues(url.Values{
			"actors": actors,
		}).
		SetResult(&Profiles{}).
		Get(baseURL + getProfiles)

	if err != nil {
		return nil, err
	}
	return res.Result().(*Profiles).Profiles, nil
}
