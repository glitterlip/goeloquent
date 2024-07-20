package tests

import (
	"fmt"
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var DB *goeloquent.DatabaseManager

func Setup() {
	fmt.Println("test setup")
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
	}
	DB = goeloquent.Open(defaultConfig)
	chatConfig := GetChatConfig()
	DB.AddConfig("chat", &chatConfig)

	DB.Listen(goeloquent.EventExecuted, func(result goeloquent.Result) {
		fmt.Println(result.Sql)
		fmt.Println(result.Bindings)
		fmt.Println(result.Time)
	})
}
func init() {
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
	}
	DB = goeloquent.Open(defaultConfig)
	chatConfig := GetChatConfig()
	DB.AddConfig("chat", &chatConfig)
	DB.Listen(goeloquent.EventExecuted, func(result goeloquent.Result) {
		fmt.Println(result.Sql)
		fmt.Println(result.Bindings)
		fmt.Println(result.Time)
	})
}

func GetDefaultConfig() goeloquent.DBConfig {
	return goeloquent.DBConfig{
		Host:            "127.0.0.1",
		Port:            "8889",
		Database:        "goeloquent",
		Username:        "root",
		Password:        "root",
		MultiStatements: true,
		Driver:          "mysql",
		EnableLog:       true,
		ParseTime:       true,
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

func RunWithDB(create, drop interface{}, test func()) {
	defer func() {
		if strs, ok := drop.([]string); ok {
			for _, str := range strs {
				DB.Raw("default").Exec(strings.ReplaceAll(str, `"`, "`"))
			}
		} else {
			DB.Raw("default").Exec(strings.ReplaceAll(drop.(string), `"`, "`"))
		}
	}()
	if strs, ok := create.([]string); ok {
		for _, str := range strs {
			_, err := DB.Raw("default").Exec(strings.ReplaceAll(str, `"`, "`"))
			if err != nil {
				panic(err.Error())
				return
			}
		}
	} else {
		_, err := DB.Raw("default").Exec(strings.ReplaceAll(create.(string), `"`, "`"))
		if err != nil {
			panic(err.Error())

		}
	}
	test()

}

func UserTableSql() (create, drop string) {
	create = `
DROP TABLE IF EXISTS users;
CREATE TABLE "users" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "name" varchar(255) NOT NULL,
  "info" json NULL,
  "roles" varchar(255),
  "age" tinyint(10) unsigned NOT NULL DEFAULT '0',
  "status" tinyint(10) unsigned NOT NULL DEFAULT '0',
  "created_at" datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  "deleted_at" datetime DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4;
`
	drop = `DROP TABLE IF EXISTS users`
	return
}
func CreateUsers() {
	DB.Raw("default").Exec(`
truncate table user_models;
INSERT INTO user_models (id, no, name, email, status, age, info, tags, created_at, updated_at, deleted_at) VALUES (1, '1', 'qwe', 'qwe', 1, 12, '{"Age": 18, "Links": ["https://x.com", "https://twitch.com"], "Address": "NY", "Verified": true}', 'tag1,tag2', null, null, null);
INSERT INTO user_models (id, no, name, email, status, age, info, tags, created_at, updated_at, deleted_at) VALUES (2, '2', 'asd', 'asd', 2, 12, '{"Age": 18, "Links": ["https://x.com", "https://twitch.com"], "Address": "NY", "Verified": true}', 'tag1,tag2', null, null, null);
INSERT INTO user_models (id, no, name, email, status, age, info, tags, created_at, updated_at, deleted_at) VALUES (3, '3', 'zxc', 'zxc', 1, 12, '{"Age": 18, "Links": ["https://x.com", "https://twitch.com"], "Address": "NY", "Verified": true}', 'tag1,tag2', null, null, null); 
INSERT INTO user_models (id, no, name, email, status, age, info, tags, created_at, updated_at, deleted_at) VALUES (4, '4', '123', '123', 2, 12, '{"Age": 18, "Links": ["https://x.com", "https://twitch.com"], "Address": "NY", "Verified": true}', 'tag1,tag2', null, null, null); 
INSERT INTO user_models (id, no, name, email, status, age, info, tags, created_at, updated_at, deleted_at) VALUES (5, '5', 'qaz', 'qaz', 1, 12, '{"Age": 18, "Links": ["https://x.com", "https://twitch.com"], "Address": "NY", "Verified": true}', 'tag1,tag2', null, null, null); 
`)
}
