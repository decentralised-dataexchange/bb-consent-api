package image

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Image data type
type Image struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Data []byte
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("images")
}

// Add Adds an image to image store
func Add(image []byte) (imageID string, err error) {

	i := Image{primitive.NewObjectID(), image}
	_, err = collection().InsertOne(context.TODO(), &i)
	if err != nil {
		return "", err
	}

	return i.ID.Hex(), err
}

// Get Fetches the image by ID
func Get(imageID string) (Image, error) {
	var image Image

	imageId, err := primitive.ObjectIDFromHex(imageID)
	if err != nil {
		return image, err
	}

	err = collection().FindOne(context.TODO(), bson.M{"_id": imageId}).Decode(&image)
	return image, err
}
