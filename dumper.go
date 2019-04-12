package mysqldumper

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Dumper struct {
	db     *sql.DB
	config *Config
	logger *logrus.Logger
}

func New(config *Config, db *sql.DB, logger *logrus.Logger) *Dumper {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &Dumper{db: db, config: config, logger: logger}
}

func (s *Dumper) Dump(w DumpWriter) error {
	startQueries := `
		SET NAMES utf8;
		SET FOREIGN_KEY_CHECKS = 0;
		SET SESSION sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION';
	`
	endQueries := `
		SET FOREIGN_KEY_CHECKS = 1;
	`
	err := w.Write(startQueries)
	if err != nil {
		return err
	}

	tables, err := s.GetDefinitions()
	if err != nil {
		return err
	}

	err = s.DumpDefinitions(w, tables)
	if err != nil {
		return err
	}

	err = s.DumpData(w, tables)
	if err != nil {
		return err
	}

	return w.Write(endQueries)
}

func (s *Dumper) DumpDefinitions(w DumpWriter, tables *Tables) error {
	var drops string
	var definitions string
	for _, table := range *tables {
		if !s.config.CanDumpDefinition(table.Name) {
			continue
		}

		if strings.Index(strings.ToLower(table.Type), "table") >= 0 {
			drops = drops + fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table.Name)
		} else if strings.ToLower(table.Type) == "view" {
			drops = drops + fmt.Sprintf("DROP VIEW IF EXISTS `%s`;\n", table.Name)
		} else {
			return fmt.Errorf("can't drop table type of %q", table.Name)
		}

		definitions = definitions + fmt.Sprintf("%s;\n", table.Definition)
	}

	err := w.Write(fmt.Sprintf("%s\n", drops))
	if err != nil {
		return err
	}

	err = w.Write(fmt.Sprintf("%s\n", definitions))
	return err
}

func (s *Dumper) DumpData(w DumpWriter, tables *Tables) error {
	for _, table := range *tables {
		if strings.Index(strings.ToLower(table.Type), "table") < 0 {
			continue
		}

		rows, columns, err := s.selectAllDataFor(table)
		if err != nil {
			return err
		}

		values := make([]*sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		query := fmt.Sprintf("INSERT INTO `%s` VALUES", table.Name)
		var data []string
		for rows.Next() {
			if err = rows.Scan(scanArgs...); err != nil {
				return err
			}
			var vals []string
			for _, col := range values {
				val := "NULL"
				if col != nil {
					val = fmt.Sprintf("'%s'", MySQLEscape(string(*col)))
				}
				vals = append(vals, val)
			}

			data = append(data, fmt.Sprintf("( %s )", strings.Join(vals, ", ")))
			if len(data) >= 100 {
				err := w.Write(fmt.Sprintf("%s\n%s;\n", query, strings.Join(data, ",\n")))
				if err != nil {
					return err
				}
				data = make([]string, 0)
			}
		}

		if len(data) > 0 {
			err := w.Write(fmt.Sprintf("%s\n%s;\n", query, strings.Join(data, ",\n")))
			if err != nil {
				return err
			}
		}

		rows.Close()
	}

	return nil
}

func (s *Dumper) GetDefinitions() (*Tables, error) {
	list := &Tables{}

	_, err := s.db.Exec("SET SESSION group_concat_max_len = 1000000;")
	if err != nil {
		return nil, err
	}

	sqlQuery := `
		SELECT 
			t.table_name,
			t.table_type,
			(t.data_length + t.index_length),
			GROUP_CONCAT(c.column_name ORDER BY c.ordinal_position SEPARATOR '|') 
		FROM information_schema.tables t 
		JOIN information_schema.columns c ON (t.table_schema=c.table_schema AND t.table_name=c.table_name) 
		WHERE 
			t.table_schema=DATABASE() 
		GROUP BY 
			t.table_schema, t.table_name
		ORDER BY 
			t.table_type;
	`

	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var concatedColumns string
	for rows.Next() {
		table := &Table{}
		err = rows.Scan(&table.Name, &table.Type, &table.Size, &concatedColumns)
		if err != nil {
			return nil, err
		}

		table.Columns = &Columns{}
		*table.Columns = append(*table.Columns, strings.Split(concatedColumns, "|")...)

		table.Definition, err = s.GetDefinition(table.Name, table.Type)
		if err != nil {
			return nil, err
		}

		*list = append(*list, *table)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	// order views by references
	s.reorderViews(list)

	return list, nil
}

func (s *Dumper) GetDefinition(objName, objType string) (string, error) {
	s.logger.Debugf("Getting definition for the %q which is type of %q.", objName, objType)

	if strings.Index(strings.ToLower(objType), "table") >= 0 {
		return s.getTableDefinition(objName)
	} else if strings.ToLower(objType) == "view" {
		return s.getViewDefinition(objName)
	}
	return "", fmt.Errorf("wrong type %q", objType)
}

func (s *Dumper) getTableDefinition(objName string) (string, error) {
	sqlQuery := fmt.Sprintf("SHOW CREATE TABLE `%s`", objName)

	var tableName, tableDefinition string
	err := s.db.QueryRow(sqlQuery).Scan(&tableName, &tableDefinition)
	if err != nil {
		return "", err
	}

	return tableDefinition, nil
}

func (s *Dumper) getViewDefinition(objName string) (string, error) {
	sqlQuery := fmt.Sprintf("SHOW CREATE VIEW `%s`", objName)

	var viewName, viewDefinition, characterSet, collationConnection string
	err := s.db.QueryRow(sqlQuery).Scan(&viewName, &viewDefinition, &characterSet, &collationConnection)
	if err != nil {
		return "", err
	}

	return viewDefinition, nil
}

func (s *Dumper) selectAllDataFor(table Table) (*sql.Rows, []string, error) {
	filter := s.config.GetDumpFilter(table)
	selectQuery := fmt.Sprintf("SELECT %s FROM `%s` %s;", s.config.GetDumpColumns(table), table.Name, filter)

	s.logger.Debugf("Executing query %q", selectQuery)

	rows, err := s.db.Query(selectQuery)
	if err != nil {
		return nil, nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	return rows, columns, nil
}

func (s *Dumper) reorderViews(tables *Tables) {
	swapped := true
	for swapped {
		swapped = false
		for i := 0; i < len(*tables); i++ {
			if (*tables)[i].Type == "VIEW" {
				tbl1 := (*tables)[i]
				j := s.findReferenceIndex(tbl1.Name, tables)
				if j != -1 && j < i {
					// swap values
					tmp := (*tables)[j]
					(*tables)[j] = (*tables)[i]
					(*tables)[i] = tmp

					swapped = true
				}
			}
		}
	}
}

func (s *Dumper) findReferenceIndex(viewName string, tables *Tables) int {
	for idx, table := range *tables {
		if table.Type == "VIEW" && table.Name != viewName {
			if strings.Contains(table.Definition, viewName) {
				return idx
			}
		}
	}

	return -1
}
