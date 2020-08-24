package main

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
