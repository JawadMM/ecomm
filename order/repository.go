package order

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Close(ctx context.Context)
	PutOrder(ctx context.Context, o Order) error
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
	HasRecentOrder(ctx context.Context, accountID string, since time.Time) (bool, error)
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(url string) (*mongoRepository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}
	return &mongoRepository{
		client:     client,
		collection: client.Database("orders_db").Collection("orders"),
	}, nil
}

func (r *mongoRepository) Close(ctx context.Context) {
	r.client.Disconnect(ctx)
}

func (r *mongoRepository) PutOrder(ctx context.Context, o Order) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":         o.ID,
		"created_at":  o.CreatedAt,
		"account_id":  o.AccountID,
		"total_price": o.TotalPrice,
		"products":    o.Products,
	})
	return err
}

func (r *mongoRepository) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"account_id": accountID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *mongoRepository) HasRecentOrder(ctx context.Context, accountID string, since time.Time) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"account_id": accountID,
		"created_at": bson.M{"$gte": since},
	})
	return count > 0, err
}
