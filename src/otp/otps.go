package otp

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Otp Otp holds the generated OTP info
type Otp struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string
	Email    string
	Phone    string
	Otp      string
	Verified bool
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("otps")
}

// Add Adds the otp to the db
func Add(otp Otp) (Otp, error) {
	s := session()
	defer s.Close()

	otp.ID = bson.NewObjectId()

	return otp, collection(s).Insert(&otp)
}

// Delete Deletes the otp entry by ID
func Delete(otpID string) error {
	s := session()
	defer s.Close()

	return collection(s).RemoveId(bson.ObjectIdHex(otpID))
}

// UpdateVerified Updates the verified filed
func UpdateVerified(o Otp) error {
	s := session()
	defer s.Close()
	c := collection(s)

	err := c.Update(bson.M{"_id": o.ID}, bson.M{"$set": bson.M{"verified": o.Verified}})

	return err
}

// PhoneNumberExist Check if phone number is already in the colleciton
func PhoneNumberExist(phone string) (o Otp, err error) {
	s := session()
	defer s.Close()

	q := collection(s).Find(bson.M{"phone": phone}).Limit(1)

	c, err := q.Count()
	if err != nil {
		return o, err
	}

	if c == 0 {
		return o, err
	}
	q.One(&o)

	return o, err
}

// SearchPhone Search phone number in otp db
func SearchPhone(phone string) (Otp, error) {
	s := session()
	defer s.Close()

	var result Otp
	err := collection(s).Find(bson.M{"phone": phone}).One(&result)
	if err != nil {
		return result, err
	}

	return result, err
}
