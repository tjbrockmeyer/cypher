package neohttp

import (
	"github.com/pkg/errors"
	"github.com/tjbrockmeyer/cypher"
)

type result struct {
	res           *response
	index         int
	columnMapping map[string]int
	deferredErr   error

	parseStarted bool
	readingRows  bool
	consumed     bool
	lastRow      cypher.Row

	Columns []string `json:"columns"`
	Stats   stats    `json:"stats"`
}

func (r *result) Index() int {
	return r.index
}

func (r *result) NextRow() bool {
	if r.deferredErr != nil || r.consumed {
		return r.nextRowDone()
	}
	err := r.nextRow()
	if err != nil {
		r.deferredErr = errMsg(err, "failed to get the next row")
		return r.nextRowDone()
	}
	if r.lastRow == nil {
		return r.nextRowDone()
	}
	return true
}

func (r *result) GetRow() cypher.Row {
	return r.lastRow
}

func (r *result) Err() error {
	return r.deferredErr
}

func (r *result) Consume() (cypher.Stats, error) {
	for {
		err := r.nextRow()
		if err != nil {
			return nil, errMsg(err, "failed to get the next row")
		}
		if r.lastRow == nil {
			return &r.Stats, nil
		}
	}
}

func (r *result) parseKeys() error {
	if r.consumed {
		return nil
	}
	if !r.parseStarted {
		_, err := r.res.dec.Token()
		if err != nil {
			return err
		}
	}
	r.parseStarted = true
	for r.res.dec.More() {
		t, err := r.res.dec.Token()
		if err != nil {
			return err
		}
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
	_, err := r.res.dec.Token()
	if err != nil {
		return err
	}
	return nil
}

// Parse the next row of the response.
// If the end of the list of rows is reached:
//   parse the remaining keys of the result, filling in 'Stats'
//   return nil for the next row.
func (r *result) nextRow() error {
	if r.consumed {
		return nil
	}
	if !r.res.dec.More() {
		r.lastRow = nil
		_, err := r.res.dec.Token()
		if err != nil {
			return err
		}
		return r.parseKeys()
	}
	r.lastRow = &row{columns: r.columnMapping}
	err := r.res.dec.Decode(&r.lastRow)
	if err != nil {
		return err
	}
	return nil
}

func (r *result) nextRowDone() bool {
	if r.res.singleResult {
		r.deferredErr = r.res.Consume()
	}
	return false
}
