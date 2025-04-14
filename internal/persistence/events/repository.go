package events

import (
	"context"
	"errors"
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

	coll := client.Database("skylytics").Collection("events")
	_, err = coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys:    bson.D{{"did", 1}},
			Options: options.Index().SetUnique(true),
		}, {
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

func (r Repository) SaveRaw(ctx context.Context, raw []byte) error {
	var jsonData bson.M
	if err := bson.UnmarshalExtJSON(raw, false, &jsonData); err != nil {
		return err
	}
	_, err := r.coll.InsertOne(context.TODO(), jsonData)
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return err
		}
	}
	return nil
}

func (r Repository) Shutdown() error {
	return r.client.Disconnect(context.TODO())
}

func (r Repository) HealthCheck() error {
	return r.client.Ping(context.TODO(), nil)
}
