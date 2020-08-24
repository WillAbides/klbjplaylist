package main

import (
	"time"
	_ "time/tzdata"
)

var tzLocation *time.Location

func init() {
	var err error
	tzLocation, err = time.LoadLocation("America/Chicago")
	if err != nil {
		panic(err)
	}
}

func main() {
	data, err := fetchXML()
	if err != nil {
		panic(err)
	}
	dataDir := "data"
	var plays []trackPlay
	plays, err = trackPlaysFromXML(data, plays)
	if err != nil {
		panic(err)
	}
	err = updateDataFiles(plays, dataDir)
	if err != nil {
		panic(err)
	}
}
