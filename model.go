package mgou

import(
    "reflect"
    "labix.org/v2/mgo/bson"
    "labix.org/v2/mgo"
    "strings"
    "time"
)

// Model interface
type Modeler interface{
    Clean()         Error
    Valid(err Error)         //custom Validation
    Collection()    string
    Document()      Document // returns
}

type SubModeler interface{
    Valid()
}

//Search
func Search(m Modeler,q *Query) (*mgo.Query){
    // Do not close session
    s := Mongo()
    count, _:= s.DB("").C(m.Collection()).Find(q.Q).Count()
    q.Count = count
    res := s.DB("").C(m.Collection()).Find(q.Q).Limit(q.Limit).Skip(q.Limit*q.Page)
    return res
}

// Delete model
func Delete(m Modeler) Error{
    doc := m.Document()
    var err Error = make(map[string]string)
    var e   error
    // run custom Cleaning function
    err = m.Clean()

    if len(err) > 0 {
        return err
    }

    if doc.Id != "" {
        s := Mongo()
        field := reflectValue(m).FieldByName("Doc").FieldByName("Id")
        e = s.DB("").C(m.Collection()).RemoveId(field.Interface().(bson.ObjectId))
        if e != nil{
            err["error"] = "not found"
        }
    }else{
            err["error"] = "can't delete object if not loaded"
    }
    return err
}

// Load model into struct 
func Load(m Modeler) error{
    s := Mongo()
    defer s.Close()
    var err error
    field := reflectValue(m).FieldByName("Doc").FieldByName("Id")
    err = s.DB("").C(m.Collection()).FindId(field.Interface().(bson.ObjectId)).One(m)
    // recover Collection
    return err
}

/* 
Save Model into DB - If quick is true, It does not check relations into DB (avoid reading from DB)
    tags:
        - relation:"collection" // hard relation FK  not null
        - soft:"collection"     // soft relation FK  *** MUST HAVE TAG bson:",omitempty"
        - required:"-"
        - enum:"val1,val2,...,valN"
        - unique:"-"
*/

func validate(m interface{},db *mgo.Database,col string,id bson.ObjectId,quick bool,bsonstate string,err Error){
   checkRequired(m,err)
   checkEnum(m,err)
   if !quick {
      //check soft and hard relations
      checkRelations(m,db,err)
   }
   checkUnique(m,db,col,id,bsonstate,err)

   val := Tags(m,"validate")
   if len(val) > 0 {
       for fname,_ :=range(val){
            //get bson tag and add it to bson state
            btag :=BsonTag(m,fname)
            if btag == ""{
                bsonstate += fname + "."
            }else{
                bsonstate += btag + "."
            }
            field := reflectValue(m).FieldByName(fname)
            // All sub structures are not Models
            validate(field.Interface(),db,col,id,quick,bsonstate,err)
       }
   }
   return
}

// Normal Save
func Save(m interface{},db string,col string,quick bool) (Error){
   ms := Mongo()
   defer ms.Close()
   mdb :=  ms.DB(db)
   var field reflect.Value
   var e error

   err := NewError()
   // element
   field = reflectValue(m)
   // type
   t := field.Type()

   // Must be Kind of struct
   if field.Kind() != reflect.Struct{
        panic("Only structs are supported to save into Mongo")
   }

   // Must check if it implements Modeler function
   model := t.Implements(reflect.TypeOf((*Modeler)(nil)).Elem())

   var doc Document
   // Run custom validation, get Document if Modeler and cast type
   if model {
       m.(Modeler).Valid(err)
       doc = m.(Modeler).Document()
       col = m.(Modeler).Collection()
       //No errors with custom validation, start normal validation
       if len(err) == 0{
           validate(m,mdb,col,doc.Id,quick,"",err)
       }else{
           return err
       }
   }else{
       validate(m,mdb,col,"",quick,"",err)
   }

   if len(err) == 0{
        if !model{
            e = mdb.C(col).Insert(m)
        }else{
            // check to update or insert
            if doc.Id == ""{
                // set Id
                d := reflectValue(m).FieldByName("Doc").FieldByName("Id")
                t := reflectValue(m).FieldByName("Doc").FieldByName("Created")
                id := bson.NewObjectId()
                now := time.Now()
                d.Set(reflect.ValueOf(id))
                t.Set(reflect.ValueOf(now))
                e = mdb.C(col).Insert(m)
            }else{
                //t := reflectValue(m).FieldByName("Doc").FieldByName("Updated")
                //now := time.Now()
                //t.Set(reflect.ValueOf(now))
                e = mdb.C(col).UpdateId(doc.Id,m)
            }
        }
   }
   if e != nil{
       err["model"] = e.Error()
   }


   return err
}

func isExportableField(field reflect.StructField) bool {
	return field.PkgPath == ""
}

func reflectValue(obj interface{}) reflect.Value {
    var val reflect.Value

    if reflect.TypeOf(obj).Kind() == reflect.Ptr {
        val = reflect.ValueOf(obj).Elem()
    } else {
        val = reflect.ValueOf(obj)
    }

    return val
}

func BsonTag(obj interface{},name string) string{
    if reflect.TypeOf(obj).Kind() != reflect.Struct && reflect.TypeOf(obj).Kind() != reflect.Ptr{
        return ""
    }
    var tag string
	objValue := reflectValue(obj)
    objType := objValue.Type()
    f , _:= objType.FieldByName(name)
    tag = f.Tag.Get("bson")
    stag := strings.Split(tag,",")
    if len(stag) > 0{
        return stag[0]
    }else{
        return ""
    }
}

