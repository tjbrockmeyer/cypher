package neohttp

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
	"io"
	"net/http"
	"strings"
)

type response struct {
	deferredErr error
	dec         *json.Decoder
	resBody     io.Closer

	parseStarted   bool
	readingResults bool
	consumed       bool
	singleResult   bool

	resultCount int
	lastResult  *result

	statusCode int
	header     http.Header
	errors     []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	commit      string
	transaction *struct {
		Expires string `json:"expires"`
	}
}

func (r *response) NextResult() bool {
	if r.deferredErr != nil || r.consumed {
		return false
	}
	err := r.nextResult()
	if err != nil {
		r.deferredErr = errMsg(err, "failed to get the next result")
		return false
	}
	if r.lastResult == nil {
		r.deferredErr = r.Consume()
		return false
	}
	return true
}

func (r *response) GetResult() cypher.Result {
	return r.lastResult
}

func (r *response) Err() error {
	return r.deferredErr
}

func (r *response) Consume() error {
	if r.deferredErr != nil {
		return r.deferredErr
	}
	for {
		err := r.nextResult()
		if err != nil {
			return errMsg(err, "failed to get the next result")
		}
		if r.lastResult == nil {
			debugLog("response has been completely consumed")
			r.consumed = true
			if err := r.getErrors(); err != nil {
				debugLog("response errors found: %v", err)
				return errMsg(err, "database returned errors")
			}
			return errMsg(r.resBody.Close(), "failed to close the response body")
		}
		if _, err = r.lastResult.Consume(); err != nil {
			return err
		}
	}
}

func (r *response) parseKeys() error {
	if !r.parseStarted {
		t, err := r.dec.Token()
		if err != nil {
			return err
		}
		debugLog("decoding opening brace for response: token(%v)", t)
	}
	r.parseStarted = true
	for r.dec.More() {
		t, err := r.dec.Token()
		if err != nil {
			return err
		}
		debugLog("reading next token, expecting valid response key: token(%v)", t)
		switch t {
		case "results":
			r.readingResults = true
			t, err = r.dec.Token()
			if err != nil {
				return err
			}
			debugLog("reading next token, expecting '[': token(%v)", t)
			return nil
		case "errors":
			err = r.dec.Decode(&r.errors)
		case "commit":
			err = r.dec.Decode(&r.commit)
		case "transaction":
			err = r.dec.Decode(&r.transaction)
		default:
			return errors.New("invalid token found: " + fmt.Sprint(t))
		}
		if err != nil {
			return errors.WithMessage(err, "failed to read key "+t.(string))
		}
	}
	r.consumed = true
	t, err := r.dec.Token()
	if err != nil {
		return err
	}
	debugLog("expected end of response '}': token(%v)", t)
	return nil
}

// Read the next result.
// If there are no more results, the remaining keys will be read and processed.
// The transaction key will be non-nil if found.
func (r *response) nextResult() error {
	if r.consumed {
		return nil
	}
	if r.lastResult != nil && !r.lastResult.consumed {
		debugLog("last result was not consumed - calling lastResult.Consume() automatically")
		_, err := r.lastResult.Consume()
		if err != nil {
			return err
		}
	}
	if !r.dec.More() {
		r.lastResult = nil
		t, err := r.dec.Token()
		if err != nil {
			return err
		}
		debugLog("expecting end of list of results ']': token(%v)", t)
		return r.parseKeys()
	}
	r.lastResult = &result{
		res:   r,
		index: r.resultCount,
	}
	r.resultCount++
	err := r.lastResult.parseKeys()
	if err != nil {
		return err
	}
	return nil
}

// Returns any attached errors found as an error.
func (r *response) getErrors() error {
	debugLog("looking for response errors")
	if r.errors != nil && len(r.errors) > 0 {
		b := strings.Builder{}
		for _, err := range r.errors {
			b.WriteString(err.Code + ": " + err.Message)
		}
		return errors.New(b.String())
	}
	return nil
}
