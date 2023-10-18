package pool

import "errors"

var (
	ErrDriverClosed = errors.New("pool: driver is closed")

	ErrDupClose = errors.New("pool: duplicate driverConn close")

	// ErrConnDone is returned by any operation that is performed on a connection
	// that has already been returned to the connection pool.
	ErrConnDone = errors.New("pool: connection is already closed")
)

const (
	// This is the size of the connectionOpener request chan (DB.openerCh).
	// This value should be larger than the maximum typical value
	// used for db.maxOpen. If maxOpen is significantly larger than
	// connectionRequestQueueSize then it is possible for ALL calls into the *DB
	// to block until the connectionOpener can satisfy the backlog of requests.
	connectionRequestQueueSize = 1000000

	// This define the number of how many idle connection should be keep in pool
	defaultMaxIdleConns = 5

	// maxBadConnRetries is the number of maximum retries if the driver returns
	// driver.ErrBadConn to signal a broken connection before forcing a new
	// connection to be opened.
	maxBadConnRetries = 2
)

// connReuseStrategy determines how (*DB).conn returns database connections.
type connReuseStrategy uint8

const (
	// alwaysNewConn forces a new connection to the database.
	alwaysNewConn connReuseStrategy = iota
	// cachedOrNewConn returns a cached connection, if available, else waits
	// for one to become available (if MaxOpenConns has been reached) or
	// creates a new database connection.
	cachedOrNewConn
)
