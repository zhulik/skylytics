package core

import "github.com/orsinium-labs/enum"

type Interaction enum.Member[string]

var (
	Like            = Interaction{"like"}
	Repost          = Interaction{"repost"}
	Reply           = Interaction{"reply"}
	Quote           = Interaction{"quote"}
	AllInteractions = enum.New(Like, Repost, Reply, Quote)
)
