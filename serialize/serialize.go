// Serializes data to a custom tsv format
// Each 'type' has to have a type name on a line by itself
// Next, each type has a list of column names (Tab separated)
// Next, 1 or more rows of values (with the same number of columns as above)
// If there are more types, an empty line before starting the next type
package serialize

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Serializer struct {
	w                   io.Writer
	err                 error // A remembered error, to be retrieved at the writer's leisure
	columnCount         int   // Row count for current data type. Not allowed to go to a new type with zero.
	state               int   // 0 == initial, 1 == type name set, 2 == schema set, 3 == writing a row, 4 == finished, any other == err
	currentTypeRowCount int
	currentDataRow      *Row
}

type Row struct {
	s     *Serializer
	count int
}

// Can write directly to any writer. If you want to make sure there are no writers before committing to a response, maybe put it into some buffer and check error first.
func NewWriter(w io.Writer) *Serializer {
	return &Serializer{w: w}
}

// Check if there have been any errors.
func (s *Serializer) Err() error {
	return s.err
}

func (s *Serializer) Finish() {
	if s.state != 2 {
		s.maybeSetError(errors.New("can't finish without being done with a data type"))
		return
	}
	if s.columnCount == 0 || s.currentTypeRowCount == 0 {
		s.maybeSetError(errors.New("can't finish without writing to the previous data type"))
		return
	}
	s.state = 4
	s.writeUnsafeString("\n")
}

func (s *Serializer) NextDataType(t string) {
	if s.Err() != nil {
		return
	}
	if s.state != 0 && s.state != 2 {
		s.maybeSetError(errors.New("Not in a state to write a new data type: " + t))
		return
	}
	if s.state == 2 && (s.columnCount == 0 || s.currentTypeRowCount == 0) {
		s.maybeSetError(errors.New("Can't start a new type without writing any rows of the previous type." + t))
		return
	}

	if s.state == 2 {
		s.writeUnsafeString("\n")
	}
	s.writeEscapedString(t)
	s.writeUnsafeString("\n")
	s.state = 1
}

func (s *Serializer) WriteColumnNames(cols []string) {
	if s.state != 1 {
		s.maybeSetError(fmt.Errorf("not in the right state to start writing colum names: %+v", cols))
		return
	}
	if len(cols) == 0 {
		s.maybeSetError(errors.New("tried to write 0 column names"))
		return
	}
	for i, v := range cols {
		if i > 0 {
			s.writeUnsafeString("\t")
		}
		s.writeEscapedString(v)
	}
	s.writeUnsafeString("\n")
	s.state = 2
	s.columnCount = len(cols)
}

func (s *Serializer) AddRow() *Row {
	if s.state != 2 {
		s.maybeSetError(errors.New("tried to add a row, but we aren't in the right state"))
		return &Row{s: s} // This returned row can't write, because it is never the currentDataRow
	}
	if s.currentDataRow != nil {
		s.maybeSetError(errors.New("tried to start a new row, but we have an open row"))
		return &Row{s: s} // This returned row can't write, because it is never the currentDataRow
	}
	s.currentTypeRowCount++
	s.state = 3
	r := &Row{s: s}
	s.currentDataRow = r
	return r
}

// Writes a single boolean value as 'true' or 'false'
func (r *Row) WriteBoolValue(b bool) *Row {
	r.WriteStringValue(strconv.FormatBool(b))
	return r
}

func (r *Row) WriteIntValue(i int) *Row {
	r.WriteStringValue(strconv.FormatInt(int64(i), 10))
	return r
}

// Writes a single column of a row.
func (r *Row) WriteStringValue(val string) *Row {
	if r != r.s.currentDataRow {
		r.s.maybeSetError(errors.New("this is not the current data row"))
		return r
	}
	if r.count >= r.s.columnCount {
		r.s.maybeSetError(fmt.Errorf("tried to write too many columns when writing %s", val))
		return r
	}
	if r.count > 0 {
		r.s.writeUnsafeString("\t")
	}
	r.count++
	r.s.writeEscapedString(val)
	return r
}

func (r *Row) Done() {
	if r != r.s.currentDataRow {
		r.s.maybeSetError(errors.New("this is not the current data row"))
		return
	}
	if r.count != r.s.columnCount {
		r.s.maybeSetError(fmt.Errorf("did not write enough columns, got: %d, want: %d", r.count, r.s.columnCount))
		return
	}
	r.s.state = 2
	r.s.currentDataRow = nil
	r.s.writeUnsafeString("\n")
}

func (s *Serializer) writeEscapedString(orig string) {
	tabsReplaced := strings.ReplaceAll(orig, "\t", "#")
	newlinesReplaced := strings.ReplaceAll(tabsReplaced, "\n", "@")
	nullsReplaced := strings.ReplaceAll(newlinesReplaced, "\x00", "!")
	s.writeUnsafeString(nullsReplaced)
}

func (s *Serializer) writeUnsafeString(raw string) {
	if s.Err() != nil {
		return
	}
	_, err := s.w.Write([]byte(raw))
	s.maybeSetError(err)
}

func (s *Serializer) maybeSetError(err error) {
	// sets only the first error
	if err != nil && s.err == nil {
		s.err = err
		s.state = -1
	}
}
