package events

import (
	"context"
	"errors"
	"github.com/samber/lo"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"skylytics/internal/core"

	"github.com/samber/do"
)

var (
	ErrNoMongodbURI = errors.New("no MONGODB_URI env provided")
)

type Repository struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewRepository(_ *do.Injector) (core.EventRepository, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, ErrNoMongodbURI
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	coll := client.Database("admin").Collection("events")
	_, err = coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{{
		Keys: bson.D{{"kind", 1}},
	},
	})

	if err != nil {
		return nil, err
	}

	return Repository{
		client: client,
		coll:   coll,
	}, nil
}

func (r Repository) SaveRaw(ctx context.Context, raws ...[]byte) error {
	datas := lo.Map(raws, func(raw []byte, _ int) bson.M {
		var jsonData bson.M
		if err := bson.UnmarshalExtJSON(raw, false, &jsonData); err != nil {
			panic(err)
		}
		return jsonData
	})

	res, err := r.coll.InsertMany(ctx, datas)
	if err != nil {
		return err
	}
	log.Printf("Saved %d events: ", len(res.InsertedIDs))
	return nil
}

func (r Repository) Shutdown() error {
	return r.client.Disconnect(context.TODO())
}

func (r Repository) HealthCheck() error {
	return r.client.Ping(context.TODO(), nil)
}
