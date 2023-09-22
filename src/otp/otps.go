package otp

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Otp Otp holds the generated OTP info
type Otp struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string
	Email    string
	Phone    string
	Otp      string
	Verified bool
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("otps")
}

// Add Adds the otp to the db
func Add(otp Otp) (Otp, error) {

	otp.ID = primitive.NewObjectID()

	_, err := collection().InsertOne(context.TODO(), otp)
	if err != nil {
		return Otp{}, err
	}

	return otp, nil
}

// Delete Deletes the otp entry by ID
func Delete(otpID string) error {
	otpId, err := primitive.ObjectIDFromHex(otpID)
	if err != nil {
		return err
	}

	_, err = collection().DeleteOne(context.TODO(), bson.M{"_id": otpId})
	if err != nil {
		return err
	}

	return nil
}

// UpdateVerified Updates the verified filed
func UpdateVerified(o Otp) error {
	filter := bson.M{"_id": o.ID}
	update := bson.M{"$set": bson.M{"verified": o.Verified}}

	_, err := collection().UpdateOne(context.TODO(), filter, update)

	return err
}

// PhoneNumberExist Check if phone number is already in the colleciton
func PhoneNumberExist(phone string) (o Otp, err error) {
	filter := bson.M{"phone": phone}

	err = collection().FindOne(context.TODO(), filter).Decode(&o)
	if err == mongo.ErrNoDocuments {
		return o, err
	} else if err != nil {
		return o, err
	}

	return o, err
}

// SearchPhone Search phone number in otp db
func SearchPhone(phone string) (Otp, error) {
	filter := bson.M{"phone": phone}

	var result Otp
	err := collection().FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}
