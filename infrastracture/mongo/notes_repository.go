package infrastracture_mongo

import (
	"context"
	"fmt"
	"github.com/konradkar2/marcus_discord/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type NoteMongo struct {
	ID         bson.ObjectID `bson:"_id"`
	Content    string        `bson:"content"`
	Number     int           `bson:"number"`
	BookNumber int           `bson:"bookNumber"`
}

func noteToDto(noteMongo *NoteMongo) domain.Note {
	return domain.Note{
		Content:    noteMongo.Content,
		Number:     noteMongo.Number,
		BookNumber: noteMongo.BookNumber,
	}
}

type NotesRepository struct {
	coll_notes *mongo.Collection
}

func (repo *NotesRepository) GetRandom(ctx context.Context) (domain.Note, error) {

	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := repo.coll_notes.Aggregate(ctx, pipeline)
	if err != nil {
		return domain.Note{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var noteMongo NoteMongo
		if err := cursor.Decode(&noteMongo); err != nil {
			return domain.Note{}, err
		}

		return noteToDto(&noteMongo), nil
	}

	return domain.Note{}, fmt.Errorf("no note found")
}

func NewNotesRepository(mongoDriver *MongoDriver) (*NotesRepository) {
	println("Creating NotesRepository")
	coll_notes := mongoDriver.getCollection("notes")
	
	return &NotesRepository{
		coll_notes: coll_notes}
}
