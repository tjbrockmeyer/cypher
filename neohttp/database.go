package neohttp

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
	"io/ioutil"
	"net/http"
	"time"
)

type database struct {
	uri       string
	basicAuth string
	discovery struct {
		BoltDirect string `json:"bolt_direct"`
		Cluster    string `json:"cluster"`
		TX         string `json:"transaction"`
		Version    string `json:"neo4j_version"`
		Edition    string `json:"neo4j_edition"`
	}
}

func (db *database) Run(statement string, params map[string]interface{}) cypher.Result {
	return db.run("/commit", statement, params)
}

func (db *database) RunMany(cypherOrParams ...interface{}) cypher.Response {
	return db.runMany("/commit", cypherOrParams...)
}

func (db *database) TX() (cypher.Transaction, error) {
	return &transaction{
		db:    db,
		alive: true,
	}, nil
}

func (db *database) TXJob(job func(tx cypher.Transaction) (interface{}, error)) (interface{}, error) {
	tx := &transaction{db: db, alive: true}
	val, err := job(tx)
	if err != nil {
		err = tx.Rollback()
		if err != nil {
			return nil, errMsg(err, "error during rollback")
		}
		return nil, errMsg(err, "error during TX job")
	}
	return val, errMsg(tx.Commit(), "error during commit")
}

func (db *database) Close() error {
	return nil
}

func (db *database) connectWithRetry(retries int) error {
	debugLog("connecting to the database via http(s) at (%s)", db.uri)
	res, err := http.Get(db.uri)
	if err != nil {
		err = errMsg(err, "failed to get at uri ("+db.uri+")")
	} else {
		var body []byte
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			err = errMsg(err, "failed to read response body from uri ("+db.uri+")")
		} else {
			err = json.Unmarshal(body, &db.discovery)
			if err != nil {
				err = errMsg(err, "failed to unmarshal response body at uri ("+db.uri+")")
			} else {
				debugLog("successfully connected to the database at (%s)", db.uri)
				return nil
			}
		}
	}

	retries -= 1
	if retries <= 0 {
		debugLog("failed to connect to the database")
		return errMsg(err, "failed to connect - no more retries")
	}
	debugLog("failed to connect - retries remaining: %v | retrying in 3 seconds...", retries)
	<-time.After(time.Second * 3)
	return db.connectWithRetry(retries)
}

func (db *database) getResponse(method, relPath string, body interface{}) *response {
	r := new(response)
	b, err := json.Marshal(body)
	if err != nil {
		r.deferredErr = errors.WithMessage(err, "could not marshal request body")
		return r
	}
	reqBody := bytes.NewReader(b)
	req, err := http.NewRequest(method, db.discovery.TX+relPath, reqBody)
	if err != nil {
		r.deferredErr = errors.WithMessage(err, "could not create request")
		return r
	}
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Stream", "true")
	if db.basicAuth != "" {
		req.Header.Set("Authorization", db.basicAuth)
	}
	debugLog("requesting %s with payload %+v", req.URL.String(), body)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		r.deferredErr = errors.WithMessage(err, "could not send request / receive response")
		return r
	}
	r.statusCode = res.StatusCode
	r.header = res.Header
	r.dec = json.NewDecoder(res.Body)
	r.resBody = res.Body
	if cypher.Debug {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			r.deferredErr = errors.WithMessage(err, "DEBUG ERROR ioutil.ReadAll() failed to read body")
			return r
		}
		debugLog("response body: %v", string(bodyBytes))
		r.dec = json.NewDecoder(bytes.NewReader(bodyBytes))
	}
	debugLog("response received with status: %v", r.statusCode)
	err = r.parseKeys()
	if err != nil {
		r.deferredErr = err
	}
	return r
}

func (db *database) run(id, cypher string, params map[string]interface{}) cypher.Result {
	res := db.getResponse("POST", id, request{
		Statements: []query{{
			Statement:    cypher,
			Parameters:   params,
			IncludeStats: true,
		}},
	})
	res.singleResult = true
	if !res.NextResult() {
		return &result{
			res:         res,
			deferredErr: res.Err(),
		}
	}
	return res.GetResult()
}

func (db *database) runMany(id string, cypherOrParams ...interface{}) cypher.Response {
	statements := make([]query, 0, 10)
	for _, val := range cypherOrParams {
		switch v := val.(type) {
		case string:
			statements = append(statements, query{Statement: v, IncludeStats: true})
		case map[string]interface{}:
			if len(statements) == 0 {
				continue
			}
			statements[len(statements)-1].Parameters = v
		default:
			return &response{deferredErr: errors.New(
				"RunMany() accepts only string cypher statements, or map[string]interface{} parameter declarations")}
		}
	}
	return db.getResponse("POST", id, request{
		Statements: statements,
	})
}
