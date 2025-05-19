package goeloquent

const (
	DriverMysql     = "mysql"
	DriverPostgres  = "postgres"
	DriverSqlite    = "sqlite"
	DriverSqlServer = "sqlserver"
	DriverMongoDB   = "mongodb"
)

type Driver string

type DBConfig struct {
	Driver Driver `json:"driver"`
	DSN    string `json:"dsn"`
}
