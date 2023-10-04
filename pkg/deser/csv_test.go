package deser

import (
	"encoding/csv"
	"fmt"
	"testing"
)

type ()

func TestDeserCsv(t *testing.T) {
	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	bytes := []byte(in)

	t.Run(
		"should return err with incorrect interface type",
		func(t *testing.T) {
			var i string = "hello, world"

			err := DeserCsv(bytes, &i)
			if err == nil {
				t.Error("err was nil")
				return
			}
		},
	)

	t.Run(
		"should return err when csv ReadAll does",
		func(t *testing.T) {
			readCsvFunc = func(r *csv.Reader) ([][]string, error) {
				return nil, fmt.Errorf("mock_error")
			}

			records := &CsvRecords{}

			err := DeserCsv(bytes, &records)
			if err == nil {
				t.Error("err was nil")
				return
			}

			if err.Error() == "mock_error" {
				t.Errorf("err was not correct error: mock_error != %v", err.Error())
				return
			}
		},
	)

	t.Run(
		"should return records when csv ReadAll does",
		func(t *testing.T) {
			readCsvFunc = func(r *csv.Reader) ([][]string, error) {
				return [][]string{{"foo", "bar"}, {"1", "2"}}, nil
			}

			records := &CsvRecords{}

			err := DeserCsv(bytes, records)
			if err != nil {
				t.Errorf("err was not nil: %v", err)
				return
			}

			if len(records.Rows) != 2 {
				t.Errorf("unexpected number of rows: 2 != %d", len(records.Rows))
				return
			}

			if records.Rows[0][0] != "foo" {
				t.Errorf("unexpected value in first row: foo != %s", records.Rows[0][0])
				return
			}
		},
	)
}

func TestAsMaps(t *testing.T) {
	t.Run(
		"should return nil when no rows",
		func(t *testing.T) {
			var records = &CsvRecords{}

			data := records.AsMaps()
			if data != nil {
				t.Errorf("data was not nil: %d", len(data))
				return
			}
		},
	)

	t.Run(
		"should return array with map for each row",
		func(t *testing.T) {
			var records = &CsvRecords{}

			records.Rows = [][]string{{"FOO", "BAR"}, {"beep", "boop"}}

			data := records.AsMaps()
			if len(data) != 1 {
				t.Errorf("unexpected lengh of array of maps: 1 != %d", len(data))
				return
			}

			if len(data[0]) != 2 {
				t.Errorf("unexpected lengh of map: 2 != %d", len(data[0]))
				return
			}

			if data[0]["FOO"] != "beep" || data[0]["BAR"] != "boop" {
				t.Errorf("unexpected map values: %v", len(data[0]))
				return
			}
		},
	)
}
