package mybench

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/client"
)

// The database config object that can be turned into a single connection
// (without connection pooling).
type DatabaseConfig struct {
	// TODO: Unix socket
	Host     string `short:"H" long:"host" description:"host of the database instance" group:"database"`
	Port     int    `short:"P" long:"port" description:"port of the database instance" default:"3306" group:"database"`
	User     string `short:"u" long:"user" description:"username for the database" default:"root" group:"database"`
	Pass     string `short:"p" long:"pass" description:"password for the database" default:"" group:"database"`
	Database string `short:"d" long:"db" description:"name of the database schema (if applicable)" default:"mybench" group:"database"`

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

func (c *Connection) Close() error {
	if c.Conn == nil {
		return nil // This happens if NoConnection is true
	}

	return c.Conn.Close()
}
