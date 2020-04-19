package cypher

import (
	"encoding/json"
	"github.com/pkg/errors"
)

var drivers = make(map[string]Driver)
var Debug = false

// Register a neo4j driver as some name.
// This method panics on failure.
func Register(name string, driver Driver) {
	if driver == nil {
		panic("driver cannot be nil: " + name)
	}
	if _, ok := drivers[name]; ok {
		panic("driver is already defined: " + name)
	}
	drivers[name] = driver
}

// Unregister a driver. This method will not panic.
func Unregister(name string) {
	delete(drivers, name)
}

// Connect to a database using a particular driver.
func Connect(driverName, uri, dbName, username, password string) (DB, error) {
	if d, ok := drivers[driverName]; ok {
		return d.Connect(uri, dbName, username, password)
	} else {
		return nil, errors.New("cypher: driver is not defined: " + driverName)
	}
}

// Collect all rows of a result into a slice, consuming the result in the process.
func Collect(result Result) ([]Row, error) {
	rows := make([]Row, 0, 30)
	for result.NextRow() {
		rows = append(rows, result.GetRow())
	}
	return rows, errors.WithMessage(result.Err(), "failed to Collect rows")
}

// Get the first row of a result, consuming the remainder of the result. Returns nil if there were no rows.
func Single(result Result) (Row, error) {
	if !result.NextRow() {
		return nil, result.Err()
	}
	row := result.GetRow()
	_, err := result.Consume()
	return row, err
}

// Collect all rows of the result and unmarshal them into the given slice of structs.
func CollectUnmarshal(result Result, asStructSlice interface{}) error {
	rows, err := Collect(result)
	if err != nil {
		return err
	}
	return UnmarshalRows(rows, asStructSlice)
}

// Get the single row from the result, unmarshaling it into the given struct.
// If no row was found, the struct will be unmodified.
func SingleUnmarshal(result Result, asStruct interface{}) error {
	row, err := Single(result)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	return UnmarshalRow(row, asStruct)
}

// Unmarshal a row into a given struct type.
// Fields will be unmarshalled with the names of columns from the row.
func UnmarshalRow(row Row, asStruct interface{}) error {
	b, err := json.Marshal(row)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal row into json")
	}
	return errors.WithMessage(json.Unmarshal(b, &asStruct), "failed to unmarshal row into struct")
}

// Unmarshall a list of rows into a given list of structs.
// Fields will be unmarshalled with the names of columns from the row.
func UnmarshalRows(rows []Row, asStructSlice interface{}) error {
	b, err := json.Marshal(rows)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal rows into json")
	}
	return errors.WithMessage(json.Unmarshal(b, &asStructSlice), "failed to unmarshal rows into struct slice")
}

type Driver interface {
	// Connect to a database.
	Connect(uri, dbName, username, password string) (DB, error)
}

type DB interface {
	// The database is capable of running automatic-committing transactions.
	Runner

	// Returns a query runner which runs all queries in a single transaction.
	TX() (Transaction, error)

	// Run the given function, returning the result.
	// All queries which are run inside the provided QueryRunner will be run in the same transaction.
	TXJob(func(runner Transaction) (interface{}, error)) (interface{}, error)

	// Close the driver.
	Close() error
}

type Runner interface {
	// Returns the result of running the given query.
	// Errors are deferred to the response object.
	Run(cypher string, params map[string]interface{}) Result

	// Returns the summary of the result while discarding the records.
	// Errors are deferred to the response object.
	RunMany(cypherOrParams ...interface{}) Response
}

type Transaction interface {
	// Any transaction is capable of making queries.
	Runner

	// Close the transaction, keeping all changes made.
	Commit() error

	// Close the transaction, undoing all changes made.
	Rollback() error
}

type Response interface {
	// Returns true if another result has been read, false otherwise.
	// The error should be checked after false is returned.
	NextResult() bool

	// Returns the last read result. NextResult() must be called before using this function.
	GetResult() Result

	// Returns an error if an error occurred while reading the last result.
	Err() error

	// Dispose of all results in the output.
	Consume() error
}

type Result interface {
	// The index of this result in the list of results returned from the server.
	Index() int

	// Returns true if another row has been read, false otherwise.
	// The error should be checked after false is returned.
	NextRow() bool

	// Returns the last read row.
	GetRow() Row

	// Returns an error if an error occurred while reading the last row.
	Err() error

	// Discard all of the rows and get the stats.
	Consume() (Stats, error)
}

type Row interface {
	// Should marshal into an object which has the column names as keys and the values from the row as values.
	json.Marshaler

	// Get a single column by index.
	GetAt(i int) interface{}

	// Get a single column by name.
	Get(n string) interface{}
}

type Stats interface {
	ConstraintsAdded() int
	ConstraintsRemoved() int
	ContainsUpdates() bool
	IndexesAdded() int
	IndexesRemoved() int
	LabelsAdded() int
	LabelsRemoved() int
	NodesCreated() int
	NodesDeleted() int
	PropertiesSet() int
	RelationshipDeleted() int
	RelationshipsCreated() int
	ContainsSystemUpdates() bool
	SystemUpdates() int
}
