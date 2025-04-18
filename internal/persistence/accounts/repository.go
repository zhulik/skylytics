package accounts

import (
	"context"
	"os"
	"skylytics/internal/persistence"
	"skylytics/pkg/async"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"skylytics/internal/core"

	"github.com/samber/do"
)

type Repository struct {
	client *mongo.Client
	coll   *mongo.Collection
}

func NewRepository(_ *do.Injector) (core.AccountRepository, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, persistence.ErrNoMongodbURI
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	coll := client.Database("admin").Collection("accounts")
	_, err = coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{{
		Keys:    bson.D{{"did", 1}},
		Options: options.Index().SetUnique(true),
	}, {
		Keys:    bson.D{{"handle", 1}},
		Options: options.Index().SetUnique(true),
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
	datas, err := async.AsyncMap(nil, raws, func(_ context.Context, raw []byte) (bson.M, error) {
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
	return r.client.Ping(context.Background(), nil)
}

func (r Repository) Shutdown() error {
	return r.client.Disconnect(context.Background())
}
