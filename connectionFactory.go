package goeloquent

const (
	DriverMysql = "mysql"
)

type ConnectionFactory struct {
}
type Connector interface {
	connect(config DBConfig) IConnection
}

func (f ConnectionFactory) Make(config DBConfig) IConnection {
	return f.MakeConnection(config)
}
func (f ConnectionFactory) MakeConnection(config DBConfig) IConnection {
	return f.CreateConnection(config)
}
func (f ConnectionFactory) CreateConnection(config DBConfig) IConnection {
	switch config.Driver {
	case DriverMysql:
		connector := MysqlConnector{}
		conn := connector.connect(config)
		return &*conn
	}
	return nil
}
