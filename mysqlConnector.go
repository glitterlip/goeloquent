package goeloquent

import (
	"database/sql"
	"fmt"
)

type MysqlConnector struct {
}

func (c MysqlConnector) connect(config DBConfig) *MysqlConnection {
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

	db, err := sql.Open(DriverMysql, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.Username, config.Password, config.Host, config.Port, config.Database))
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return &MysqlConnection{
		Connection: db,
		Config:     config,
	}
}
