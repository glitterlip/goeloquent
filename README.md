# WORK IN PROGRESS
# Get Started
A golang ORM Framework like Laravel's Eloquent

Example

```golang
var user models.User
var users []models.User
DB.Table("users").Find(&user, 6)
DB.Table("users").Where("age", ">", 10).Limit(10).Get(&users)
p := &goeloquent.Paginator{
    Items:       &users,
    PerPage:     10,
    CurrentPage: 3,
}
DB.Model(&models.User{}).With("Info","Videos","Posts.Comments").Paginate(p,6)
```
More Details,visit [Docs](https://glitterlip.github.io/go-eloquent-docs/)
# Credits
[https://github.com/go-gorm/gorm](https://github.com/go-gorm/gorm)  
[https://github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)  
[https://github.com/qclaogui/database](https://github.com/qclaogui/database)        