func Tag(obj interface{},name string,key string) string{
    if reflect.TypeOf(obj).Kind() != reflect.Struct && reflect.TypeOf(obj).Kind() != reflect.Ptr{
        return ""
    }
    var tag string
	objValue := reflectValue(obj)
    objType := objValue.Type()
    f , _:= objType.FieldByName(name)
    tag = f.Tag.Get(key)
    return tag
}

func Tags(obj interface{}, key string) (map[string]string) {
    if reflect.TypeOf(obj).Kind() != reflect.Struct && reflect.TypeOf(obj).Kind() != reflect.Ptr{
        return nil
    }
    var tag string

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()
	tags := make(map[string]string)

	for i := 0; i < fieldsCount; i++ {
		structField := objType.Field(i)
        tag = structField.Tag.Get(key)
        if tag != "" {
            tags[structField.Name] = tag
        }
	}

	return tags
}

func checkRequired(m interface{},err Error){
    req := Tags(m,"required")
    if len(req) > 0{
        for fname, _:= range(req){
            if !checkField(m,fname){ // if don't
                err[fname] = "required"
            }else{
                field := reflectValue(m).FieldByName(fname)
                // check if it's empty, depending on Kind
                switch field.Kind() {
                    case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
                        if field.Len() == 0 {
                            err[fname] = "required"
                        }
                    case reflect.Interface, reflect.Ptr: if field.IsNil() { err[fname] = "required"
                        }
                }
            }
        }
    }
    return
}

func checkField(m interface{},fname string) bool{
    field := reflectValue(m).FieldByName(fname)
    if field == reflect.ValueOf(nil){
        return false
    }
    return true
}

func checkEnum(m interface{},err Error){
    enumFields := Tags(m,"enum")
    if len(enumFields) > 0{
        field := reflectValue(m)
        valid := false
        for fname,enuml :=range(enumFields){
            enumValues := strings.Split(enuml,",")
            f := field.FieldByName(fname)
            valid = false
            for _ ,e := range(enumValues){
                if e == f.String(){
                    valid = true
                }
            }
            if !valid {
                err[fname] = "Invalid enum value"
            }
        }
    }
    return
}

func checkRelations(m interface{},db *mgo.Database,err Error){
    relations := Tags(m,"relation")
    if len(relations) > 0 {
        field := reflectValue(m)
        var count int
        for fname,col:=range(relations){
            if checkField(m,fname){
                count ,_ = db.C(col).FindId(field.FieldByName(fname).Interface().(bson.ObjectId)).Count()
                if count == 0 {
                    err[fname] = "invalid reference"
                }
            }else{
                err[fname] = "required"
            }
        }
    }

    //check soft relations
    if len(err) > 0{
        return
    }
    soft := Tags(m,"soft")
    if len(soft) > 0 {
        field := reflectValue(m)
        var count int
        for fname,coll:=range(soft){
            ref := field.FieldByName(fname).Interface().(bson.ObjectId)
            if ref != ""{
                count ,_ = db.C(coll).FindId(ref).Count()
                if count == 0 {
                    err[fname] = "invalid reference"
                }
           }
        }
    }

}
// checks unique values , on update or insert
func checkUnique(m interface{},db *mgo.Database,col string,id bson.ObjectId,bsonstate string,err Error){
    uniques := Tags(m,"unique")
    if len(uniques) > 0 {
        if id == "" {
            for fname,_:=range(uniques){    // not a model or a model without Id
                field:= reflectValue(m).FieldByName(fname)
                var count int
                var e  error
                if field.Kind() != reflect.String{
                    panic("invalid type to check unique: " + field.Kind().String())
                }
                // must get bson value
                tag := BsonTag(m,fname)
                index := ""
                if tag != ""{
                    index = bsonstate + tag
                }else{
                    index = bsonstate + fname
                }
                index = strings.ToLower(index)
                count, e = db.C(col).Find(bson.M{ index : field.String()}).Count()
                if e != nil{
                    panic(e)
                }
                if count > 0{
                    err[fname] = "not unique"
                }
            }
        }else{
            var auxModel interface{}
            var dbValue  string
            // retrieve model
            db.C(col).FindId(id).One(&auxModel)
            for fname,_:=range(uniques){    // not a model or a model without Id
                field:= reflectValue(m).FieldByName(fname)
                var count int
                var e  error
                if field.Kind() != reflect.String{
                    panic("invalid type to check unique: " + field.Kind().String())
                }

                tag := BsonTag(m,fname)
                index := ""
                if tag != ""{
                    index = bsonstate + tag
                    dbValue = GetDbValue(auxModel.(bson.M),index)
                }else{
                    index = bsonstate + fname
                    dbValue = GetDbValue(auxModel.(bson.M),index)
                }
                if dbValue == field.String(){
                    // Did not change
                    count = 0
                }else{
                    index = strings.ToLower(index)
                    count, e= db.C(col).Find(bson.M{ index : field.String() }).Count()
                }
                if e != nil{
                    err[fname] = "db error"
                }
                if count > 0{
                    err[fname] = "not unique"
                }
            }
        }
    }
    return
}

func GetDbValue(m bson.M,bsonstate string) string{
    var v string
    var bmap bson.M = m
    var field reflect.Value
    btags := strings.Split(bsonstate,".")
    if len(btags) > 0 {
        //needs to retrieve value from maps of maps
        for _,tag :=range(btags){
            tag = strings.ToLower(tag)
            _,ok := bmap[tag]
            if ok {
                field = reflectValue(bmap[tag])
                if field.Kind() == reflect.Map{
                    bmap = bmap[tag].(bson.M)
                }else{
                    if field.Kind() == reflect.String{
                        v = bmap[tag].(string)
                    }
                }
            }else{
                if field.Kind() == reflect.String{
                    v = bmap[tag].(string)
                }else{
                    return ""
                }
            }
        }
    }
    return v
}
