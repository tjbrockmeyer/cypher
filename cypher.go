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
	_, err := result.Rows(func(row Row) (interface{}, error) {
		rows = append(rows, row)
		return true, nil
	})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to Collect rows")
	}
	return rows, nil
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

// Collect all rows and unmarshal them all into a slice of structs, consuming the result in the process.
// Fields will be unmarshalled with the names of columns from the row.
func CollectAndUnmarshal(result Result, structSlice interface{}) error {
	rows, err := Collect(result)
	if err != nil {
		return err
	}
	b, err := json.Marshal(rows)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal rows into json")
	}
	return errors.WithMessage(json.Unmarshal(b, &structSlice), "failed to unmarshal rows into struct slice")
}

// Get the first row from the first result, discarding all remaining rows and stats.
func GetSingleResultRow(response Response) (Row, error) {
	row, err := response.Results(func(result Result) (interface{}, error) {
		return result.Rows(func(row Row) (interface{}, error) {
			return row, nil
		})
	})
	if err != nil {
		return nil, err
	}
	return row.(Row), nil
}

// Consume a the response, returning the stats from the first result.
func ConsumeSingleResult(response Response) (Stats, error) {
	stats, err := response.Results(func(result Result) (interface{}, error) {
		return result.Consume()
	})
	if err != nil {
		return nil, err
	}
	return stats.(Stats), nil
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
	Run(cypher string, params map[string]interface{}) Response

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
	// Run the job on each result in the output.
	// Return false to stop processing after the current job is complete, true otherwise.
	// Return an error from the job to have it propogate out.
	Results(job func(result Result) (interface{}, error)) (interface{}, error)

	// Dispose of all results in the output.
	Consume() error
}

type Result interface {
	// The index of this result in the list of results returned from the server.
	Index() int

	// Run the job on each row of the output.
	// Return nil from the job to continue, anything else to stop iterating and discard the remainder of the results.
	// Return an error from the job to have it propogate out.
	Rows(job func(row Row) (interface{}, error)) (interface{}, error)

	// Discard all of the results and get the stats.
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
