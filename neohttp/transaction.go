package neohttp

import (
	"github.com/tjbrockmeyer/cypher"
	"strings"
)

type transaction struct {
	db       *database
	id       string
	location string
	alive    bool
}

func (tx *transaction) Run(statement string, params map[string]interface{}) cypher.Result {
	return tx.db.run(tx.id, statement, params)
}

func (tx *transaction) RunMany(cypherOrParams ...interface{}) cypher.Response {
	return tx.db.runMany(tx.id, cypherOrParams...)
}

func (tx *transaction) Commit() error {
	res := tx.db.getResponse("POST", tx.id+"/commit", request{Statements: []query{}})
	if res.deferredErr != nil {
		return errMsg(res.deferredErr, "error during commit request")
	}
	if err := res.Consume(); err != nil {
		tx.alive = false
		return err
	}
	if err := tx.handleResponse(res); err != nil {
		return errMsg(err, "database returned errors")
	}
	return nil
}

func (tx *transaction) Rollback() error {
	res := tx.db.getResponse("DELETE", tx.id, request{Statements: []query{}})
	if err := res.Consume(); err != nil {
		tx.alive = false
		return err
	}
	if err := tx.handleResponse(res); err != nil {
		return errMsg(err, "database returned errors")
	}
	return nil
}

func (tx *transaction) handleResponse(res *response) error {
	if tx.id == "" {
		tx.location = res.header.Get("Location")
		index := strings.LastIndex(tx.location, "/")
		if index != -1 {
			tx.id = tx.location[index:]
		}
	}
	tx.alive = res.transaction != nil
	return nil
}
