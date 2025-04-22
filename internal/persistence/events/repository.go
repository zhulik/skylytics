package events

import (
	"context"
	"os"
	"time"

	"skylytics/internal/persistence"
	"skylytics/pkg/async"

	"skylytics/internal/core"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/samber/do"
)

type Repository struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewRepository(_ *do.Injector) (core.EventRepository, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, persistence.ErrNoMongodbURI
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	coll := client.Database("admin").Collection("events")
	_, err = coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "kind", Value: 1}},
		}, {
			Keys: bson.D{{Key: "did", Value: 1}},
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
	datas, err := async.AsyncMap(ctx, raws, func(_ context.Context, raw []byte) (bson.M, error) {
		var jsonData bson.M
		return jsonData, bson.UnmarshalExtJSON(raw, false, &jsonData)
	})

	if err != nil {
		return nil, err
	}

	res, err := r.coll.InsertMany(ctx, datas)
	if err != nil {
		return nil, err
	}
	return res.InsertedIDs, nil
}

func (r Repository) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return r.client.Ping(ctx, nil)
}

func (r Repository) Shutdown() error {
	return r.client.Disconnect(context.Background())
}
