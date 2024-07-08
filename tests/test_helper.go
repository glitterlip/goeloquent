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

}
func init() {
	defaultConfig := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
	}
	DB = goeloquent.Open(defaultConfig)
	chatConfig := GetChatConfig()
	DB.AddConfig("chat", &chatConfig)
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
func CreateRelationTables() (create, drop string) {
	create = `
DROP TABLE IF EXISTS "comment","image","post","tag","tagables","user","users","view_record","friends","address";
CREATE TABLE "comment" (
  "cid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "post_id" int(11) NOT NULL,
  "content" varchar(255) DEFAULT NULL,
  "user_id" int(11) NOT NULL,
  "upvotes" int(11) NOT NULL DEFAULT '0',
  "downvotes" int(11) NOT NULL DEFAULT '0',
  "comment_id" int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY ("cid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "image" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "url" varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  "imageable_id" int(11) DEFAULT NULL,
  "imageable_type" varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "post" (
  "pid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "title" varchar(255) NOT NULL DEFAULT '',
  "author_id" int(10) unsigned NOT NULL DEFAULT '0',
  "created_at" datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL,
  "deleted_at" datetime DEFAULT NULL,
  PRIMARY KEY ("pid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "tag" (
  "tid" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "name" varchar(255) DEFAULT '',
  PRIMARY KEY ("tid") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "tagables" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "tag_id" int(11) NOT NULL,
  "tagable_id" int(11) DEFAULT NULL,
  "status" int(10) unsigned NOT NULL DEFAULT '0',
  "tagable_type" varchar(255) DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "user" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "age" int(10) unsigned NOT NULL DEFAULT '0',
  "name" varchar(255) NOT NULL DEFAULT '',
  "created_at" datetime DEFAULT CURRENT_TIMESTAMP,
  "updated_at" datetime DEFAULT NULL,
  PRIMARY KEY ("id") USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "view_record" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(11) DEFAULT NULL,
  "post_id" int(11) DEFAULT NULL,
  "view_time" datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "friends" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(11) NOT NULL,
  "friend_id" int(11) NOT NULL,
  "status" tinyint(4) NOT NULL,
  "time" datetime DEFAULT NULL,
  "additional" varchar(255) DEFAULT '',
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE "address" (
  "id" int(10) unsigned NOT NULL AUTO_INCREMENT,
  "user_id" int(10) unsigned NOT NULL,
  "country" tinyint(4) NOT NULL,
  "state" varchar(255) NOT NULL,
  "city" varchar(255) NOT NULL,
  "detail" varchar(255) DEFAULT NULL,
  PRIMARY KEY ("id")
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`
	drop = `
DROP TABLE IF EXISTS "comment","image","post","tag","tagables","user","users","view_record","friends","address";
`
	return
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
