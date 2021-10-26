package goeloquent

type DBConfig struct {
	Driver string
	//ReadHost  []string
	//WriteHost []string
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Charset  string
	Prefix   string
	// 	mysql
	Collation  string
	UnixSocket string
	// pgsql
	Sslmode string
}
