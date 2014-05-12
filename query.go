package mgou

import(
    "reflect"
)

type Query struct {
    Q       interface{}   `json:"-"`
    Count   int           `json:"total"`
    Page    int           `json:"page"`
    Limit   int           `json:"limit"`
    Result  interface{}   `json:"result"`
}

func (q *Query) Like(m Modeler){
   q.Q = reflect.ValueOf(m).Interface()
}
