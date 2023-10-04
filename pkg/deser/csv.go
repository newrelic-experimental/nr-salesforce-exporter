package deser

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

var (
	// allow tests to mock
	readCsvFunc = readCsv
)

type CsvRecords struct {
	Rows [][]string
}

// CSV deserializer
func DeserCsv(data []byte, v interface{}) error {
	records, ok := v.(*CsvRecords)
	if !ok {
		return fmt.Errorf("csv can only be deserialized to [][]string")
	}

	r := csv.NewReader(bytes.NewReader(data))

	rows, err := readCsvFunc(r)
	if err != nil {
		return err
	}

	records.Rows = rows

	return nil
}

func (c *CsvRecords) AsMaps() []map[string]string {
	if len(c.Rows) <= 1 {
		return nil
	}

	m := make([]map[string]string, len(c.Rows)-1)

	headers := c.Rows[0]
	count := len(headers)

	for index, row := range c.Rows[1:] {
		m[index] = make(map[string]string, count)
		rowCount := len(row)

		for jindex, header := range headers {
			if jindex >= rowCount {
				break
			}

			m[index][header] = row[jindex]
		}
	}

	return m
}

func readCsv(r *csv.Reader) ([][]string, error) {
	return r.ReadAll()
}
