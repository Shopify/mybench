package mybench

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/client"
)

// The database config object that can be turned into a single connection
// (without connection pooling).
type DatabaseConfig struct {
	// TODO: Unix socket
	Host     string
	Port     int
	User     string
	Pass     string
	Database string

	// If this is set, a connection will not be established. This is useful for
	// non-database-related tests such as selfbench.
	// TODO: this is kind of a hack...
	NoConnection bool
}

// Creates a new database if it doesn't exist
func (c DatabaseConfig) CreateDatabaseIfNeeded() error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := client.Connect(addr, c.User, c.Pass, "")
	if err == nil {
		_, err = conn.Execute(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.Database))
		conn.Close()
	}

	return err
}

// Returns a connection object based on the database configuration
func (c DatabaseConfig) Connection() (*Connection, error) {
	if c.NoConnection {
		return &Connection{}, nil
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conn, err := client.Connect(addr, c.User, c.Pass, c.Database)
	if err != nil {
		return nil, err
	}

	return &Connection{
		Conn: conn,
	}, nil
}

type Connection struct {
	*client.Conn
}
