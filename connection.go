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

	// The number of underlying connections per Connection object, implemented as a
	// static pool from which connections are fetched in a round-robin sequence with
	// each successive request. The sole purpose of this feature is to multiply the
	// number of open connections to the database to assess any performance impact
	// specific to the overall number of open database connections.
	//
	// Note: this feature does not work on all benchmarks at this moment. It only
	// works with benchmarks that uses the mybench.Connection.GetRoundRobinConnection
	// function to get their connections and execute queries.
	ConnectionMultiplier int
}

// A thin wrapper around https://pkg.go.dev/github.com/go-mysql-org/go-mysql/client#Conn
// for now. It is possible in the future to extend this to support databases
// other than MySQL.
//
// This should only be initialized via DatabaseConfig.Connection().
type Connection struct {
	*client.Conn
	connList  []*client.Conn
	connIndex int
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
		Conn:      connList[0],
		connList:  connList,
		connIndex: 0,
	}, nil
}

func (c *Connection) GetRoundRobinConnection() *client.Conn {
	c.connIndex = (c.connIndex + 1) % len(c.connList)
	return c.connList[c.connIndex]
}

func (c *Connection) Close() error {
	var err error
	for _, conn := range c.connList {
		err = conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
