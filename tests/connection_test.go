package tests

import (
	"fmt"
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var db *goeloquent.DB

func setup() {
	fmt.Println("test setup")
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": getDefaultConfig(),
	}
	db = goeloquent.Open(defaultConfig)
	chatConfig := getChatConfig()
	db.AddConfig("chat", &chatConfig)

}
func getDefaultConfig() goeloquent.DBConfig {
	return goeloquent.DBConfig{
		Host:     os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_HOST"),
		Port:     os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_PORT"),
		Database: os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_DATABASE"),
		Username: os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_USERNAME"),
		Password: os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_PASSWORD"),
		Driver:   "mysql",
	}
}
func getChatConfig() goeloquent.DBConfig {
	return goeloquent.DBConfig{
		Host:     os.Getenv("GOELOQUENT_TEST_SECOND_DB_HOST"),
		Port:     os.Getenv("GOELOQUENT_TEST_SECOND_DB_PORT"),
		Database: os.Getenv("GOELOQUENT_TEST_SECOND_DB_DATABASE"),
		Username: os.Getenv("GOELOQUENT_TEST_SECOND_DB_USERNAME"),
		Password: os.Getenv("GOELOQUENT_TEST_SECOND_DB_PASSWORD"),
		Driver:   "mysql",
	}
}
func teardown() {
	fmt.Println("test teardown")
}
func DB() *goeloquent.DB {
	return db
}

func TestOpenFailed(t *testing.T) {

	openCases := map[string]interface{}{
		//testIfDriverIsntSetExceptionIsThrown
		"missing driver": map[string]interface{}{
			"msg": "a driver must be specified",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Host: "mssql",
				},
			},
		},
		//testExceptionIsThrownOnUnsupportedDriver
		"unsupported driver": map[string]interface{}{
			"msg": "unsupported driver:mssql",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mssql",
				},
			},
		},
		"missing default config": map[string]interface{}{
			"msg": "Database connection default not configured.",
			"config": map[string]goeloquent.DBConfig{
				"defaul": {
					Driver: "mssql",
				},
			},
		},
		"wrong host": map[string]interface{}{
			"msg": "dial tcp: lookup test: no such host",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mysql",
					Host:   "test",
				},
			},
		},
	}
	for caseName, i := range openCases {
		m, ok := i.(map[string]interface{})
		if ok {
			errStr := m["msg"].(string)
			config := m["config"].(map[string]goeloquent.DBConfig)
			assert.PanicsWithErrorf(t, errStr, func() {
				_ = goeloquent.Open(config)
			}, "TestOpen case %s failed", caseName)
		}
	}

}

func TestOpenSuccess(t *testing.T) {
	setup()
	for name, dbConfig := range db.GetConfigs() {
		var dbName string
		if name == "default" {
			db.Connection(name).GetDB().QueryRow("SELECT DATABASE()").Scan(&dbName)
			assert.Equal(t, dbConfig.Database, dbName)
		} else {
			mysqlErr := func() (e error) {
				defer func() {
					err := recover()
					if err != nil {
						e = err.(error)
					}
				}()
				db.Connection(name).GetDB().QueryRow("SELECT DATABASE()").Scan(&dbName)
				return
			}
			e := mysqlErr()
			assert.IsType(t, &mysql.MySQLError{
				Number:  0,
				Message: "",
			}, e)
		}

	}
}

func TestConnectionCanBeCreated(t *testing.T) {
	//testConnectionCanBeCreated
	setup()
	conn := goeloquent.Connection{}
	assert.IsType(t, DB().Conn("default"), conn)
	assert.IsType(t, DB().Connection("chat"), &conn)
}
func TestConnectionHasProperConfig(t *testing.T) {
	//testConnectionFromUrlHasProperConfig
	setup()
	configs := map[string]goeloquent.DBConfig{
		"default": getDefaultConfig(),
		"chat":    getChatConfig(),
	}
	for name, config := range configs {
		assert.Equal(t, config.Host, (*db.GetConfig(name)).Host)
		assert.Equal(t, config.Port, (*db.GetConfig(name)).Port)
		assert.Equal(t, config.Database, (*db.GetConfig(name)).Database)
		assert.Equal(t, config.Username, (*db.GetConfig(name)).Username)
		assert.Equal(t, config.Password, (*db.GetConfig(name)).Password)
	}
}

func TestSingleConnectionNotCreatedUntilNeeded(t *testing.T) {
	//testSingleConnectionNotCreatedUntilNeeded
	setup()
	assert.Equal(t, 1, len(db.Connections))
	_, ok := db.Connections["chat"]
	assert.False(t, ok)
}
