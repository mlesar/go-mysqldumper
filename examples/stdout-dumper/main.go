package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	mysqldumper "github.com/mlesar/go-mysqldumper"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = time.RFC3339Nano
	log.SetFormatter(customFormatter)
}

func main() {
	// connect to the production DB
	db, err := NewDB("mysql", "productionuser:productionpass@tcp(8.8.8.8:3306)/google1?timeout=5s")
	if err != nil {
		log.Fatal(err)
	}

	// read the config file
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// parse config file
	config, err := mysqldumper.ParseConfig(configData)
	if err != nil {
		log.Fatal(err)
	}

	// create new dumper
	dumper := mysqldumper.New(config, db, log.StandardLogger())

	// create new dump writer
	w := mysqldumper.NewFileWriter(os.Stdout)

	// start dumping the data
	err = dumper.Dump(w)
	if err != nil {
		log.Fatal(err)
	}

}

func NewDB(driver, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}
