package mgou

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

// Error maps
type Error map[string]string

//copy default mongodb session
func NewError() Error {
	var e Error = make(map[string]string)
	return e
}

// Add error to field
func (e Error) Add(f string, m string) {
	e[f] = m
	return
}

type Document struct {
	Id      bson.ObjectId `json:"id"       bson:"_id,omitempty"`
	Created time.Time     `json:"created"  bson:"_t,omitempty"`
	Updated time.Time     `json:"updated"  bson:"_u,omitempty"`
}

func (d *Document) Create() {
	d.Id = bson.NewObjectId()
	d.Created = time.Now()
}

func (d *Document) Update() {
	d.Updated = time.Now()
}

func (d *Document) ID() bson.ObjectId {
	return d.Id
}

func (d *Document) SetId(id string) {
	d.Id = bson.ObjectIdHex(id)
}

func (d *Document) SetCreated(t time.Time) {
	d.Created = t
	return
}

//check if and ID exist is valid and exist into a collection

func (d *Document) Exist(s *mgo.Session, col string) bool {
	var count int
	var err error
	oid := d.ID()
	if col == "" {
		col = "test"
	}
	c := col

	if !oid.Valid() {
		return false
	}

	coll := s.DB("").C(c)
	count, err = coll.FindId(oid).Count()

	if err != nil {
		return false
	}

	if count == 1 {
		return true
	} else {
		return false
	}
}
