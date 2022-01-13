package goeloquent

const (
	DriverMysql = "mysql"
)

type ConnectionFactory struct {
}
type Connector interface {
	connect(config *DBConfig) *Connection
}

func (f ConnectionFactory) Make(config *DBConfig) *Connection {
	return f.MakeConnection(config)
}
func (f ConnectionFactory) MakeConnection(config *DBConfig) *Connection {
	return f.CreateConnection(config)
}
func (f ConnectionFactory) CreateConnection(config *DBConfig) *Connection {
	switch config.Driver {
	case DriverMysql:
		connector := MysqlConnector{}
		conn := connector.connect(config)
		return conn
	}
	return nil
}
