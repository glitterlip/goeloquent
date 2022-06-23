package goeloquent

type DBConfig struct {
	Driver string
	//ReadHost  []string
	//WriteHost []string
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
	// pgsql
	Sslmode   string
	TLS       string
	EnableLog bool
}
