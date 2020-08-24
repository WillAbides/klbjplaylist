package main

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

type playingDoc struct {
	Entries []playingEntry `xml:"nowplaying-info"`
}

type entryProperty struct {
	Name  string `xml:"name,attr"`
	Value []byte `xml:",chardata"`
}

type playingEntry struct {
	Timestamp  int64           `xml:"timestamp,attr"`
	Properties []entryProperty `xml:"property"`
}

func (e playingEntry) trackPlay() (trackPlay, error) {
	play := trackPlay{
		StartTime: time.Unix(e.Timestamp, 0).UTC(),
	}
	for _, property := range e.Properties {
		switch property.Name {
		case "cue_title":
			play.Title = string(property.Value)
		case "track_artist_name":
			play.Artist = string(property.Value)
		case "program_id":
			play.ProgramID = string(property.Value)
		case "cue_time_duration":
			durMils, err := strconv.Atoi(string(property.Value))
			if err != nil {
				return play, err
			}
			play.Duration = time.Duration(durMils) * time.Millisecond
		}
	}
	return play, nil
}

type trackPlay struct {
	StartTime time.Time
	Title     string
	Artist    string
	ProgramID string
	Duration  time.Duration
}

func trackPlaysFromXML(xmlData []byte, plays []trackPlay) ([]trackPlay, error) {
	var doc playingDoc
	err := xml.Unmarshal(xmlData, &doc)
	if err != nil {
		return nil, err
	}
	for _, entry := range doc.Entries {
		play, err := entry.trackPlay()
		if err != nil {
			return nil, err
		}
		plays = addToPlays(&play, plays)
	}
	sort.Slice(plays, func(i, j int) bool {
		return plays[i].StartTime.After(plays[j].StartTime)
	})
	return plays, nil
}

func trackPlaysByDay(plays []trackPlay) map[time.Time][]trackPlay {
	result := map[time.Time][]trackPlay{}
	for i := range plays {
		play := plays[i]
		tm := play.StartTime
		day := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.UTC)
		result[day] = addToPlays(&play, result[day])
	}
	return result
}

func addToPlays(play *trackPlay, plays []trackPlay) []trackPlay {
	if plays == nil {
		plays = []trackPlay{}
	}
	for _, existingPlay := range plays {
		if existingPlay.StartTime.Equal(play.StartTime) {
			return plays
		}
	}
	plays = append(plays, *play)
	return plays
}

func playsToCSV(out io.Writer, plays []trackPlay) error {
	header := []string{"start time", "title", "artist", "duration", "program id"}
	w := csv.NewWriter(out)
	err := w.Write(header)
	if err != nil {
		return err
	}

	for _, play := range plays {
		err = w.Write([]string{
			play.StartTime.Format(time.RFC3339),
			play.Title,
			play.Artist,
			strconv.Itoa(int(play.Duration / time.Second)),
			play.ProgramID,
		})
		if err != nil {
			return err
		}
	}
	w.Flush()
	if w.Error() != nil {
		return w.Error()
	}
	return nil
}

func playsFromCSV(in io.Reader) ([]trackPlay, error) {
	r := csv.NewReader(in)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}
	records = records[1:]
	result := make([]trackPlay, len(records))
	for i := range records {
		record := records[i]
		play := trackPlay{
			Title:     record[1],
			Artist:    record[2],
			ProgramID: record[4],
		}
		play.StartTime, err = time.Parse(time.RFC3339, record[0])
		if err != nil {
			return nil, err
		}
		durSecs, err := strconv.Atoi(record[3])
		if err != nil {
			return nil, err
		}
		play.Duration = time.Duration(durSecs) * time.Second
		result[i] = play
	}
	return result, nil
}

func csvFileName(date time.Time) string {
	return date.Format("plays-2006-01-02.csv")
}

func updateDataFiles(plays []trackPlay, dataDir string) error {
	err := os.MkdirAll(dataDir, 0o700)
	if err != nil {
		return err
	}
	playsByDay := trackPlaysByDay(plays)
	for date, newPlays := range playsByDay {
		filename := filepath.Join(dataDir, csvFileName(date))
		var plays []trackPlay
		existingData, err := ioutil.ReadFile(filename) //nolint:gosec // so insecure
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if len(existingData) != 0 {
			plays, err = playsFromCSV(bytes.NewReader(existingData))
			if err != nil {
				return err
			}
		}
		for i := range newPlays {
			play := newPlays[i]
			plays = addToPlays(&play, plays)
		}
		sort.Slice(plays, func(i, j int) bool {
			return plays[i].StartTime.After(plays[j].StartTime)
		})
		outFile, err := os.Create(filename)
		if err != nil {
			return err
		}
		err = playsToCSV(outFile, plays)
		if err != nil {
			_ = outFile.Close() //nolint:errcheck // already returning an error
			return err
		}
		err = outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchXML() ([]byte, error) {
	u := `http://np.tritondigital.com/public/nowplaying?mountName=KLBJFM&numberToFetch=50&eventType=track`
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}
