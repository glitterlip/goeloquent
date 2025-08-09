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
