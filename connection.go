package mybench

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
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

	// The number of underlying connections per Connection object, implemented as a
	// static pool from which connections are fetched in a round-robin sequence with
	// each successive request. The sole purpose of this feature is to multiply the
	// number of open connections to the database to assess any performance impact
	// specific to the overall number of open database connections.
	ConnectionMultiplier int
}
type Connection struct {
	ConnList  []*client.Conn
	ConnIndex int
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
	if c.ConnectionMultiplier == 0 {
		c.ConnectionMultiplier = 1
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	connList := make([]*client.Conn, c.ConnectionMultiplier)
	for i := 0; i < c.ConnectionMultiplier; i++ {
		conn, err := client.Connect(addr, c.User, c.Pass, c.Database)
		if err != nil {
			return nil, err
		}
		connList[i] = conn
	}
	return &Connection{
		ConnList:  connList,
		ConnIndex: 0,
	}, nil
}

func (c *Connection) GetConn() *client.Conn {
	c.ConnIndex = (c.ConnIndex + 1) % len(c.ConnList)
	return c.ConnList[c.ConnIndex]
}

func (c *Connection) Close() error {
	var err error
	for _, conn := range c.ConnList {
		err = conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) Execute(query string, args ...interface{}) (*mysql.Result, error) {
	return c.GetConn().Execute(query, args...)
}

func (c *Connection) Prepare(query string) (*client.Stmt, error) {
	return c.GetConn().Prepare(query)
}
