package tests

import (
	"github.com/glitterlip/goeloquent"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

//DatabaseConnectionFactoryTest

func TestOpenFailed(t *testing.T) {

	openCases := map[string]interface{}{
		//testIfDriverIsntSetExceptionIsThrown
		"missing driver": map[string]interface{}{
			"msg": "a driver must be specified",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Host: "mssql",
				},
			},
		},
		//testExceptionIsThrownOnUnsupportedDriver
		"unsupported driver": map[string]interface{}{
			"msg": "unsupported driver:mssql",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mssql",
				},
			},
		},
		"missing default config": map[string]interface{}{
			"msg": "Database connection default not configured.",
			"config": map[string]goeloquent.DBConfig{
				"defaul": {
					Driver: "mssql",
				},
			},
		},
		"wrong host": map[string]interface{}{
			"msg": "dial tcp: lookup test: no such host",
			"config": map[string]goeloquent.DBConfig{
				"default": {
					Driver: "mysql",
					Host:   "test",
				},
			},
		},
	}
	for caseName, i := range openCases {
		m, ok := i.(map[string]interface{})
		if ok {
			errStr := m["msg"].(string)
			config := m["config"].(map[string]goeloquent.DBConfig)
			assert.PanicsWithErrorf(t, errStr, func() {
				_ = goeloquent.Open(config)
			}, "TestOpen case %s failed", caseName)
		}
	}

}

func TestOpenSuccess(t *testing.T) {
	Setup()
	for name, dbConfig := range DB.GetConfigs() {
		var dbName string
		DB.Connection(name).GetDB().QueryRow("SELECT DATABASE()").Scan(&dbName)
		assert.Equal(t, dbConfig.Database, dbName)
	}
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
	//testSingleConnectionNotCreatedUntilNeeded
	Setup()
	assert.Equal(t, 1, len(DB.Connections))
	_, ok := DB.Connections["chat"]
	assert.False(t, ok)
}
func TestDSNConfig(t *testing.T) {
	DB.AddConfig("test", &goeloquent.DBConfig{
		Driver: "mysql",
		Dsn:    os.Getenv("GOELOQUENT_TEST_DEFAULT_DSN"),
	})
	conn := DB.Connection("test")
	err := conn.DB.Ping()
	assert.Nil(t, err)
}

//TODO: testCustomConnectorsCanBeResolvedViaContainer
//TODO: testSqliteForeignKeyConstraints
//TODO: testReadWriteConnectionsNotCreatedUntilNeeded

//DatabaseConnectionTest
//TODO: testSettingDefaultCallsGetDefaultGrammar
//TODO: testSettingDefaultCallsGetDefaultPostProcessor
//TODO: testSelectOneCallsSelectAndReturnsSingleResult
//TODO: testSelectProperlyCallsPDO
//TODO: testInsertCallsTheStatementMethod
//TODO: testUpdateCallsTheAffectingStatementMethod
//TODO: testDeleteCallsTheAffectingStatementMethod
//TODO: testStatementProperlyCallsPDO
//TODO: testAffectingStatementProperlyCallsPDO
//TODO: testTransactionLevelNotIncrementedOnTransactionException
//TODO: testBeginTransactionMethodRetriesOnFailure
//TODO: testBeginTransactionMethodReconnectsMissingConnection
//TODO: testBeginTransactionMethodNeverRetriesIfWithinTransaction
//TODO: testSwapPDOWithOpenTransactionResetsTransactionLevel
//TODO: testBeganTransactionFiresEventsIfSet
//TODO: testCommittedFiresEventsIfSet
//TODO: testRollBackedFiresEventsIfSet
//TODO: testRedundantRollBackFiresNoEvent
//TODO: testTransactionMethodRunsSuccessfully
//TODO: testTransactionRetriesOnSerializationFailure
//TODO: testTransactionMethodRetriesOnDeadlock
//TODO: testTransactionMethodRollsbackAndThrows
//TODO: testOnLostConnectionPDOIsNotSwappedWithinATransaction
//TODO: testOnLostConnectionPDOIsSwappedOutsideTransaction
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
