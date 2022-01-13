package goeloquent

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type MysqlConnector struct {
}

func (c MysqlConnector) connect(config *DBConfig) *Connection {
	/**
	[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...¶mN=valueN]
	// user@unix(/path/to/socket)/dbname
	// root:pw@unix(/tmp/mysql.sock)/myDatabase?loc=Local
	// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
	// user:password@/dbname?sql_mode=TRADITIONAL
	// user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?timeout=90s&collation=utf8mb4_unicode_ci
	// id:password@tcp(your-amazonaws-uri.com:3306)/dbname
	// user@cloudsql(project-id:instance-name)/dbname
	// user@cloudsql(project-id:regionname:instance-name)/dbname
	// user:password@tcp/dbname?charset=utf8mb4,utf8&sys_var=esc%40ped
	// user:password@/dbname
	// user:password@/
	*/

	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 10
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 86400
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 7200
	}
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.Collation == "" {
		config.Collation = "utf8mb4_unicode_ci"
	}
	db, err := sql.Open(DriverMysql, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=%s&collation=%s", config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset, config.Collation))
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)
	return &Connection{
		DB:     db,
		Config: config,
	}
}
