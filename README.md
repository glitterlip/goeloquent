# Get Started

## Install

```
go get github.com/glitterlip/go-eloquent
```

## Configuration

```
config := map[string]goeloquent.DBConfig{
    "default": {//connection name
        Host:     "172.17.0.1",
        Port:     "3506",
        Database: "eloquent",
        Username: "root",
        Password: "root",
        Driver:   "mysql",
    },
    "chat":{
        Host:     "172.17.0.2",
        Port:     "3506",
        Database: "chat",
        Username: "root",
        Password: "root",
        Driver:   "mysql",
    },
}

goeloquent.Open(config)
//set logger
goeloquent.Eloquent.SetLogger(func(log goeloquent.Log) {
    fmt.Println(log)
})
```

## Running SQL Queries

### Select

```
var row = make(map[string]interface{})
var rows []map[string]interface{}
result,err := goeloquent.Eloquent.Select("select * from users where id = ? ", []interface{}{1}, &row)    
result,err = goeloquent.Eloquent.Select("select * from users limit 10", nil, &rows)    
```

first arguement is sql,second is parameter bindings,third is the destination (should be a pointer),the function return two results, first is Golang
Standard libray `database/sql`,`sql.Result`,sql error will bind to second result

### Insert

```
result,err := goeloquent.Eloquent.Insert("insert into posts (user_id,title,summary,content,created_at) values (?,?,?,?,?) ", []interface{}{2,"title","nope","test insert function",time.Now()})
if err != nil {
    panic(err.Error())
}
fmt.Println(result.LastInsertId())
```

you can get id by call `result.LastInsertId()`

### Update

```
result,err := goeloquent.Eloquent.Update("update posts set published_at = ? where id = ?", []interface{}{time.Now(),2})
if err != nil {
    panic(err.Error())
}
fmt.Println(result.RowsAffected())
```

you can get affected rows count by call `result.RowsAffected()`

### Delete

```
result,err := goeloquent.Eloquent.Delete("delete from users where age > ?", []interface{}{150})
if err != nil {
    panic(err.Error())
}
fmt.Println(result.RowsAffected())
```

### Change Connections

you may have more than one database connections, in this case you can use `Conn` to chose which connection to run sql.    
If you didn't specify which connection, we will use the connection with name `default`

```
goeloquent.Eloquent.Conn("chat").Select("select * from users order by id desc limit 1 ", nil, &user)
goeloquent.Eloquent.Conn("default").Select("select * from users limit 1 ", nil, &user1)
goeloquent.Eloquent.Conn("chat").Select("select * from users where id = ? ", []interface{}{1}, &user2)
goeloquent.Eloquent.Conn("default").Select("select * from users where phone is null  limit 3 ", nil, &userSlice)
```

# QueryBuilder

You may realise that it's not eloquent at all !!! You still need to write sql manually and process bindings. While query builder provides a more
convenient way to create sqls.

## Retrieving All Rows From A Table

```
type ChatUser struct {
	Table    string `goelo:"TableName:users"`
	Id       int64
	UserName string
	Email    sql.NullString
	Location sql.NullString
}

var users []map[string]interface{}
var userStructs []ChatUser
goeloquent.Eloquent.Conn("chat").Table("users").Get(&users,"id","user_name")
//select `id`,`user_name` from `users` [] {23} 66.375062ms
goeloquent.Eloquent.Conn("chat").Table("users").Get(&userStructs)
//select * from `users` [] {23} 69.783547ms
goeloquent.Eloquent.Conn("chat").Model(&ChatUser{}).Get(&userModelStructs)
//select * from `users` [] {23} 71.613366ms
```

As we said before, you can use `Conn` to decide which connection we use to run sql.And if you don't we use the `default` connection.   
`Table` function specify table name.    
`Get` method first paramerter is destination, it takes a second variadic parameter as columns to select

## Use Struct as destination

You can pass by a pointer of struct that has a table field with a `goelo:"TableName:realname` tag to specify table name,otherwise we will use
struct'snakename(e.g. User=>user,ChatUsers=>chat_users). We will bind struct filed snake name to table column name.(filed name 'UserName' => table
columnname 'user_name')
You can change this convension by add an additional tag `column:real_column_name`

```
//An example of struct
type ModelUser struct {
	goeloquent.EloquentModel
	Id        int64     `goelo:"column:id;primaryKey"`
	UserName  string    `goelo:"column:username"`
	Age       int       `goelo:"column:age"`
	Balance   int       `goelo:"column:balance"`
	Email     string    `goelo:"column:email"`
	CreatedAt time.Time `goelo:"column:created_at;CREATED_AT"`
	UpdatedAt time.Time `goelo:"column:updated_at;UPDATED_AT"`
}
```

## Retrieving A Single Row / Column From A Table

### Retrieving A Single Row

