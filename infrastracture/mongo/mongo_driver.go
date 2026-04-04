package infrastracture_mongo

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDriver struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDriver(connectionUri string, databaseName string) (*MongoDriver, error) {
	println("Creating MongoDriver")
	client, err := mongo.Connect(options.Client().
		ApplyURI(connectionUri))

	if err != nil {
		return nil, err
	}

	database := client.Database(databaseName)
	return &MongoDriver{client: client, database: database}, nil
}

func (self *MongoDriver) getCollection(collection string) *mongo.Collection {
	return self.database.Collection(collection)
}
