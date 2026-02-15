package infrastracture_mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/konradkar2/marcus_discord/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SubscriptionMongo struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	UserId   string        `bson:"userId"`
	NextSendAt time.Time   `bson:"nextSendAt"`
}

func dtoToSubscriptionMongo(sub *domain.Subscription) SubscriptionMongo {
	return SubscriptionMongo{
		UserId:   sub.UserId,
		NextSendAt: sub.NextSendAt}
}

func SubscriptionMongoToDto(sub *SubscriptionMongo) domain.Subscription {
	return  domain.Subscription{
		UserId:   sub.UserId,
		NextSendAt: sub.NextSendAt}
}



type SubscriptionRepository struct {
	coll_subs *mongo.Collection
}

func (repo *SubscriptionRepository) Insert(ctx context.Context, subscription domain.Subscription) error {
	fmt.Printf("Adding sub: %v\n", subscription)
	_, err := repo.coll_subs.InsertOne(ctx, dtoToSubscriptionMongo(&subscription))

	return err
}

func NewSubscriptionRepository(mongoDriver *MongoDriver) *SubscriptionRepository {
	println("Creating SubscriptionRepository")
	coll_subs := mongoDriver.getCollection("subscriptions")

	return &SubscriptionRepository{
		coll_subs: coll_subs}
}

func (repo *SubscriptionRepository) FindDue(ctx context.Context, now time.Time) ([]domain.Subscription, error){
	filter := bson.M{
		"nextSendAt": bson.M{"$lte": now},
	}

	opts := options.Find().SetLimit(100)
	cursor, err := repo.coll_subs.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []domain.Subscription

	if cursor.Next(ctx) {
		var mongoSub SubscriptionMongo
		if err := cursor.Decode(&mongoSub); err != nil {
			return nil, err
		}

		results = append(results, SubscriptionMongoToDto(&mongoSub))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (repo *SubscriptionRepository) UpdateSubscription(ctx context.Context, userId string, nextSendAt time.Time) error {
	filter := bson.M{"userId": userId}

	update := bson.M{
		"$set": bson.M{
			"nextSendAt": nextSendAt,
		},
	}
	
	opts := options.UpdateOne().SetUpsert(true)
	_, err := repo.coll_subs.UpdateOne(ctx, filter, update, opts)
	return err
}
