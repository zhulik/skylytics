package posts

import (
	"log"
	"skylytics/internal/core"

	"github.com/Jeffail/gabs"
	"github.com/bluesky-social/jetstream/pkg/models"
)

type PostReconstructor struct {
	Post *core.Post
}

func (r *PostReconstructor) AddEvent(event *models.Event) error {
	if r.Post == nil {
		r.Post = &core.Post{
			DID: event.Did,
		}
	}

	switch event.Commit.Operation {
	case models.CommitOperationCreate:
		container, err := gabs.ParseJSON(event.Commit.Record)
		if err != nil {
			return err
		}
		r.Post.Text = container.Path("text").Data().(string)
	default:
		log.Printf("%+v", event)
		panic("operation not yet supported")
	}
	return nil
}