```
var user = make(map[string]interface{})
var userStruct ChatUser

goeloquent.Eloquent.Conn("chat").Table("users").First(&user)
//select * from `users` limit 1 [] {1} 62.54864ms}

goeloquent.Eloquent.Conn("chat").Model(&userStruct).First(&user)
//select * from `users` limit 1 [] {1} 63.045787ms

goeloquent.Eloquent.Conn("chat").Model(&userStruct).First(&userStruct)
//select * from `users` limit 1 [] {1} 63.53492ms


goeloquent.Eloquent.Conn("chat").Model(&userStruct).Get(&user)
//select * from `users` [] {23} 83.86117ms

goeloquent.Eloquent.Conn("chat").Model(&userStruct).Get(&userStruct)
fmt.Println(userStruct) //{ 100 asd11111211a { false} { false}}
//select * from `users` [] {23} 65.592061ms

goeloquent.Eloquent.Conn("chat").Model(&userStruct).Limit(1).Get(&userStruct1)
fmt.Println(userStruct1) //{ 1 bot {asda true} {asdas true}}
//select * from `users` limit 1 [] {1} 62.369403ms

```

If you just want retrieve a single row,use `First` method, it will add a `limit 1` constraint to the sql . First parameter is destination

> **WARNING**
> As you can see , if you pass a map or struct pointer , it returns a single row, but that's diffierent.  
> without `Limit(1)` , `Get` will retrive all records and iterate and assign value to destination every time ,finally assign id 100 to the destination  
> so if you use get to retriver a single row , don't forget add `Limit(1)`

### Retrieving A Single Column

```
var phone string
goeloquent.Eloquent.Conn("chat").Model(&userStruct).Where("location", "pv").Value(&phone, "phone")
//select `phone` from `users` where `location` = ? [pv] {1} 73.93333ms
fmt.Printf("%#v\n",phone) //"19875920025"

```

If you don't need entire row,just a single column value ,use `Value` method

### Retrieving A Single Row By Id

This method requires an addtional `goelo:"primaryKey"` tag. Like Laravel,first parameter is a struct pointer and second one is single value,or first
parameter is pointer of a slice of struct,and second one is slice

```
type ChatPkUser struct {
	Table    string `goelo:"TableName:users"`
	Id       int64  `goelo:"primaryKey"`
	UserName string
	Email    sql.NullString
	Location sql.NullString
}
var userStruct ChatPkUser
var userStructSlice []ChatPkUser

goeloquent.Eloquent.Conn("chat").Model(&userStruct).Find(&userStruct, 100)
fmt.Println(userStruct)
//select * from `users` where `id` in (?) limit 1 [100] {1} 72.993402ms
//{ 100 asd11111211a { false} { false}}

goeloquent.Eloquent.Conn("chat").Model(&userStruct).Find(&userStructSlice,[]interface{}{100, 98, 96,106})
fmt.Println(userStructSlice)
//select * from `users` where `id` in (?,?,?,?) [100 98 96 106] {3} 63.02497ms
//[{ 96 Allen { false} { false}} { 98 Hon { false} { false}} { 100 Frank { false} { false}}]

```

### Retrieving A List Of Columns Values

Sometimes you just want the name of the users,use `Pluck`,it takes a pointer of slice and a column name

```
goeloquent.Eloquent.Conn("chat").Model(&userStruct).Pluck(&names, "user_name")
fmt.Println(names)
//{select `user_name` from `users` [] {23} 62.63336ms
//[]string{"a", "aa", "asd", "asd1", "asd11", "b", "bot", "c", "d", "f", "s", "x", "z"}
```

## Chunking Results

### Chunk

developing

### ChunkById

developing

## Aggregates

```
var total float64
var avg float64
var max float64
var sum float64
var min float64
goeloquent.Eloquent.Conn("default").Model(&DefaultUser{}).Where("age",">",20).Count(&total, "balance")
//{select count(`balance`) as aggregate from `users` where `age` > ? [20] {1} 69.96036ms}
goeloquent.Eloquent.Conn("default").Model(&DefaultUser{}).Max(&max, "balance")
//{select max(`balance`) as aggregate from `users` [] {1} 72.807184ms}
goeloquent.Eloquent.Conn("default").Model(&DefaultUser{}).Min(&min, "balance")
//{select min(`balance`) as aggregate from `users` [] {1} 64.839132ms}
goeloquent.Eloquent.Conn("default").Model(&DefaultUser{}).Sum(&sum, "balance")
//{select sum(`balance`) as aggregate from `users` [] {1} 78.220135ms}
goeloquent.Eloquent.Conn("default").Model(&DefaultUser{}).Avg(&avg, "balance")
//{select avg(`balance`) as aggregate from `users` [] {1} 94.77536ms}
fmt.Println(total)//10
fmt.Println(min)//0
fmt.Println(sum)//1600
fmt.Println(max)//100
fmt.Println(avg)//53.3333

```

