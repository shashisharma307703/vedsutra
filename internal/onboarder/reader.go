package onboarder

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocarina/gocsv"
)

// DataReader reads rows of a specific type from a file.
type DataReader interface {
	ReadRows(dest interface{}) error
}

type CSVReader struct {
	filePath string
}

func NewCSVReader(filePath string) *CSVReader {
	return &CSVReader{filePath: filePath}
}

func (r *CSVReader) ReadRows(dest interface{}) error {
	f, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return gocsv.UnmarshalFile(f, dest)
}

type JSONReader struct {
	filePath string
}

func NewJSONReader(filePath string) *JSONReader {
	return &JSONReader{filePath: filePath}
}

func (r *JSONReader) ReadRows(dest interface{}) error {
	f, err := os.Open(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// DetectReader chooses CSV or JSON based on file extension.
func DetectReader(filePath string) (DataReader, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return NewCSVReader(filePath), nil
	case ".json":
		return NewJSONReader(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// ReadRowsAsMaps reads CSV/JSON rows into a slice of map[string]interface{}.
func ReadRowsAsMaps(reader DataReader) ([]map[string]interface{}, error) {
	if csvReader, ok := reader.(*CSVReader); ok {
		f, err := os.Open(csvReader.filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var rows []map[string]string
		if err := gocsv.UnmarshalFile(f, &rows); err != nil {
			return nil, err
		}
		result := make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			m := make(map[string]interface{})
			for k, v := range row {
				m[k] = v
			}
			result[i] = m
		}
		return result, nil
	} else if jsonReader, ok := reader.(*JSONReader); ok {
		f, err := os.Open(jsonReader.filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var rows []map[string]interface{}
		if err := json.NewDecoder(f).Decode(&rows); err != nil {
			return nil, err
		}
		return rows, nil
	}
	return nil, fmt.Errorf("unsupported reader type")
}