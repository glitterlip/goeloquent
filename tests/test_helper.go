package tests

import (
	"fmt"
	goeloquent "github.com/glitterlip/go-eloquent"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var DB *goeloquent.DB

func Setup() {
	fmt.Println("test setup")
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
	}
	DB = goeloquent.Open(defaultConfig)
	chatConfig := GetChatConfig()
	DB.AddConfig("chat", &chatConfig)

}
func init() {
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
	}
	DB = goeloquent.Open(defaultConfig)
	DB.SetLogger(func(l goeloquent.Log) {
		fmt.Println(l)
	})
	chatConfig := GetChatConfig()
	DB.AddConfig("chat", &chatConfig)
}

func GetDefaultConfig() goeloquent.DBConfig {
	return goeloquent.DBConfig{
		Host:      os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_HOST"),
		Port:      os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_PORT"),
		Database:  os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_DATABASE"),
		Username:  os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_USERNAME"),
		Password:  os.Getenv("GOELOQUENT_TEST_DEFAUTL_DB_PASSWORD"),
		Driver:    "mysql",
		EnableLog: true,
	}
}
func GetChatConfig() goeloquent.DBConfig {
	return goeloquent.DBConfig{
		Host:     os.Getenv("GOELOQUENT_TEST_SECOND_DB_HOST"),
		Port:     os.Getenv("GOELOQUENT_TEST_SECOND_DB_PORT"),
		Database: os.Getenv("GOELOQUENT_TEST_SECOND_DB_DATABASE"),
		Username: os.Getenv("GOELOQUENT_TEST_SECOND_DB_USERNAME"),
		Password: os.Getenv("GOELOQUENT_TEST_SECOND_DB_PASSWORD"),
		Driver:   "mysql",
	}
}
func ShouldEqual(t *testing.T, expected interface{}, b *goeloquent.Builder) {
	assert.Equal(t, expected, b.ToSql())
}
func ElementsShouldMatch(t *testing.T, a interface{}, b interface{}) {
	assert.ElementsMatch(t, a, b)
}

func RunWithDB(create, drop string, test func()) {
	defer func() {
		_, err := DB.Raw("default").Exec(strings.ReplaceAll(drop, `"`, "`"))
		if err != nil {
			return
		}
	}()
	DB.Raw("default").Exec(strings.ReplaceAll(create, `"`, "`"))
	test()

}
