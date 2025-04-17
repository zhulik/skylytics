package events

import (
	"context"
	"errors"
	"os"
	"skylytics/pkg/async"

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
	_, err = coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys: bson.D{{"kind", 1}},
		}, {
			Keys: bson.D{{"did", 1}},
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

func (r Repository) InsertRaw(ctx context.Context, raws ...[]byte) ([]any, error) {
	datas := async.Map(raws, func(raw []byte) bson.M {
		var jsonData bson.M
		if err := bson.UnmarshalExtJSON(raw, false, &jsonData); err != nil {
			panic(err)
		}
		return jsonData
	})

	res, err := r.coll.InsertMany(ctx, datas)
	if err != nil {
		return nil, err
	}
	return res.InsertedIDs, nil
}

func (r Repository) HealthCheck() error {
	return r.client.Ping(context.Background(), nil)
}

func (r Repository) Shutdown() error {
	return r.client.Disconnect(context.Background())
}
