package connectors

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
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
	//TODO: Protocol loc readTimeout serverPubKey timeout
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

	db := c.CreateConnection(c.GetDsn(config))

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)
	return &Connection{
		DB:     db,
		Config: config,
	}

}
func (c MysqlConnector) CreateConnection(dsn string) *sql.DB {
	db, err := sql.Open(string(DriverMysql), dsn)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return db
}
func (c MysqlConnector) ConfigureIsolationLevel(db *sql.DB, config DBConfig) {
	if len(config.IsolationLevel) > 0 {
		db.Exec("SET SESSION TRANSACTION ISOLATION LEVEL " + config.IsolationLevel)
	}
}
func (c MysqlConnector) ConfigureMode(db *sql.DB, config DBConfig) {
	if len(config.Mode) > 0 {
		db.Exec("SET SESSION sql_mode = '" + config.Mode + "'")
	} else if config.Strict {
		db.Exec("set session sql_mode='ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION'")
	}
}

func (c MysqlConnector) GetDsn(config *DBConfig) string {
	var dsn string
	var params []string

	if len(config.Charset) > 0 {
		params = append(params, "charset="+config.Charset)
	}
	if len(config.Collation) > 0 {
		params = append(params, "collation="+config.Collation)
	}
	if config.MultiStatements {
		params = append(params, "multiStatements=true")
	}
	if config.ParseTime {
		params = append(params, "parseTime=true")
	}
	if len(config.Dsn) > 0 {
		dsn = config.Dsn
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", config.Username, config.Password, config.Host, config.Port, config.Database, strings.Join(params, "&"))
		if len(config.DsnExtraString) > 0 {
			dsn = dsn + "&" + config.DsnExtraString
		}
	}
	return dsn
}
