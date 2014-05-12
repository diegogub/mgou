package mgou

import(
    "labix.org/v2/mgo"
    "errors"
)

var(
    mgos *mgo.Session // mongo session to mgou
)

func Mongo() *mgo.Session{
    return mgos.Clone()
}

func Init(s *mgo.Session) error{
    if s == nil{
        return errors.New("Invalid session to init mgou")
    }
    // Ping to server if availale
    err := s.Ping()
    if err != nil{
        return err
    }
    mgos = s
    return nil
}
