package tests

import (
	"fmt"
	"testing"

	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
)

type HasManyUser struct {
	*goeloquent.EloquentModel
	Id          int64              `json:"id" goelo:"column:id;primaryKey;"`
	Name        string             `json:"name" goelo:"column:name;"`
	Account     string             `json:"account" goelo:"column:account;"`
	Addresses   []*HasManyAddress  `json:"addresses" goelo:"HasMany:AddressesRelation;"`
	AddressesP  *[]HasManyAddress  `json:"addressesP" goelo:"HasMany:AddressesRelation;"`
	AddressesS  []HasManyAddress   `json:"addressesS" goelo:"HasMany:AddressesRelation;"`
	AddressesPP *[]*HasManyAddress `json:"addressesPP" goelo:"HasMany:AddressesRelation;"`
}

func (h *HasManyUser) GetTableName(st *goeloquent.Statement) string {
	return "models"
}
func (h *HasManyUser) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}
func (h *HasManyUser) AddressesRelation() *goeloquent.HasManyRelation {
	return h.HasMany(h, &HasManyAddress{}, "id", "user_id")
}

type HasManyAddress struct {
	*goeloquent.EloquentModel
	Id      int64  `json:"id" goelo:"column:id;primaryKey;"`
	UserId  int64  `json:"userId" goelo:"column:user_id;"`
	Sort    int    `json:"sort" goelo:"column:sort;"`
	Country string `json:"country" goelo:"column:country;"`
	Address string `json:"address" goelo:"column:address;"`
}

func (a *HasManyAddress) GetTableName(st *goeloquent.Statement) string {
	return "addresses"
}
func (a *HasManyAddress) GetConnectionName(st *goeloquent.Statement) string {
	return "test"
}

func TestCreateMethodProperlyCreatesNewModelHasMany(t *testing.T) {
	//todo

}
func TestFirstOrNewMethodWithValuesFindsFirstModel(t *testing.T) {
	//todo

}
func TestFirstOrCreateMethodWithValuesFindsFirstModel(t *testing.T) {
	//todo

}
func TestUpdateOrCreateMethodFindsFirstModelAndUpdates(t *testing.T) {
	//todo

}
func TestFirstOrNewMethodFindsFirstModel(t *testing.T) {
	//todo
}
func TestEagerConstraintsAreProperlyAddedHasMany(t *testing.T) {
	after := "drop table if exists models; drop table if exists addresses;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), account varchar(255)); " +
		"create table addresses (id int auto_increment primary key , user_id int,sort int, country varchar(255), address varchar(255))"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user1 := &HasManyUser{
			Name:    "John Doe",
			Account: "test",
		}
		user2 := &HasManyUser{
			Name:    "Jane Doe1",
			Account: "gmail",
		}
		st, err := user1.Init(user1).Save()
		user2.Init(user2).Save()
		assert.Nil(t, err)
		assert.Equal(t, "insert into `models` (`account`, `name`) values (?, ?)", st.RawSql)
		assert.Equal(t, user1.Id, int64(1))

		var sts []*goeloquent.Statement
		goeloquent.DB.Listen(goeloquent.EventQueryExecuted, func(name goeloquent.EventName, i ...interface{}) bool {
			sts = append(sts, i[0].(*goeloquent.Statement))
			return true
		})

		var users []HasManyUser
		_, err = conn.Model(&users).With("Addresses").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, 2, len(sts))
		assert.Equal(t, "select * from `addresses` where `addresses`.`user_id` in (?, ?)", sts[1].RawSql)
		assert.Equal(t, []interface{}{int64(1), int64(2)}, sts[1].GetBindings())
	})
}
func TestModelsAreProperlyMatchedToParentsHasMany(t *testing.T) {
	after := "drop table if exists models; drop table if exists addresses;"
	before := after + "create table models (id int auto_increment  primary key, name varchar(255), account varchar(255)); " +
		"create table addresses (id int auto_increment primary key , user_id int ,sort int, country varchar(255), address varchar(255))"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		var us []HasManyUser
		var address []HasManyAddress
		for i := 1; i <= 10; i++ {
			us = append(us, HasManyUser{
				Name:    fmt.Sprintf("User %d", i),
				Account: fmt.Sprintf("account%d", i),
			})
			for j := 1; j <= i; j++ {
				address = append(address, HasManyAddress{
					UserId:  int64(i),
					Sort:    j,
					Country: fmt.Sprintf("Country %d", j),
					Address: fmt.Sprintf("Address %d", j),
				})
			}

		}
		_, err := conn.Model(&us).Insert(&us)
		assert.Nil(t, err)
		_, err = conn.Model(&address).Insert(&address)
		assert.Nil(t, err)
		var users []HasManyUser
		_, err = conn.Model(&users).With([]string{"Addresses", "AddressesS", "AddressesS", "AddressesPP", "AddressesP"}).Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(users))
		for _, user := range users {
			assert.Equal(t, fmt.Sprintf("User %d", user.Id), user.Name)
			assert.Equal(t, fmt.Sprintf("account%d", user.Id), user.Account)
			assert.Equal(t, user.Id, int64(len(user.Addresses)))
			assert.Equal(t, user.Id, int64(len(*user.AddressesP)))
			assert.Equal(t, user.Id, int64(len(user.AddressesS)))
			assert.Equal(t, user.Id, int64(len(*user.AddressesPP)))

			for _, manyAddress := range user.Addresses {
				assert.Equal(t, fmt.Sprintf("Country %d", manyAddress.Sort), manyAddress.Country)
				assert.Equal(t, fmt.Sprintf("Address %d", manyAddress.Sort), manyAddress.Address)
			}
		}

	})
}
