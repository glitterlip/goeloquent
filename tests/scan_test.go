package tests

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestScanValue(t *testing.T) {
	after := "drop table if exists scanvalue;"
	before := `
drop table if exists scanvalue;
create table scanvalue (
id int auto_increment primary key, 
int_col int,
float_col float(8,2),
double_col double,
double_col1 double(8,3),
decimal_col decimal(12,4), 
char_col char(10),
varchar_col varchar(255),
text_col text,
binary_col binary(16), 
blob_col blob, 
enum_col ENUM('option1','option2','option3'),
set_col SET('A', 'B', 'C', 'D'),
date_col date,
time_col time,
datetime_col datetime,
timestamp_col timestamp,
year_col year
)
`
	RunWithDB(before, after, func(conn goeloquent.Connection) {

		b := conn.Query()
		_, err := b.Table("scanvalue").Insert(map[string]interface{}{
			"int_col":       123,
			"float_col":     123.45,
			"double_col":    123456.789,
			"double_col1":   12345.678,
			"decimal_col":   12345.6789,
			"char_col":      "test",
			"varchar_col":   "test varchar",
			"text_col":      "test text",
			"binary_col":    []byte("binary data"),
			"blob_col":      []byte("blob data"),
			"enum_col":      "option1",
			"set_col":       "A,C",
			"date_col":      "2023-10-01",
			"time_col":      "12:34:56",
			"datetime_col":  "2023-10-01 12:34:56",
			"timestamp_col": "2023-10-01 12:34:56",
			"year_col":      2023,
		})

		assert.Nil(t, err)

		var intV int
		var floatV float64
		var doubleV float64
		var decimalV float64
		var charV string
		var varcharV string
		var textV string
		binaryV := make([]byte, 0)
		blobV := make([]byte, 0)
		var enumV string
		var setV string
		var dateV string
		var timeV string
		var datetimeV time.Time
		var timestampV time.Time
		var yearV int
		_, err = conn.Query().Table("scanvalue").Select("int_col").First(&intV)
		assert.Nil(t, err)
		assert.Equal(t, 123, intV)
		_, err = conn.Query().Table("scanvalue").Select("float_col").First(&floatV)
		assert.Nil(t, err)
		assert.Equal(t, 123.45, floatV)
		_, err = conn.Query().Table("scanvalue").Select("double_col").First(&doubleV)
		assert.Nil(t, err)
		assert.Equal(t, 123456.789, doubleV)
		_, err = conn.Query().Table("scanvalue").Select("double_col1").First(&doubleV)
		assert.Nil(t, err)
		assert.Equal(t, 12345.678, doubleV)
		_, err = conn.Query().Table("scanvalue").Select("decimal_col").First(&decimalV)
		assert.Nil(t, err)
		assert.Equal(t, 12345.6789, decimalV)
		_, err = conn.Query().Table("scanvalue").Select("char_col").First(&charV)
		assert.Nil(t, err)
		assert.Equal(t, "test", charV)
		_, err = conn.Query().Table("scanvalue").Select("varchar_col").First(&varcharV)
		assert.Nil(t, err)
		assert.Equal(t, "test varchar", varcharV)
		_, err = conn.Query().Table("scanvalue").Select("text_col").First(&textV)
		assert.Nil(t, err)
		assert.Equal(t, "test text", textV)
		_, err = conn.Query().Table("scanvalue").Select("binary_col").First(&binaryV)
		assert.Nil(t, err)
		assert.Equal(t, []byte("binary data\x00\x00\x00\x00\x00"), binaryV)
		_, err = conn.Query().Table("scanvalue").Select("blob_col").First(&blobV)
		assert.Nil(t, err)
		assert.Equal(t, []byte("blob data"), blobV)
		_, err = conn.Query().Table("scanvalue").Select("enum_col").First(&enumV)
		assert.Nil(t, err)
		assert.Equal(t, "option1", enumV)
		_, err = conn.Query().Table("scanvalue").Select("set_col").First(&setV)
		assert.Nil(t, err)
		assert.Equal(t, "A,C", setV)
		_, err = conn.Query().Table("scanvalue").Select("date_col").First(&dateV)
		assert.Nil(t, err)
		assert.Equal(t, "2023-10-01T00:00:00Z", dateV)
		_, err = conn.Query().Table("scanvalue").Select("time_col").First(&timeV)
		assert.Nil(t, err)
		assert.Equal(t, "12:34:56", timeV)
		_, err = conn.Query().Table("scanvalue").Select("datetime_col").First(&datetimeV)
		assert.Nil(t, err)
		assert.Equal(t, "2023-10-01 12:34:56", datetimeV.Format("2006-01-02 15:04:05"))
		_, err = conn.Query().Table("scanvalue").Select("timestamp_col").First(&timestampV)
		assert.Nil(t, err)
		assert.Equal(t, "2023-10-01 12:34:56", timestampV.Format("2006-01-02 15:04:05"))
		_, err = conn.Query().Table("scanvalue").Select("year_col").First(&yearV)
		assert.Nil(t, err)
		assert.Equal(t, 2023, yearV)

	})
}

