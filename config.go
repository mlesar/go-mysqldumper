package mysqldumper

import (
	"encoding/json"
	"strings"
)

type Config struct {
	Tables  map[string]string `json:"tables"`
	Columns map[string]string `json:"columns"`
	Size    struct {
		Gt      *Size             `json:"gt"`
		Filters map[string]string `json:"filters"`
	} `json:"size"`
	Replacements map[string]string `json:"replacements"`
	Filters      map[string]string `json:"filters"`
}

func ParseConfig(data []byte) (config *Config, err error) {
	config = &Config{}

	if data != nil {
		err = json.Unmarshal(data, config)
	}

	return config, err
}

func (s *Config) CanDumpDefinition(tableName string) bool {
	if s.Filters == nil {
		return true
	}

	filter, ok := s.Filters[tableName]
	if !ok {
		filter, ok = s.Filters["*"]
	}

	return !ok || !(filter == TABLE_FILTER_IGNORE || filter == TABLE_FILTER_ONLYDATA)
}

func (s *Config) GetDumpColumns(table Table) string {
	columns := []string{}
	for _, columnName := range *table.Columns {
		column := "`" + columnName + "`"

		replacement, ok := s.Replacements[table.Name+"."+columnName]
		if ok {
			column = replacement + " AS " + column
		}

		columns = append(columns, column)
	}

	return strings.Join(columns, ",")
}

func (s *Config) GetDumpFilter(table Table) string {
	if s.Filters != nil {
		filter, ok := s.Filters[table.Name]
		if !ok {
			filter, ok = s.Filters["*"]
		}
		if ok && (filter == TABLE_FILTER_IGNORE || filter == TABLE_FILTER_NODATA) {
			return "LIMIT 0"
		}
	}

	// check for table filter
	if s.Tables != nil {
		tableFilter, ok := s.Tables[table.Name]
		if ok {
			return tableFilter
		}
	}

	// check for column filter
	if s.Columns != nil {
		for column, columnFilter := range s.Columns {
			if table.Columns.IndexOf(column) >= 0 {
				return columnFilter
			}
		}
	}

	// check for size filter
	if s.Size.Gt != nil && table.Size != nil && *s.Size.Gt < *table.Size {
		// check for size table filters
		tableColumnFilter, ok := s.Size.Filters[table.Name]
		if ok {
			return tableColumnFilter
		}

		// check for size table filters
		for tableColumnStr, tableColumnFilter := range s.Size.Filters {
			tableColumn := strings.Split(tableColumnStr, ".")
			if len(tableColumn) > 1 && tableColumn[0] == "*" && table.Columns.IndexOf(tableColumn[1]) >= 0 {
				return tableColumnFilter
			}
		}

		return "ORDER BY 1 DESC LIMIT 30" // TABLE IS TOO BIG
	}

	return ""
}
