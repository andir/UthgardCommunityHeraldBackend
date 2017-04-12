package timeseries

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TimeSeriesEntry struct {
	Value     uint64
	Timestamp time.Time
}

type TimeSeriesEntryArray []TimeSeriesEntry

type TimeSeries struct {
	Entries TimeSeriesEntryArray
}

func (t *TimeSeriesEntryArray) Less(a, b int) bool {
	return (*t)[a].Timestamp.Before((*t)[b].Timestamp)
}

func (t *TimeSeriesEntryArray) Swap(a, b int) {
	(*t)[a], (*t)[b] = (*t)[b], (*t)[a]
}

func (t *TimeSeriesEntryArray) Len() int {
	return len(*t)
}

func (t *TimeSeries) Append(value uint64, timestamp time.Time) (changed bool) {
	if t.Entries == nil {
		t.Entries = make(TimeSeriesEntryArray, 0)
	}
	// scan for an entry with the same timestamp
	l := len(t.Entries)
	for i := range t.Entries {
		n := l - 1 - i
		entry := t.Entries[n]
		if entry.Timestamp.Before(timestamp) {
			// time is sorted, if this is before the current timeframe just bail
			break
		}

		if entry.Timestamp.Equal(timestamp) {
			// update value and bail
			if entry.Value != value {
				entry.Value = value
			}
			changed = false
			return
		}
	}

	t.Entries = append(t.Entries,
		TimeSeriesEntry{
			Value:     value,
			Timestamp: timestamp,
		})
	// ensure sort order by timestamp
	sort.Sort(&t.Entries)
	changed = true
	return
}

func compact(a, b TimeSeriesEntry) TimeSeriesEntry {
	if b.Value >= a.Value {
		return b
	}
	return a
}

func (t *TimeSeries) Compact() {
	COMPACT_THRESHOLD := time.Hour * 24 * 30 * 1
	n := len(t.Entries)
	// create a new series
	series := make([]TimeSeriesEntry, 0, n)

	i := 0

	for i < n-1 {
		a, b := t.Entries[i], t.Entries[i+1]

		if b.Timestamp.Sub(a.Timestamp) < COMPACT_THRESHOLD {
			series = append(series, compact(a, b))
			i += 2
			continue
		} else {
			series = append(series, a)
		}

		i += 1
	}
	t.Entries = series
}

func (t *TimeSeries) EntriesSince(date time.Time) []TimeSeriesEntry {
	results := make([]TimeSeriesEntry, 0, 1024)

	for _, e := range t.Entries {
		if e.Timestamp.After(date) {
			results = append(results, e)
		}
	}
	return results
}
func (t *TimeSeries) ValueSince(date time.Time) int64 {
	series := t.EntriesSince(date)
	l := len(series)
	if l == 0 {
		return 0
	} else if l == 1 {
		return int64(series[0].Value)
	} else {
		first := series[0]
		last := series[l-1]
		if first.Timestamp.Before(last.Timestamp) {
			return int64(last.Value) - int64(first.Value)
		}
		return int64(first.Value) - int64(last.Value)
	}
}

func OpenOrCreateTimeseries(filename string) *TimeSeries {
	var series *TimeSeries
	series = nil
	_, err := os.Stat(filename)

	series = &TimeSeries{}

	var ts *TimeSeries = nil

	if !os.IsNotExist(err) {
		ts = OpenTimeSeries(filename)

		if ts == nil {
			log.Printf("%v seems to be corrupt. Purging.", filename)
		}
	}

	if ts == nil {
		log.Printf("Creating new ts: %v", filename)
		series.Save(filename)
		ts = series
	}

	return ts
}

func OpenTimeSeries(filename string) *TimeSeries {
	ts := &TimeSeries{}
	fh, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer fh.Close()

	gReader, err := gzip.NewReader(fh)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer gReader.Close()

	data, err := ioutil.ReadAll(gReader)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = json.Unmarshal(data, ts)
	if err != nil {
		log.Println(err)
		return nil
	}
	return ts
}

func (t *TimeSeries) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TimeSeries) Save(filename string) error {
	ensureDir(filename)
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	gWriter := gzip.NewWriter(fh)
	defer gWriter.Close()

	bytes, err := t.Serialize()
	if err != nil {
		return err
	}
	gWriter.Write(bytes)
	return nil
}

func ensureDir(path string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
}

func TimeSeriesPath(queryType, metric, key string) string {
	metric = strings.ToLower(metric)
	key = strings.ToLower(key)
	var sm string
	if len(metric) < 2 {
		sm = metric
	} else {
		sm = metric[0:2]
	}
	return filepath.Join("data", queryType, sm, metric, key+".json.gz")
}

func UpdateSeries(fn string, value uint64, timestamp time.Time) (err error) {
	ts := OpenOrCreateTimeseries(fn)
	if ts == nil {
		log.Printf("Failed to create ts for: %v", fn)
		return
	}
	changed := ts.Append(value, timestamp)
	if changed {
		err = ts.Save(fn)
	}
	return
}

func GuildXPTimeSeries(guildName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("guild", guildName, "xp"))

}

func GuildRPTimeSeries(guildName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("guild", guildName, "rp"))
}

func CharacterXPTimeSeries(characterName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("character", characterName, "xp"))
}

func CharacterRPTimeSeries(characterName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("character", characterName, "rp"))
}
func RealmXPTimeSeries(realmName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("realm", realmName, "xp"))
}

func RealmRPTimeSeries(realmName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("realm", realmName, "rp"))
}

func ClassXPTimeSeries(className string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("class", className, "xp"))
}

func ClassRPTimeSeries(className string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("class", className, "rp"))
}

func ClassCountTimeSeries(className string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("class", className, "count"))
}

func GuildCountTimeSeries(guildName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("guild", guildName, "count"))
}

func RealmCountTimeSeries(realmName string) *TimeSeries {
	return OpenTimeSeries(TimeSeriesPath("realm", realmName, "count"))
}
