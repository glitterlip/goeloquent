package goeloquent

type Driver string

const (
	DriverMysql Driver = "mysql"
)

type DBConfig struct {
	Driver          Driver
	Name            string
	ReadHost        []string
	WriteHost       []string
	Host            string
	Port            string
	Database        string
	Username        string
	Password        string
	Charset         string
	Prefix          string
	ConnMaxLifetime int
	ConnMaxIdleTime int
	MaxIdleConns    int
	MaxOpenConns    int
	ParseTime       bool
	// 	mysql
	Collation       string
	UnixSocket      string
	MultiStatements bool
	Dsn             string
	DsnExtraString  string
	Strict          bool
	Mode            string
	IsolationLevel  string
	// pgsql
	Sslmode   string
	TLS       string
	EnableLog bool
}
