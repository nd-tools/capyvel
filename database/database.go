package database

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/nd-tools/capyvel/foundation"

	"github.com/gookit/color"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

// Global variable for the database connection
var (
	DB Database

	// Custom error messages for database-related issues
	ErrNoConnections             = errors.New("please add at least one or more connections to the database config") // Triggered when no database connections are defined
	ErrNoDefaultConnection       = errors.New("please specify a default connection in the database config")         // Triggered when no default connection is specified
	ErrDefaultConnectionNotFound = errors.New("default connection not found in the database config")                // Triggered when the specified default connection is not found
	ErrConnectionFailed          = errors.New("connection failed")                                                  // Triggered when a database connection cannot be established
	ErrFailedToGetSQLDB          = errors.New("failed to get sqlDB from gorm.DB")                                   // Triggered when unable to retrieve the SQL database object from Gorm
	ErrBindingConnection         = errors.New("error binding connection")                                           // Triggered when binding a specific connection fails
)

// Struct to hold the Gorm database context
type Database struct {
	Ctx *gorm.DB
}

// Checks if a given host string is a valid IP address
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// Builds the DSN (Data Source Name) for SQL Server connections
func buildDSN(DBServer, DBUsername, DBPassword, DBDatabase, DBSsl, DBCharset string) string {
	var dsn string
	if isIPAddress(DBServer) {
		// Use IP-based connection string
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=%s;connection timeout=30;charset=%s", DBServer, DBUsername, DBPassword, DBDatabase, DBSsl, DBCharset)
	} else {
		// Use DNS-based connection string
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s?database=%s&encrypt=%s&connection+timeout=30&charset=%s", DBUsername, DBPassword, DBServer, DBDatabase, DBSsl, DBCharset)
	}
	return dsn
}

// Builds the DSN using configuration data
func buildDSNFromConfig(connection map[string]interface{}) string {
	return buildDSN(
		connection["server"].(string),
		connection["username"].(string),
		connection["password"].(string),
		connection["database"].(string),
		connection["ssl"].(string),
		connection["charset"].(string),
	)
}

// Initializes the database connections and bootstraps the main configuration
func Boot() {
	// Retrieve all connections from the configuration
	connections, ok := foundation.App.Config.Get("database.connections", nil).(map[string]interface{})
	if !ok || connections == nil {
		color.Redln(ErrNoConnections)
		os.Exit(1)
	}

	// Retrieve the default connection name
	defaultNameConnection, ok := foundation.App.Config.Get("database.default", "").(string)
	if !ok || defaultNameConnection == "" {
		color.Redln(ErrNoDefaultConnection)
		os.Exit(1)
	}

	// Retrieve the configuration for the default connection
	connectionMain, ok := connections[defaultNameConnection].(map[string]interface{})
	if !ok {
		color.Redln(ErrDefaultConnectionNotFound)
		os.Exit(1)
	}

	// Build the DSN for the default connection
	dsnMain := buildDSNFromConfig(connectionMain)

	// Determine the logging level based on the app's debug mode
	var debug bool
	if isDebug, ok := foundation.App.Config.Get("app.debug", true).(bool); ok && isDebug {
		debug = isDebug
	}
	var logLevel logger.LogLevel
	if debug {
		logLevel = logger.Info // Debug mode: log info
	} else {
		logLevel = logger.Silent // Production mode: log only errors
	}

	// Open the main database connection
	db, err := gorm.Open(sqlserver.Open(dsnMain), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table names
			NoLowerCase:   true, // Keep case-sensitive table names
		},
	})
	if err != nil {
		color.Redf("%s: %s: %v\n", ErrConnectionFailed, defaultNameConnection, err)
		os.Exit(1)
	}

	// Configure the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		color.Redf("%s: %v\n", ErrFailedToGetSQLDB, err)
		os.Exit(1)
	}

	poolConfig, ok := foundation.App.Config.Get("database.pool", nil).(map[string]interface{})
	if ok {
		// Set maximum idle connections
		if maxIdleConns, exists := poolConfig["max_idle_conns"].(int); exists {
			sqlDB.SetMaxIdleConns(maxIdleConns)
		}
		// Set maximum open connections
		if maxOpenConns, exists := poolConfig["max_open_conns"].(int); exists {
			sqlDB.SetMaxOpenConns(maxOpenConns)
		}
		// Set maximum idle time for connections
		if connMaxIdleTime, exists := poolConfig["conn_max_idletime"].(int); exists {
			sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
		}
		// Set maximum lifetime for connections
		if connMaxLifetime, exists := poolConfig["conn_max_lifetime"].(int); exists {
			sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
		}
	}

	// Initialize the DB resolver for multi-database configurations
	var resolver *dbresolver.DBResolver

	for nameConnection, connection := range connections {
		connectionMap, ok := connection.(map[string]interface{})
		if !ok {
			color.Redf("%s: %s\n", ErrBindingConnection, nameConnection)
			os.Exit(1)
		}

		// Register additional connections (non-default)
		if nameConnection != defaultNameConnection {
			dsn := buildDSNFromConfig(connectionMap)
			datas := connectionMap["datas"].([]interface{})
			datas = append(datas, nameConnection)

			// Initialize resolver if it hasn't been created yet
			if resolver == nil {
				resolver = dbresolver.Register(dbresolver.Config{
					Sources:           []gorm.Dialector{sqlserver.Open(dsn)},
					Policy:            dbresolver.RandomPolicy{},
					TraceResolverMode: debug,
				}, datas...)
			} else {
				// Register additional configurations into the resolver
				resolver = resolver.Register(dbresolver.Config{
					Sources:           []gorm.Dialector{sqlserver.Open(dsn)},
					Policy:            dbresolver.RandomPolicy{},
					TraceResolverMode: debug,
				}, datas...)
			}
		}
	}

	// Apply resolver if defined
	if resolver != nil {
		if err := db.Use(resolver); err != nil {
			color.Redf("Error registering connections: %v\n", err)
			os.Exit(1)
		} else {
			color.Greenf("Connections registered successfully.\n")
		}
	} else {
		color.Yellowf("No connections to register.\n")
	}

	// Assign the initialized database to the global `DB` variable
	DB = Database{Ctx: db}
}
