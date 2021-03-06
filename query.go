package mgou

import (
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

type Query struct {
	Q      interface{} `json:"-"`
	Count  int         `json:"total"`
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
	Result interface{} `json:"result"`
}

func NewQuery() *Query {
	var q Query
	Q := bson.M{}
	q.Q = Q
	return &q
}

func (q *Query) Like(m Modeler) {
	q.Q = reflect.ValueOf(m).Interface()
}
