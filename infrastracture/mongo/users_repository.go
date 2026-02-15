package infrastracture_mongo

import (
	"context"
	"fmt"
	"github.com/konradkar2/marcus_discord/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserMongo struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	UserId   string        `bson:"userId"`
	UserName string        `bson:"userName"`
}

func dtoToUserMongo(user *domain.User) UserMongo {
	return UserMongo{
		UserId:   user.UserId,
		UserName: user.UserName,
	}
}

type UsersRepository struct {
	coll_users *mongo.Collection
}

func (repo *UsersRepository) Insert(ctx context.Context, user domain.User) error {
	fmt.Printf("Adding user: %v\n", user)
	_, err := repo.coll_users.InsertOne(ctx, dtoToUserMongo(&user))

	return err
}

func NewUsersRepository(mongoDriver *MongoDriver) *UsersRepository {
	println("Creating UsersRepository")
	coll_users := mongoDriver.getCollection("users")

	return &UsersRepository{
		coll_users: coll_users}
}
