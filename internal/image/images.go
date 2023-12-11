package image

import (
	"context"

	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Image data type
type Image struct {
	ID   string `bson:"_id,omitempty"`
	Data []byte
}

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("images")
}

// Add Adds an image to image store
func Add(image []byte) (imageID string, err error) {

	i := Image{primitive.NewObjectID().Hex(), image}
	_, err = Collection().InsertOne(context.TODO(), &i)
	if err != nil {
		return "", err
	}

	return i.ID, err
}

// Get Fetches the image by ID
func Get(imageId string) (Image, error) {
	var image Image

	err := Collection().FindOne(context.TODO(), bson.M{"_id": imageId}).Decode(&image)
	return image, err
}
