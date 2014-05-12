package mgou

import(
    "labix.org/v2/mgo/bson"
    "time"
    "reflect"
    "strconv"
    "errors"
)

const(
    DAYCOL = "stats_day"
    MONTHCOL = "stats_month"
)

type Stats struct{
    // stats in UTC
    Id                  bson.ObjectId   `json:"id"                      bson:"_id,omitempty"`
    MainReference       bson.ObjectId   `required:"-"  json:"main_ref"  bson:"m,omitempty"`
    SecondReference     bson.ObjectId   `soft:"-"      json:"sec_ref"   bson:"se,omitempty"`
    Type                string          `enum:"d,m"  json:"type"        bson:"t,omitempty"`
    Date                time.Time       `              json:"period"    bson:"d,omitempty"`
    Stats               []map[string]interface{}     `json:"stats"      bson:"s,omitempty"`
}

type StatQuery struct{
    MainReference       bson.ObjectId   `required:"-"  json:"main_ref"  bson:"m"`
    SecondReference     bson.ObjectId   `soft:"-"      json:"sec_ref"   bson:"se,omitempty"`
    Type                string          `enum:"d,m"    json:"type"      bson:"t"`
    Date                time.Time       `              json:"period"    bson:"d,omitempty"`
}
//give period in UTC

func DailyFloatStat(key []string,from time.Time,to time.Time,offset int,main bson.ObjectId,seconds []bson.ObjectId) map[string]float64{
    // from and to times are days
    var periodStat []Stats
    var fromHr,toHr int
    var fromYear,fromDay int
    var fromMonth time.Month
    var toYear,toDay int
    var toMonth time.Month

    var result map[string]float64
    result = make(map[string]float64)

    fromYear,fromMonth,fromDay = from.Date()
    toYear,toMonth,toDay       = to.Date()
    UTC ,_ := time.LoadLocation("")

    from = time.Date(fromYear,fromMonth,fromDay,0,0,0,0,UTC)
    to = time.Date(toYear,toMonth,toDay,0,0,0,0,UTC)
    setTimes(offset,&fromHr,&toHr)

    if offset < 0 {
        from.AddDate(0,0,-1)
    }else{
        to.AddDate(0,0,1)
    }
/*
    sameDay = (fromYear == toYear)
    sameDay = (fromMonth == toMonth)
    sameDay = (fromDay   == toDay)
*/
    m := Mongo()
    m.DB("").C(DAYCOL).Find( bson.M{ "d" : bson.M{ "$gte" : from , "$lte" : to} }).All(periodStat)

    for _, k:= range(key){
        result[k] = 0
    }

    l := len(periodStat)

    if l == 0{ //not stats for period
        return result
    }else{
        //if not offset and 1 stat returned
        if offset == 0 && l == 1{
            for _,k := range(key){
                result[k] += periodStat[0].floatStat(k,fromHr,23)
            }
            return result
        }
        // not enough days tracked
        if offset < 0 && l == 1{
            if from.Equal(periodStat[0].Date){
                for _,k := range(key){
                    result[k] += periodStat[0].floatStat(k,fromHr,23)
                }
            }else{
                if periodStat[0].Date.Equal(from.AddDate(0,0,1)){
                    for _,k := range(key){
                        result[k] += periodStat[0].floatStat(k,0,toHr)
                    }
                }else{
                    for _,k := range(key){
                        result[k] += periodStat[0].floatStat(k,0,23)
                    }
                }
            }
        }

        if offset > 0 && l == 1{
            if from.Equal(periodStat[0].Date){
                for _,k := range(key){
                    result[k] += periodStat[0].floatStat(k,fromHr,23)
                }
            }else{
                for _,k := range(key){
                    result[k] += periodStat[0].floatStat(k,0,toHr)
                }
            }
        }

    }

    for _, stat := range(periodStat){
        if from.Equal(stat.Date){
            for _,k := range(key){
                result[k] += stat.floatStat(k,fromHr,23)
            }
        }
        if to.Equal(stat.Date){
            for _,k := range(key){
                result[k] += stat.floatStat(k,0,toHr)
            }
        }else{
            for _,k := range(key){
                result[k] += stat.floatStat(k,0,23)
            }
        }
    }

    return result
}

