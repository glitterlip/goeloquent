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
	Driver               Driver `json:"driver"`
	DSN                  string `json:"dsn"`
	TablePrefix          string `json:"table_prefix,omitempty"`
	DisableAutomaticPing bool   `json:"disable_automatic_ping"`
}
