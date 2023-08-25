package orgtype

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// OrgType Type related information
type OrgType struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Type     string
	ImageID  string
	ImageURL string
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("orgTypes")
}

// Add Adds an organization
func Add(ot OrgType) (OrgType, error) {
	s := session()
	defer s.Close()

	ot.ID = bson.NewObjectId()
	return ot, collection(s).Insert(&ot)
}

// Get Gets organization type by given id
func Get(organizationTypeID string) (OrgType, error) {
	s := session()
	defer s.Close()

	var result OrgType
	err := collection(s).FindId(bson.ObjectIdHex(organizationTypeID)).One(&result)

	return result, err
}