## Select Statements

### Specifying Columns

```
goeloquent.Eloquent.Conn("chat").Model(&ChatPkUser{}).Select("id","phone","location").Get(&userStructSlice)
//select `id`,`phone`,`location` from `users` [] {23} 66.385671ms
goeloquent.Eloquent.Conn("chat").Model(&ChatPkUser{}).Get(&userStructSlice,"id","phone","location")
//select `id`,`phone`,`location` from `users` [] {23} 70.993282ms
```

You can use `Sselect` or `Get` second parameter to specify which columns to select

## Raw Expressions

developing

## Joins

## Basic Where Clauses

Usually `Where` function takes 4 parameters,it's column,operator,value,`and/or` Logical Operators.    
Default Logical Operators is `and`,default operator is `=`.

```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("age",">",18,goeloquent.BOOLEAN_AND).Where("balance","=",0,goeloquent.BOOLEAN_AND).Get(&userStructSlice1)
//select * from `users` where `age` > ? and `balance` = ? [18 0] {2} 62.183917ms
fmt.Printf("%#v",userStructSlice1)
//[]main.DefaultUser{main.DefaultUser{Table:"", Id:2, Age:20, Balance:0}, main.DefaultUser{Table:"", Id:26, Age:123, Balance:0}}

```
You can skip 4th parameter when it's `and`
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("age",">",18).Where("balance","=",0).Get(&userStructSlice1)
//select * from `users` where `age` > ? and `balance` = ? [18 0] {2} 62.012133ms
```
You can skip 2nd parameter when it's `=`
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("age",">",18).Where("balance",0).Get(&userStructSlice1)
//select * from `users` where `age` > ? and `balance` = ? [18 0] {2} 62.202419ms
```
You can pass by a `[][]interface{}`,each element should be a `[]interface` containing the four parameters that pass to the `where` function  
inspired by   
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where([][]interface{}{
    {"age", ">", 18, goeloquent.BOOLEAN_AND},
    {"balance", "=", 0, goeloquent.BOOLEAN_AND},
}).Get(&userStructSlice1)
//select * from `users` where `age` > ? and `balance` = ? [18 0] {2} 62.523877ms
```
skip parameters works too
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where([][]interface{}{
    {"age", ">", 18},
    {"balance", 0},
}).Get(&userStructSlice1)
//select * from `users` where `age` > ? and `balance` = ? [18 0] {2} 61.789099ms
```

