package nats

import "github.com/zhulik/pal"

func Provide() pal.ServiceDef {
	return pal.ProvideList(
		pal.Provide(&NATS{}),
	)
}