func TestScanMap(t *testing.T) {
	after := "drop table if exists scanmap;"
	before := `drop table if exists scanmap;create table scanmap (id int auto_increment primary key, name varchar(255), age int, created_at datetime, updated_at datetime);`

	RunWithDB(before, after, func(conn goeloquent.Connection) {

		b := conn.Query()
		r, err := b.Table("scanmap").Insert(map[string]interface{}{
			"name":       "John Doe",
			"age":        30,
			"created_at": "2023-10-01 12:00:00",
			"updated_at": "2023-12-01 12:00:00",
		})

		assert.Nil(t, err)

		var result map[string]interface{}
		_, err = conn.Query().Table("scanmap").Select("name", "age", "created_at", "updated_at").First(&result)
		assert.Nil(t, err)
		assert.Equal(t, r.LastInsertId(), int64(1))
		assert.Equal(t, "John Doe", result["name"])
		assert.Equal(t, int64(30), result["age"])
		assert.Equal(t, time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC), result["created_at"])
		assert.Equal(t, time.Date(2023, time.December, 1, 12, 0, 0, 0, time.UTC), result["updated_at"])

		r, e := conn.Table("scanmap").Insert([]map[string]interface{}{
			{"name": "Jane Doe", "age": 25, "created_at": "2023-10-02 12:00:00", "updated_at": "2023-12-02 12:00:00"},
			{"name": "Alice Smith", "age": 28, "created_at": "2023-10-03 12:00:00", "updated_at": "2023-12-03 12:00:00"},
		})
		assert.Nil(t, e)
		assert.Equal(t, int64(2), r.LastInsertId())
		assert.Equal(t, r.RowsAffected(), int64(2))
		var results []map[string]interface{}

		r, err = conn.Query().Table("scanmap").Select("name", "age", "created_at", "updated_at").Get(&results)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(results))
		assert.Equal(t, int64(3), r.RowsFetched())
		assert.Equal(t, "John Doe", results[0]["name"])
		assert.Equal(t, int64(30), results[0]["age"])
		assert.Equal(t, time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC), results[0]["created_at"])
		assert.Equal(t, time.Date(2023, time.December, 1, 12, 0, 0, 0, time.UTC), results[0]["updated_at"])
		assert.Equal(t, "Jane Doe", results[1]["name"])
		assert.Equal(t, int64(25), results[1]["age"])
		assert.Equal(t, time.Date(2023, time.October, 2, 12, 0, 0, 0, time.UTC), results[1]["created_at"])
		assert.Equal(t, time.Date(2023, time.December, 2, 12, 0, 0, 0, time.UTC), results[1]["updated_at"])
		assert.Equal(t, "Alice Smith", results[2]["name"])
		assert.Equal(t, int64(28), results[2]["age"])
		assert.Equal(t, time.Date(2023, time.October, 3, 12, 0, 0, 0, time.UTC), results[2]["created_at"])
		assert.Equal(t, time.Date(2023, time.December, 3, 12, 0, 0, 0, time.UTC), results[2]["updated_at"])

	})
}
func TestScanStruct(t *testing.T) {
	type UserStruct struct {
		ID        int
		Name      string
		Age       int
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	after := "drop table if exists scanstruct;"
	before := after + "create table scanstruct (id int auto_increment primary key, name varchar(255), age int, created_at datetime, updated_at datetime);"
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		user := &UserStruct{
			Name:      "John Doe",
			Age:       30,
			CreatedAt: time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, time.December, 1, 12, 0, 0, 0, time.UTC),
		}
		user1 := UserStruct{
			Name:      "Jane Doe",
			Age:       25,
			CreatedAt: time.Date(2023, time.October, 2, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, time.December, 2, 12, 0, 0, 0, time.UTC),
		}
		user2 := UserStruct{
			Name:      "Alice Smith",
			Age:       28,
			CreatedAt: time.Date(2023, time.October, 3, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, time.December, 3, 12, 0, 0, 0, time.UTC),
		}
		r, err := conn.Query().Table("scanstruct").Insert(user)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r.LastInsertId())
		assert.Equal(t, int64(1), r.RowsAffected())
		r, err = conn.Query().Table("scanstruct").Insert([]UserStruct{user1, user2})
		assert.Nil(t, err)
		assert.Equal(t, int64(2), r.LastInsertId())
		assert.Equal(t, int64(2), r.RowsAffected())
		var scan1 UserStruct
		var users []UserStruct
		r, err = conn.Query().Table("scanstruct").First(&scan1)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r.RowsFetched())
		assert.Equal(t, "John Doe", scan1.Name)
		assert.Equal(t, 30, scan1.Age)
		assert.Equal(t, time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC), scan1.CreatedAt)
		assert.Equal(t, time.Date(2023, time.December, 1, 12, 0, 0, 0, time.UTC), scan1.UpdatedAt)
		r, err = conn.Query().Table("scanstruct").Where("id", ">", 1).Select("name", "age", "created_at", "updated_at").Get(&users)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), r.RowsFetched())
		assert.Equal(t, 2, len(users))

		assert.Equal(t, "Jane Doe", users[0].Name)
		assert.Equal(t, 25, users[0].Age)
		assert.Equal(t, time.Date(2023, time.October, 2, 12, 0, 0, 0, time.UTC), users[0].CreatedAt)
		assert.Equal(t, time.Date(2023, time.December, 2, 12, 0, 0, 0, time.UTC), users[0].UpdatedAt)
		assert.Equal(t, "Alice Smith", users[1].Name)
		assert.Equal(t, 28, users[1].Age)
		assert.Equal(t, time.Date(2023, time.October, 3, 12, 0, 0, 0, time.UTC), users[1].CreatedAt)
		assert.Equal(t, time.Date(2023, time.December, 3, 12, 0, 0, 0, time.UTC), users[1].UpdatedAt)

	})

}
func TestScanValues(t *testing.T) {
	after := "drop table if exists scanvalues;"
	before := `drop table if exists scanvalues;create table scanvalues (id int auto_increment primary key, name varchar(255), age int);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {
		conn.Query().Table("scanvalues").Insert([]interface{}{
			map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			},
			map[string]interface{}{
				"name": "Jane Doe",
				"age":  25,
			},
			map[string]interface{}{
				"name": "Alice Smith",
				"age":  28,
			},
		})

		var names []string
		var ages []int
		r, err := conn.Query().Table("scanvalues").Select("name").Get(&names)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), r.RowsFetched())
		assert.Equal(t, []string{"John Doe", "Jane Doe", "Alice Smith"}, names)
		r, err = conn.Query().Table("scanvalues").Where("age", ">", 25).Select("age").Get(&ages)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), r.RowsFetched())
		assert.Equal(t, []int{30, 28}, ages)

	})
}

type TagString struct {
	Tags []string `json:"tags"`
}
type ModelOptions struct {
	Theme      string   `json:"theme"`
	Roles      []string `json:"roles"`
	JoinedYear int      `json:"joined_year"`
	Address    struct {
		Country string `json:"country"`
		City    string `json:"city"`
	}
	IsBaned bool `json:"is_baned"`
}

func (ts *TagString) Scan(value interface{}) error {
	if value == nil {
		ts.Tags = nil
		return nil
	}
	str := string(value.([]byte))
	if str == "" {
		ts.Tags = nil
	} else {
		ts.Tags = strings.Split(str, ",")
	}
	return nil
}
func (ts TagString) Value() (driver.Value, error) {
	if ts.Tags == nil {
		return "", nil
	}
	return strings.Join(ts.Tags, ","), nil
}
func (mo *ModelOptions) Scan(value interface{}) error {
	bs, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bs, mo)
}
func (mo ModelOptions) Value() (driver.Value, error) {

	bs, err := json.Marshal(mo)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

type TestModel struct {
	*goeloquent.EloquentModel
	Id      int          `goelo:"column:id;primaryKey"`
	Tags    TagString    `goelo:"column:tags;"`
	Options ModelOptions `goelo:"column:options;"`
}

func (m *TestModel) GetTableName(st *goeloquent.Statement) string {
	return "test_models"
}
func TestValuerScanner(t *testing.T) {

	after := "drop table if exists test_model;"
	before := after + `create table test_model (
id int auto_increment primary key,
tags varchar(255),
options json
);`
	RunWithDB(before, after, func(conn goeloquent.Connection) {

		r, err := conn.Query().Insert(&TestModel{
			Tags: TagString{
				Tags: []string{"tag1", "tag2", "tag3"},
			},
			Options: ModelOptions{
				Theme:      "dark",
				Roles:      []string{"admin", "user"},
				JoinedYear: 2023,
				Address: struct {
					Country string `json:"country"`
					City    string `json:"city"`
				}{
					Country: "USA",
					City:    "New York",
				},
				IsBaned: false,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r.RowsAffected())

		var model TestModel
		r, err = conn.Query().Table("test_model").First(&model)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), r.RowsFetched())
		assert.Equal(t, 1, model.Id)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, model.Tags.Tags)
		assert.Equal(t, "dark", model.Options.Theme)
		assert.Equal(t, []string{"admin", "user"}, model.Options.Roles)
		assert.Equal(t, 2023, model.Options.JoinedYear)
		assert.Equal(t, "USA", model.Options.Address.Country)
		assert.Equal(t, "New York", model.Options.Address.City)
		assert.False(t, model.Options.IsBaned)

		conn.Query().Table("test_models").Truncate()

		r, err = conn.Query().Insert([]TestModel{
			{
				Tags:    TagString{},
				Options: ModelOptions{},
			},
			{
				Tags: TagString{
					Tags: []string{"tag4", "tag5"},
				},
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, "insert into `test_model` (`options`, `tags`) values (?, ?), (?, ?)", r.RawSql)
		assert.Equal(t, r.GetBindings(), []interface{}{nil, nil, nil, "tag4,tag5"})

		var models = [1]TestModel{}
		r, err = conn.Query().Table("test_model").Get(&models)

		assert.Nil(t, err)
		assert.Equal(t, int64(3), r.RowsFetched())
		assert.Equal(t, 1, len(models))
		assert.Equal(t, 1, models[0].Id)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, models[0].Tags.Tags)
		assert.Equal(t, ModelOptions{
			Theme:      "dark",
			Roles:      []string{"admin", "user"},
			JoinedYear: 2023,
			Address: struct {
				Country string `json:"country"`
				City    string `json:"city"`
			}{
				Country: "USA",
				City:    "New York",
			},
			IsBaned: false,
		}, models[0].Options)
		var models1 []TestModel
		r, err = conn.Query().Table("test_model").Get(&models1)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), r.RowsFetched())
		assert.Equal(t, 3, len(models1))
		assert.Equal(t, 0, len(models1[1].Tags.Tags))
		assert.True(t, reflect.ValueOf(models1[1].Options).IsZero())
		assert.Equal(t, []string{"tag4", "tag5"}, models1[2].Tags.Tags)
		assert.Equal(t, ModelOptions{}, models1[2].Options)
	})
}