## Or Where Clauses
For more readable reason,you may want a `OrWhere` function
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("age",">",18).OrWhere("balance","=",0).Get(&userStructSlice1)
select * from `users` where `age` > ? or `balance` = ? [18 0] {24} 62.61687ms
```
## Additional Where Clauses
### WhereBetween/OrWhereBetween/WhereNotBetween/OrWhereNotBetween
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("balance", ">", 100).WhereBetween("age", []interface{}{18, 35}).Get(&userStructSlice)
//select * from `users` where `balance` > ? and `age` between ? and ? [100 18 35] {0} 68.290583ms

goeloquent.Eloquent.Model(&DefaultUser{}).WhereNotBetween("age", []interface{}{18, 35}).Get(&userStructSlice)
//select * from `users` where `age` not between ? and ? [18 35] {23} 69.032302ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("balance", ">", 100).OrWhereNotBetween("age", []interface{}{18, 35}).Get(&userStructSlice)
//select * from `users` where `balance` > ? or `age` not between ? and ? [100 18 35] {23} 62.927148ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("balance", ">", 100).OrWhereBetween("age", []interface{}{18, 35}).Get(&userStructSlice)
//select * from `users` where `balance` > ? or `age` between ? and ? [100 18 35] {7} 63.241122ms
```
### WhereIn/OrWhereIn/WhereNotIn/OrWhereNotIn
```
goeloquent.Eloquent.Model(&DefaultUser{}).WhereIn("id", []interface{}{1,2,3}).Get(&userStructSlice)
//select * from `users` where `id` in (?,?,?) [1 2 3] {1} 62.159353ms

goeloquent.Eloquent.Model(&DefaultUser{}).WhereNotIn("id", []interface{}{2,3,4}).Get(&userStructSlice)
//select * from `users` where `id` not in (?,?,?) [2 3 4] {28} 68.078067ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("username","john").OrWhereIn("email", []interface{}{"john@gmail.com","john@hotmail.com","john@apple.com","john@outlook.com"}).Get(&userStructSlice)
//select * from `users` where `username` = ? or `email` in (?,?,?,?) [john john@gmail.com john@hotmail.com john@apple.com john@outlook.com] {3} 61.692218ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("username","joe").OrWhereNotIn("email", []interface{}{"joe@gmail.com","joe@hotmail.com","joe@apple.com","joe@outlook.com"}).Get(&userStructSlice)
//select * from `users` where `username` = ? or `email` not in (?,?,?,?) [joe joe@gmail.com joe@hotmail.com joe@apple.com joe@outlook.com] {30} 64.416506ms
```
### WhereNull/OrWhereNull/OrWhereNotNull/WhereNotNull
```
goeloquent.Eloquent.Model(&DefaultUser{}).WhereIn("id", []interface{}{1,2,3}).WhereNull("email").Get(&userStructSlice)
//select * from `users` where `id` in (?,?,?) and `email` is null  [1 2 3] {0} 61.984595ms

goeloquent.Eloquent.Model(&DefaultUser{}).WhereNotIn("id", []interface{}{2,3,4}).WhereNotNull("email").Get(&userStructSlice)
//select * from `users` where `id` not in (?,?,?) and `email` is  not null  [2 3 4] {27} 62.228735ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("username","john").OrWhereNull("email").Get(&userStructSlice)
//select * from `users` where `username` = ? or `email` is null  [john] {1} 62.454664ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("username","joe").OrWhereNotNull("email").Get(&userStructSlice)
//select * from `users` where `username` = ? or `email` is  not null  [joe] {29} 62.256084ms
```
### WhereDate/WhereMonth/WhereDay/WhereYear/WhereTime
```
var now = time.Now()
fmt.Println(now)
//2021-11-03 16:00:35.461691 +0800 CST m=+0.166644409
goeloquent.Eloquent.Model(&DefaultUser{}).WhereDate("created_at", now).Get(&userStructSlice)
//{select * from `users` where date(`created_at`) = ? [2021-11-03] {0} 65.800819ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereDate("created_at", "2008-01-03").Get(&userStructSlice)
//{select * from `users` where date(`created_at`) = ? [2008-01-03] {0} 66.675012ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereDay("created_at", now).Get(&userStructSlice)
//{select * from `users` where day(`created_at`) = ? [03] {0} 65.159437ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereDay("created_at", "06").Get(&userStructSlice)
//{select * from `users` where day(`created_at`) = ? [06] {0} 64.92847ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereMonth("created_at", now).Get(&userStructSlice)
//{select * from `users` where month(`created_at`) = ? [11] {10} 70.454652ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereMonth("created_at", "06").Get(&userStructSlice)
//{select * from `users` where month(`created_at`) = ? [11] {10} 66.694005ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereYear("created_at", now).Get(&userStructSlice)
//{select * from `users` where year(`created_at`) = ? [2021] {11} 64.805563ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereYear("created_at", "2020").Get(&userStructSlice)
//{select * from `users` where year(`created_at`) = ? [2020] {0} 64.970053ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereTime("created_at", now).Get(&userStructSlice)
//{select * from `users` where time(`created_at`) = ? [16:00:35] {0} 65.73327ms}

goeloquent.Eloquent.Model(&DefaultUser{}).WhereTime("created_at", "3:05:16").Get(&userStructSlice)
//{select * from `users` where time(`created_at`) = ? [3:05:16] {0} 66.24917ms}
```
### WhereColumn/OrWhereColumn
```
goeloquent.Eloquent.Model(&DefaultUser{}).WhereColumn("age", "=", "balance").Get(&userStructSlice)
//{select * from `users` where `age` = `balance` [] {1} 65.095414ms}

goeloquent.Eloquent.Model(&DefaultUser{}).Where("id",4).OrWhereColumn("age", "=", "balance").Get(&userStructSlice)
//{select * from `users` where `id` = ? or `age` = `balance` [4] {2} 66.101059ms}
```
## Logical Grouping
If you need to group an `where` condition within parentheses,you can pass by a function to  `Where` or use `WhereNested` function
```
goeloquent.Eloquent.Model(&DefaultUser{}).Where("age", ">", 30).OrWhere(func(builder *goeloquent.Builder) {
    builder.Where("age", ">", 18)
    builder.Where("balance", ">", 5000)
}).Get(&userStructSlice, "username", "email")
//select `username`,`email` from `users` where `age` > ? or (`age` > ? and `balance` > ?) [30 18 5000] {8} 62.204423ms

goeloquent.Eloquent.Model(&DefaultUser{}).Where("age", ">", 30).WhereNested([][]interface{}{
    {"age", ">", 18},
    {"balance", ">", 5000},
},goeloquent.BOOLEAN_OR).Get(&userStructSlice, "username", "email")
//select `username`,`email` from `users` where `age` > ? or (`age` > ? and `balance` > ?) [30 18 5000] {8} 64.868523ms	
```
## Subquery Where Clauses

[https://github.com/qclaogui/database](https://github.com/qclaogui/database)      
[https://github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)