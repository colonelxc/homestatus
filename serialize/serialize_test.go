// Serializes data to a custom tsv format

// Each 'type' has to have a type name on a line by itself
// Next, each type has a list of column names (Tab separated)
// Next, 1 or more rows of values (with the same number of columns as above)
// If there are more types, or we are completely done, add an empty line

package serialize

import (
	"bytes"
	"testing"
)

func TestNewWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	s := NewWriter(buf)

	s.NextDataType("Test")
	s.WriteColumnNames([]string{"A", "B", "C"})
	s.AddRow().WriteBoolValue(true).WriteIntValue(1).WriteStringValue("abc").Done()
	r := s.AddRow()
	r.WriteBoolValue(false)
	r.WriteIntValue(2)
	r.WriteStringValue("def")
	r.Done()
	s.NextDataType("Second")
	s.WriteColumnNames([]string{"X"})
	r = s.AddRow()
	r.WriteStringValue("Y")
	r.Done()
	s.Finish()

	if s.Err() != nil {
		t.Errorf("Err() = %v", s.Err())
	}
	expected := "Test\nA\tB\tC\ntrue\t1\tabc\nfalse\t2\tdef\n\nSecond\nX\nY\n\n"
	if out := buf.String(); out != expected {
		t.Errorf("Invalid serialization, got\n%s\nwant\n%s", out, expected)
	}
}

func TestSerializer_Err(t *testing.T) {
	tests := []struct {
		name string
		f    func(*Serializer)
	}{
		{
			"NextDataType two times in a row",
			func(s *Serializer) {
				s.NextDataType("a")
				s.NextDataType("b")
			},
		},
		{
			"Zero columns",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{})
			},
		},
		{
			"No Data columns",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"c"})
				s.NextDataType("b")
			},
		},
		{
			"Too few data columns",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).Done()
			},
		},
		{
			"Too many data columns",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).WriteBoolValue(false).WriteBoolValue(true).Done()
			},
		},
		{
			"Not calling Done on a Row, before the next row",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).WriteBoolValue(false)
				s.AddRow()
			},
		},
		{
			"Not calling Done on a Row, before the next data type",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).WriteBoolValue(false)
				s.NextDataType("b")
			},
		},
		{
			"Calling finish twice",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).WriteBoolValue(false).Done()
				s.Finish()
				s.Finish()
			},
		},
		{
			"Writing column names twice",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.WriteColumnNames([]string{"d", "e"})
			},
		},
		{
			"Writing column names without a data type",
			func(s *Serializer) {
				s.WriteColumnNames([]string{"b", "c"})
			},
		},
		{
			"Writing column names without a data type, for the second type",
			func(s *Serializer) {
				s.NextDataType("a")
				s.WriteColumnNames([]string{"b", "c"})
				s.AddRow().WriteBoolValue(true).WriteBoolValue(false).Done()
				s.WriteColumnNames([]string{"b", "c"})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			s := NewWriter(b)
			tt.f(s)
			if err := s.Err(); err == nil {
				t.Errorf("Serializer.Err() error = nil, wanted error")
			}
		})
	}
}
