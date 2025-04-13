package core

import "time"

type Commit struct {
	Type      string    `json:"$type"`
	CreatedAt time.Time `json:"createdAt"`
}
