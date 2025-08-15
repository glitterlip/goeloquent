package goeloquent

import (
	"database/sql"
	"sync"
)

var DB *DatabaseManager

func init() {

	DB = &DatabaseManager{
		Connections: sync.Map{},
	}
}

func Open(name string, config DBConfig) (Connection, error) {
	if conn, ok := DB.Connections.Load(name); ok {
		return conn.(*MysqlConnection), nil
	}
	db, err := sql.Open(string(config.Driver), config.DSN)
	if err != nil {
		return nil, err
	}
	if !config.DisableAutomaticPing {
		if err := db.Ping(); err != nil {
			return nil, err
		}
	}

	connection := MysqlConnection{
		DBConfig: &config,
		DB:       db,
	}
	DB.Connections.Store(name, &connection)
	return &connection, nil
}
