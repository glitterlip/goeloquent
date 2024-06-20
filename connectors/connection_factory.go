package connectors

import (
	"errors"
	"fmt"
	"github.com/glitterlip/goeloquent"
)

type ConnectionFactory struct {
}
type Connector interface {
	connect(config *DBConfig) *goeloquent.Connection
}

func (f ConnectionFactory) Make(config *DBConfig) *goeloquent.Connection {
	return f.MakeConnection(config)
}
func (f ConnectionFactory) MakeConnection(config *DBConfig) *goeloquent.Connection {
	return f.CreateConnection(config)
}
func (f ConnectionFactory) CreateConnection(config *DBConfig) *goeloquent.Connection {
	switch config.Driver {
	case DriverMysql:
		connector := MysqlConnector{}
		conn := connector.connect(config)
		return conn
	case "":
		panic(errors.New("a driver must be specified"))
	default:
		panic(errors.New(fmt.Sprintf("unsupported driver:%s", config.Driver)))
	}
	return nil
}
