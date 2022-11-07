# WORK IN PROGRESS
# Get Started
A golang ORM Framework like Laravel's Eloquent
```shell
go get github.com/glitterlip/goeloquent
```
Example

```golang
//define a model
type User struct {
    *goeloquent.EloquentModel
    Id        int64          `goelo:"column:id;primaryKey"`
    UserName  sql.NullString `goelo:"column:name"`
    Age       int            `goelo:"column:age"`
    Status    int            `goelo:"column:status"`
    Friends   []UserT        `goelo:"BelongsToMany:FriendsRelation"`
    Address   Address        `goelo:"HasOne:AddressRelation"`
    CreatedAt time.Time      `goelo:"column:created_at,timestatmp:create"`
    UpdatedAt sql.NullTime   `goelo:"column:updated_at,timestatmp:update"`
}

//Find/First/Get
var user User
DB.Table("users").Find(&user,1)
DB.Table("users").Where("name","john").First(&user)
DB.Table("users").Where("name","john").FirstOrCreate(&user)

//Column/Aggeregate
var age int
DB.Table("users").Where("id","=",20).Value(&age,"age")
DB.Table("users").Max(&age,"age")
var salary int
DB.Table("user").Where("age",">=",30).Avg(&salary,"salary")

//Find record to map
var m = make(map[string]interface{})
DB.Query().From("users").Find(&m,3)
var ms []map[string]interface{}
DB.Query().From("users").Get(&ms)

//Pagination
var users []User
DB.Model(&User{}).Where("id", ">", 10).Where("id", "<", 28).Paginate(&users, 10, 1)

//Chunk/ChunkById
var total int
totalP := &total
DB.Table("users").OrderBy("id").Chunk(&[]User{}, 10, func(dest interface{}) error {
    us := dest.(*[]User)
    for _, user := range *us {
        assert.Equal(t, user.UserName, sql.NullString{
            String: fmt.Sprintf("user-%d", user.Age),
            Valid:  true,
        })
        *totalP++
    }
    return nil
})

var total int
totalP := &total
DB.Table("users").ChunkById(&[]User{}, 10, func(dest interface{}) error {
    us := dest.(*[]User)
    for _, user := range *us {
        assert.Equal(t, user.UserName, sql.NullString{
            String: fmt.Sprintf("user-%d", user.Age),
            Valid:  true,
        })
        *totalP++
    }
    return nil
})
//Query clause 
DB.Where("name","john@apple.com").OrWhere("email","john@apple.com").First(&user)

DB.Where("is_admin", 1).Where(map[string]interface{}{
    "name": "Joe", "location": "LA",
}, goeloquent.BOOLEAN_OR).Where(func(builder *goeloquent.Builder){
    builder.WhereYear("created_at", "<", 2010).WhereColumn("first_name", "last_name").OrWhereNull("activited_at")
}).ToSql()
// sql:"select `name`, `age`, `email` where `is_admin` = ? or (`name` = ? and `location` = ?) and (year(`created_at`) < ? and `first_name` = `last_name` or `activited_at` is null)"

// Insert/Update/Delete
DB.Table("users").Insert(map[string]interface{}{
    "name":       "go-eloquent",
    "age":        18,
    "created_at": now,
})
DB.Table("users").Where("id", id).Update(map[string]interface{}{
    "name":       "newname",
    "age":        18,
    "updated_at": now.Add(time.Hour * 1),
})
DB.Table("users").Where("id", id).Delete()


//Relations (suppeort hasone/hasmany/belongsto/belongstomany/morphone/morphto/morphmany/morphbymany)
type Address struct {
    *goeloquent.EloquentModel
    ID      int64  `goelo:"column:id;primaryKey"`
    User    *User `goelo:"BelongsTo:UserRelation"`
    UserId  int64  `goelo:"column:user_id"`
    Country string `goelo:"column:country"`
    State   string `goelo:"column:state"`
    City    string `goelo:"column:city"`
    Detail  string `goelo:"column:detail"`
}
DB.Model(&User{}).With("Address").Where("id", ">", 10).Where("id", "<", 28).Paginate(&users, 10, 1)
user.FriendsRelation().Get(&friends)

//WithPivot WherePivot also supported
DB.Model(&User{}).With("Friends").WherePivot("status", FriendStatusNormal).WithPivot("status", "user_id", "friend_id", "additional").Find(&user, u1.Id)
var friends []User
```
More Details,visit [Docs](https://glitterlip.github.io/go-eloquent-docs/)
# Credits
[https://github.com/go-gorm/gorm](https://github.com/go-gorm/gorm)  
[https://github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)  
[https://github.com/qclaogui/database](https://github.com/qclaogui/database)        
