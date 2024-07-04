package tests

import (
	"time"
)

const (
	FriendStatusWaiting = 1
	FriendStatusNormal  = 2
	FriendStatusDeleted = 3
	FriendStatusBlocked = 4
)

type Friends struct {
	ID         int64     `goelo:"column:id;primaryKey"`
	UserId     int64     `goelo:"column:user_id"`
	FriendId   int64     `goelo:"column:friend_id"`
	Status     int       `goelo:"column:status"`
	Time       time.Time `goelo:"column:time"`
	Additional string    `goelo:"column:additional"`
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
