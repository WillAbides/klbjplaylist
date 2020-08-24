package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readExamples(t *testing.T) []trackPlay {
	t.Helper()
	ex1, err := ioutil.ReadFile("testdata/ex1.xml")
	require.NoError(t, err)
	ex2, err := ioutil.ReadFile("testdata/ex2.xml")
	require.NoError(t, err)
	var plays []trackPlay
	plays, err = trackPlaysFromXML(ex1, plays)
	require.NoError(t, err)
	plays, err = trackPlaysFromXML(ex2, plays)
	require.NoError(t, err)
	return plays
}

func Test_trackPlaysFromXML(t *testing.T) {
	ex1, err := ioutil.ReadFile("testdata/ex1.xml")
	require.NoError(t, err)
	ex2, err := ioutil.ReadFile("testdata/ex2.xml")
	require.NoError(t, err)
	var plays []trackPlay
	plays, err = trackPlaysFromXML(ex1, plays)
	require.NoError(t, err)
	plays, err = trackPlaysFromXML(ex2, plays)
	require.NoError(t, err)
	require.Len(t, plays, 69)
	wantFirst := trackPlay{
		StartTime: time.Date(2020, 8, 24, 10, 30, 17, 0, tzLocation),
		Title:     "Why Can't This Be Love",
		Artist:    "Van Halen",
		ProgramID: "203",
		Duration:  210 * time.Second,
	}
	require.Equal(t, wantFirst, plays[0])
	wantLast := trackPlay{
		StartTime: time.Date(2020, 8, 24, 4, 19, 0, 0, tzLocation),
		Title:     "Don't Do Me Like That",
		Artist:    "Tom Petty & The Heartbreaker",
		ProgramID: "201",
		Duration:  157 * time.Second,
	}
	require.Equal(t, wantLast, plays[len(plays)-1])
}

func Test_playsToCsv(t *testing.T) {
	plays := readExamples(t)
	want, err := ioutil.ReadFile("testdata/ex.csv")
	require.NoError(t, err)
	var buf bytes.Buffer
	err = playsToCSV(&buf, plays)
	require.NoError(t, err)
	require.Equal(t, string(want), buf.String())
}

func Test_playsFromCSV(t *testing.T) {
	want := readExamples(t)
	csvFile, err := os.Open("testdata/ex.csv")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, csvFile.Close())
	})
	got, err := playsFromCSV(csvFile)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func Test_trackPlaysByDay(t *testing.T) {
	examples := readExamples(t)
	for i := 0; i < 10; i++ {
		examples[i].StartTime = examples[i].StartTime.Add(24 * time.Hour)
	}
	got := trackPlaysByDay(examples)
	require.Len(t, got, 2)
	require.Len(t, got[time.Date(2020, 8, 24, 0, 0, 0, 0, tzLocation)], 59)
	require.Len(t, got[time.Date(2020, 8, 25, 0, 0, 0, 0, tzLocation)], 10)
}

func Test_csvFileName(t *testing.T) {
	date := time.Date(2020, 8, 24, 0, 0, 0, 0, tzLocation)
	got := csvFileName(date)
	want := "plays-2020-08-24.csv"
	require.Equal(t, want, got)
}

func Test_updateDataFiles(t *testing.T) {
	dataDir := filepath.FromSlash("testdata/data")
	examples := readExamples(t)
	for i := 0; i < 10; i++ {
		examples[i].StartTime = examples[i].StartTime.Add(24 * time.Hour)
	}
	err := updateDataFiles(examples, dataDir)
	require.NoError(t, err)
}