func DailyIntStat(key []string,from time.Time,to time.Time,offset int,main bson.ObjectId,seconds []bson.ObjectId) map[string]int{
    // from and to times are days
    var periodStat []Stats
    var fromHr,toHr int
    var fromYear,fromDay int
    var fromMonth time.Month
    var toYear,toDay int
    var toMonth time.Month

    var result map[string]int
    result = make(map[string]int)

    fromYear,fromMonth,fromDay = from.Date()
    toYear,toMonth,toDay       = to.Date()
    UTC ,_ := time.LoadLocation("")

    from = time.Date(fromYear,fromMonth,fromDay,0,0,0,0,UTC)
    to = time.Date(toYear,toMonth,toDay,0,0,0,0,UTC)
    setTimes(offset,&fromHr,&toHr)

    if offset < 0 {
        from.AddDate(0,0,-1)
    }else{
        to.AddDate(0,0,1)
    }

    m := Mongo()
    m.DB("").C(DAYCOL).Find( bson.M{ "d" : bson.M{ "$gte" : from , "$lte" : to} }).All(periodStat)

    for _, k:= range(key){
        result[k] = 0
    }

    l := len(periodStat)

    if l == 0{
        return result
    }

    for i, stat := range(periodStat){
        if i == 0{
            for _,k := range(key){
                result[k] += stat.intStat(k,fromHr,23)
            }
        }
        if i == l-1{
            for _,k := range(key){
                result[k] += stat.intStat(k,0,toHr)
            }
        }else{
            for _,k := range(key){
                result[k] += stat.intStat(k,0,23)
            }
        }
    }

    return result
}

func (s Stats) floatStat(key string,from, to int) float64{
    var i int
    var ok bool
    var sum float64
    for i = from; i <= to ; i++{
        _, ok = s.Stats[i][key]
        if ok {
            sum += s.Stats[i][key].(float64)
        }
    }
    return sum
}

func (s Stats) intStat(key string,from, to int) int{
    var i int
    var ok bool
    var sum int
    for i = from; i <= to ; i++{
        _, ok = s.Stats[i][key]
        if ok {
            sum += s.Stats[i][key].(int)
        }
    }
    return sum
}

func setTimes(offset int ,start, end  *int){
    if offset > 0{
        *start += offset
    }else{
        if offset < 0 {
            *start = 24 + offset
            *end = *start - 1
        }else{
            *start = 0
            //Would be like *end= 24 - 1
            *end= 23
        }
    }
}

func GetStats(t string,year int,month int,day int,ids...bson.ObjectId) []Stats{
    var stats []Stats
    return stats
}

func NewStat(main bson.ObjectId,sec bson.ObjectId,t string,year int,month int,day int) *Stats{
    var s Stats
    if main == ""{
        return nil
    }
    if t == "m" || t == "d"{
        s.Type =  t
    }else{
        return nil
    }
    switch s.Type{
        case "m":
            s.Stats = make([]map[string]interface{},12)
        case "d":
            s.Stats = make([]map[string]interface{},24)
    }
    for i,_:=range(s.Stats){
        s.Stats[i] = make(map[string]interface{})
    }
    // set date
    UTC ,_ := time.LoadLocation("")
    m := time.Month(month)
    date := time.Date(year,m,day,0,0,0,0,UTC)
    //--------------------------------------//
    s.MainReference = main
    s.SecondReference = sec
    s.Date = date
    return &s
}

func (s Stats) Collection() string{
    switch s.Type{
        case "m":
            return MONTHCOL
        case "d":
            return DAYCOL
        default:
            return ""
    }
}

func(s *Stats) Save() (bson.ObjectId,error){
    m := Mongo()
    defer m.Close()
    if s.MainReference == "" {
        return "",errors.New("Must have main referece to save stat")
    }
    if s.Type == "" {
        return "",errors.New("Must have type")
    }
    if s.Id != ""{
        // Id must be "" to save it 
        return "",errors.New("Already Exist")
    }else{
        id :=bson.NewObjectId()
        s.Id = id
        err := m.DB("").C(s.Collection()).Insert(s)
        if err != nil{
            return "",err
        }
        return id,nil
    }
}

// Increments counters 
// per references
func (s *Stats) Inc(main string,sec string,t string,date time.Time,key string,amount interface{}) error{
    sum := reflect.ValueOf(amount)
    var stat StatQuery
    var num int

    //Check if it's numeric
    switch sum.Kind(){
        case reflect.Int,reflect.Float32,reflect.Float64:

        default:
            return errors.New("Invalid amount type")
    }

    switch t{
        case "m":
            stat.Type = t
            // 0 - 11 month
            num = int(date.Month()) - 1
        case "d":
            stat.Type = t
            num = date.Hour()
        default:
            return errors.New("Invalid Type")
    }
    stat.MainReference = bson.ObjectIdHex(main)

    if sec != ""{
        stat.SecondReference = bson.ObjectIdHex(sec)
    }
 //   stat.Date = date
    mongo := Mongo()
    defer mongo.Close()

    // Update amount
    err := mongo.DB("").C(s.Collection()).Update(stat, bson.M{ "$inc" : bson.M{ "s." + strconv.Itoa(num) + "." + key : amount }})
    return err
}
