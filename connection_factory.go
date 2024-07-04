package goeloquent

import (
	"fmt"
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
	case "":
		panic("a driver must be specified")
	default:
		panic(fmt.Sprintf("unsupported driver:%s", config.Driver))
	}
	return nil
}
