package helper

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"
	"log"
	"os"
	"time"
)

type DatabaseHelper interface {
	Query(string, string, string, interface{}) error
	QueryByUid(string, string, primitive.ObjectID, interface{}) error
	Insert(string, interface{}) error
	Delete(string, interface{}) error
	Increment(string, string, string, string, int) error
}

type MongoDBHelper struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDatabase() DatabaseHelper {

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	db := client.Database(os.Getenv("MONGO_DATABASE"))
	return &MongoDBHelper{
		client: client,
		db:     db,
	}
}

func (mdb *MongoDBHelper) Query(collectionName string, key string, value string, data interface{}) error {

	collection := mdb.db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	result := collection.FindOne(ctx, bson.M{key: value})
	err := result.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func (mdb *MongoDBHelper) QueryByUid(collectionName string, key string, value primitive.ObjectID, data interface{}) error {

	collection := mdb.db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	result := collection.FindOne(ctx, bson.M{key: value})
	err := result.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

func (mdb *MongoDBHelper) Insert(collectionName string, data interface{}) error {

	collection := mdb.db.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	new_user, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, new_user)

	if err != nil {
		return err
	}

	return err
}

func (mdb *MongoDBHelper) Delete(string, interface{}) error {

	return nil

}

func (mdb *MongoDBHelper) Increment(collectionName string, filter_key string, filter_value string, field string, scale int) error {

	collection := mdb.db.Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{filter_key, filter_value}}
	update := bson.D{{"$inc", bson.D{{field, scale}}}}

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}
