package neohttp

import (
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
)

type result struct {
	res *response
	index int
	columnMapping map[string]int

	parseStarted bool
	readingRows bool
	consumed    bool

	Columns []string `json:"columns"`
	Stats   stats `json:"stats"`
}

func (r *result) Index() int {
	return r.index
}

func (r *result) Rows(job func(row cypher.Row) (bool, error)) error {
	for {
		debugLog("getting next row")
		nextRow, err := r.nextRow()
		if err != nil {
			return errMsg(err, "failed to get the next row")
		}
		debugLog("row has been read - row is nil? (%v)", nextRow == nil)
		if nextRow == nil {
			return nil
		}
		resume, err := job(nextRow)
		if err != nil {
			return errMsg(err, "failed during Results row job")
		}
		if !resume {
			_, err := r.Consume()
			if err != nil {
				return err
			}
		}
	}
}

func (r *result) Consume() (cypher.Stats, error) {
	debugLog("consuming the remaining rows")
	for {
		nextRow, err := r.nextRow()
		if err != nil {
			return nil, errMsg(err, "failed to get the next row")
		}
		if nextRow == nil {
			return &r.Stats, nil
		}
	}
}

func (r *result) parseKeys() error {
	if r.consumed {
		return nil
	}
	if !r.parseStarted {
		t, err := r.res.dec.Token()
		if err != nil {
			return err
		}
		debugLog("decoding opening brace for result: token(%v)", t)
	}
	r.parseStarted = true
	for r.res.dec.More() {
		t, err := r.res.dec.Token()
		if err != nil {
			return err
		}
		debugLog("reading next token, expecting valid result key: token(%v)", t)
		switch t {
		case "columns":
			err = r.res.dec.Decode(&r.Columns)
			if err == nil {
				r.columnMapping = make(map[string]int, len(r.Columns))
				for index, column := range r.Columns {
					r.columnMapping[column] = index
				}
			}
		case "data":
			r.readingRows = true
			t, err = r.res.dec.Token()
			if err != nil {
				return err
			}
			debugLog("reading next token, expecting '[': token(%v)", t)
			return nil
		case "stats":
			err = r.res.dec.Decode(&r.Stats)
		default:
			return errors.Errorf("found unexpected token: %v", t)
		}
		if err != nil {
			return errors.WithMessage(err, "failed to read key "+t.(string))
		}
	}
	r.consumed = true
	t, err := r.res.dec.Token()
	if err != nil {
		return err
	}
	debugLog("expected end of result at index %v '}': token(%v)", r.index, t)
	return nil
}

// Parse the next row of the response.
// If the end of the list of rows is reached:
//   parse the remaining keys of the result, filling in 'Stats'
//   return nil for the next row.
func (r *result) nextRow() (cypher.Row, error) {
	if r.consumed {
		return nil, nil
	}
	if !r.res.dec.More() {
		t, err := r.res.dec.Token()
		if err != nil {
			return nil, err
		}
		debugLog("expecting end of list of rows ']': token(%v)", t)
		return nil, r.parseKeys()
	}
	nextRow := &row{columns: r.columnMapping}
	err := r.res.dec.Decode(&nextRow)
	if err != nil {
		return nil, err
	}
	return nextRow, nil
}
