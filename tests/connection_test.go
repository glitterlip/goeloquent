package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenFailed(t *testing.T) {

	openCases := map[string]map[string]interface{}{
		//testIfDriverIsntSetExceptionIsThrown
		"missing driver": {
			"msg": "a driver must be specified",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Host: "mssql",
				},
			},
		},
		//testExceptionIsThrownOnUnsupportedDriver
		"unsupported driver": {
			"msg": "unsupported driver:mssql",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mssql",
				},
			},
		},
		"missing default config": {
			"msg": "Database connection default not configured.",
			"config": map[string]goeloquent.DBConfig{
				"defaul": {
					Driver: "mssql",
				},
			},
		},
		"wrong host": {
			"msg": "dial tcp: lookup test: no such host",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mysql",
					Host:   "test",
				},
			},
		},
	}
	for caseName, m := range openCases {
		errStr := m["msg"].(string)
		config := m["config"].(map[string]goeloquent.DBConfig)
		assert.PanicsWithValuef(t, errStr, func() {
			_ = goeloquent.Open(config)
		}, "TestOpen case %s failed", caseName)
	}

}

func TestOpenSuccess(t *testing.T) {
	Setup()
	var res int
	_, err := DB.Select("SELECT 1 + 1", nil, &res)
	assert.Nil(t, err)
	assert.Equal(t, 2, res)
}

func TestConnectionCanBeCreated(t *testing.T) {
	//testConnectionCanBeCreated
	Setup()
	conn := goeloquent.Connection{}
	assert.IsType(t, DB.Conn("default"), conn)
	assert.IsType(t, DB.Connection("chat"), &conn)
}
func TestConnectionHasProperConfig(t *testing.T) {
	//testConnectionFromUrlHasProperConfig
	Setup()
	configs := map[string]goeloquent.DBConfig{
		"default": GetDefaultConfig(),
		"chat":    GetChatConfig(),
	}
	for name, config := range configs {
		assert.Equal(t, config.Host, (*DB.GetConfig(name)).Host)
		assert.Equal(t, config.Port, (*DB.GetConfig(name)).Port)
		assert.Equal(t, config.Database, (*DB.GetConfig(name)).Database)
		assert.Equal(t, config.Username, (*DB.GetConfig(name)).Username)
		assert.Equal(t, config.Password, (*DB.GetConfig(name)).Password)
	}
}

func TestSingleConnectionNotCreatedUntilNeeded(t *testing.T) {
	Setup()
	assert.Equal(t, 1, len(DB.Connections))
	_, ok := DB.Connections["chat"]
	assert.False(t, ok)
}
func TestDSNConfig(t *testing.T) {
	DB.AddConfig("test", &goeloquent.DBConfig{
		Driver: "mysql",
		Dsn:    "root:root@tcp(127.0.0.1:8889)/goeloquent?charset=utf8mb4&parseTime=true",
	})
	//fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", config.Username, config.Password, config.Host, config.Port, config.Database, strings.Join(params, "&"))
	conn := DB.Connection("test")
	err := conn.DB.Ping()
	assert.Nil(t, err)
}
func TestConnectionEvent(t *testing.T) {

}
func TestConnectionPool(t *testing.T) {
	//TODO:testConnectionPool

}

//TODO: testReadWriteConnectionsNotCreatedUntilNeeded

//DatabaseConnectionTest
//TODO: testSelectOneCallsSelectAndReturnsSingleResult
//TODO: testInsertCallsTheStatementMethod
//TODO: testUpdateCallsTheAffectingStatementMethod
//TODO: testDeleteCallsTheAffectingStatementMethod
//TODO: testTransactionLevelNotIncrementedOnTransactionException
//TODO: testBeginTransactionMethodRetriesOnFailure
//TODO: testBeginTransactionMethodReconnectsMissingConnection
//TODO: testBeginTransactionMethodNeverRetriesIfWithinTransaction
//TODO: testBeganTransactionFiresEventsIfSet
//TODO: testCommittedFiresEventsIfSet
//TODO: testRollBackedFiresEventsIfSet
//TODO: testRedundantRollBackFiresNoEvent
//TODO: testTransactionMethodRunsSuccessfully
//TODO: testTransactionRetriesOnSerializationFailure
//TODO: testTransactionMethodRetriesOnDeadlock
//TODO: testTransactionMethodRollsbackAndThrows
//TODO: testRunMethodRetriesOnFailure
//TODO: testRunMethodNeverRetriesIfWithinTransaction
//TODO: testFromCreatesNewQueryBuilder
//TODO: testPrepareBindings
//TODO: testLogQueryFiresEventsIfSet
//TODO: testBeforeExecutingHooksCanBeRegistered
//TODO: testPretendOnlyLogsQueries
//TODO: testSchemaBuilderCanBeCreated

//DatabaseConnectorTest
//TODO: testOptionResolution
//TODO: testMySqlConnectCallsCreateConnectionWithProperArguments
//TODO: testMySqlConnectCallsCreateConnectionWithIsolationLevel
//TODO: testPostgresConnectCallsCreateConnectionWithProperArguments
//TODO: testPostgresSearchPathIsSet
//TODO: testPostgresSearchPathArraySupported
//TODO: testPostgresApplicationNameIsSet
//TODO: testPostgresConnectorReadsIsolationLevelFromConfig
//TODO: testSQLiteMemoryDatabasesMayBeConnectedTo
//TODO: testSQLiteFileDatabasesMayBeConnectedTo
//TODO: testSqlServerConnectCallsCreateConnectionWithProperArguments
//TODO: testSqlServerConnectCallsCreateConnectionWithOptionalArguments
//TODO: testSqlServerConnectCallsCreateConnectionWithPreferredODBC
