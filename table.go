package dumper

import (
	"math/big"
	"strconv"

	humanize "github.com/dustin/go-humanize"
)

const (
	TABLE_FILTER_NONE     = ""
	TABLE_FILTER_NODATA   = "nodata"
	TABLE_FILTER_IGNORE   = "ignore"
	TABLE_FILTER_ONLYDATA = "onlydata"
)

var TableFilterMap = map[string]bool{
	TABLE_FILTER_NONE:     true,
	TABLE_FILTER_NODATA:   true,
	TABLE_FILTER_IGNORE:   true,
	TABLE_FILTER_ONLYDATA: true,
}

type Size int64

func (s *Size) String() string {
	if s == nil {
		return ""
	}

	return humanize.BigIBytes(big.NewInt(int64(*s)))
}

func (s *Size) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Size) UnmarshalJSON(data []byte) error {
	str, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	val, err := humanize.ParseBigBytes(str)
	if err != nil {
		return err
	}

	*s = Size(val.Int64())

	return nil
}

type Columns []string

func (s *Columns) IndexOf(value string) int {
	if s == nil {
		return -1
	}

	return IndexOf(*s, value)
}

type Table struct {
	Name       string   `json:"name,omitempty"`
	Type       string   `json:"type,omitempty"`
	Columns    *Columns `json:"columns,omitempty"`
	Size       *Size    `json:"size,omitempty"`
	Definition string   `json:"definition,omitempty"`
	IsFinished bool     `json:"is_finished,omitempty"`
}

type Tables []Table
