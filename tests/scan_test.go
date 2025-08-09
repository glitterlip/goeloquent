package tests

import (
	"github.com/glitterlip/goeloquent/v2"
	"github.com/stretchr/testify/assert"
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