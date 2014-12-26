/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

// Package db provides configuration for database connection
package db

import (
	"fmt"
	"sync"

	"github.com/gluee/backend/config"

	_ "github.com/go-sql-driver/mysql" // Get the MySQL driver
	"github.com/jmoiron/sqlx"
)

type (
	// Keep the database connection and the number of times it was used
	dbSlave struct {
		Usage uint64
		DB    *sqlx.DB
	}

	// Keep the database slave connections and a mutex to be able to safely keep track of the least used connection
	dbSlaves struct {
		sync.Mutex
		MinSlave int
		Slaves   []*dbSlave
	}
)

var (
	masterConnection = &sqlx.DB{}
	slaveConnections = &dbSlaves{}
)

/**
 * Open a connection to master server
 * @param cfg, configuration that holds database settings
 */
func openMasterConnection(cfg *config.Config) {
	// Read settings from configuration
	masterDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8&collation=utf8_general_ci",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Master.Host,
		cfg.DB.Master.Port,
		cfg.DB.Database,
	)

	// Establish connection to master
	masterConnection = sqlx.MustConnect("mysql", masterDSN)
	masterConnection.DB.Ping()
	masterConnection.DB.SetMaxIdleConns(10)
	masterConnection.DB.SetMaxOpenConns(100)
}

/**
 * Open the connections to the slave servers
 * @param cfg, configuration that holds database settings
 */
func openSlaveConnections(cfg *config.Config) {
	for _, slave := range cfg.DB.Slaves {
		slaveConnection := &sqlx.DB{}
		slaveDSN := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8&collation=utf8_general_ci",
			cfg.DB.Username,
			cfg.DB.Password,
			slave.Host,
			slave.Port,
			cfg.DB.Database,
		)

		// Establish connection to slaves
		slaveConnection = sqlx.MustConnect("mysql", slaveDSN)
		slaveConnection.DB.Ping()
		slaveConnection.DB.SetMaxIdleConns(10)
		slaveConnection.DB.SetMaxOpenConns(100)

		slaveConnections.Slaves = append(slaveConnections.Slaves, &dbSlave{DB: slaveConnection})
	}
}

/**
 * InitDatabases initializes the connections to the servers
 * @param cfg, configuration that holds database settings
 */
func InitDatabases(cfg *config.Config) {
	openMasterConnection(cfg)
	openSlaveConnections(cfg)
}

/**
 * GetMaster is used to write to the database.
 * @return masterConnection, master database connection
 */
func GetMaster() *sqlx.DB {
	return masterConnection
}

/**
 * GetSlave is used to read from database.
 * If there's no slave configured, it returns master.
 * @return slaveConnections, slave database connection from the connection pool
 */
func GetSlave() *sqlx.DB {
	if len(slaveConnections.Slaves) == 0 {
		return masterConnection
	}

	// Make sure we don't select the wrong master after we finish counting the current least used one
	slaveConnections.Lock()

	min := slaveConnections.MinSlave
	slaveConnections.Slaves[min].Usage = slaveConnections.Slaves[min].Usage + 1
	newMin := min
	minVal := slaveConnections.Slaves[min].Usage

	// Find the least used slave
	for key, slave := range slaveConnections.Slaves {
		if slave.Usage < minVal {
			newMin = key
			minVal = slave.Usage
		}
	}
	slaveConnections.MinSlave = newMin

	slaveConnections.Unlock()

	return slaveConnections.Slaves[min].DB
}
