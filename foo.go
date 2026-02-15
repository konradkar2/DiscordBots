package main

import (
	"context"

	//"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DatabaseDriver struct {
	ctx           context.Context
	client*        mongo.Client
}

func (drv *DatabaseDriver) Close() {
	if err := drv.client.Disconnect(drv.ctx); err != nil {
		panic(err)
	}
}

func NewDatabaseDriver(connectionUri string) (*DatabaseDriver, error) {
	ctx := context.TODO()
	client, err := mongo.Connect(options.Client().
		ApplyURI(connectionUri))
	if err != nil {
		return nil, err
	}

	return &DatabaseDriver{client: client, ctx: ctx}, nil
}
