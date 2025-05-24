package stormy

import (
	"context"
	"net/url"
)

const (
	getPosts = "/xrpc/app.bsky.actor.getProfiles"
)

// https://docs.bsky.app/docs/api/app-bsky-actor-get-profile
type Post struct {
}

// https://docs.bsky.app/docs/api/app-bsky-feed-get-posts
// max=25
func (c *Client) GetPosts(ctx context.Context, uris ...string) ([]*Post, error) {
	type Posts struct {
		Posts []*Post `json:"posts"`
	}

	res, err := c.r(ctx).
		SetQueryParamsFromValues(url.Values{
			"uris": uris,
		}).
		SetResult(&Posts{}).
		Get(getPosts)

	if err != nil {
		return nil, err
	}
	return res.Result().(*Posts).Posts, nil
}
