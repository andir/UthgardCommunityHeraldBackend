package main

import (
	"log"
	"sync"
	"time"

	"github.com/andir/UthgardCommunityHeraldBackend/timeseries"
)

func UpdateTimeseries(statistics *Statistics, now time.Time) {
	type Item struct {
		metric string
		value  uint64
	}

	if now.IsZero() {
		now = time.Now()
	}

	log.Println("Updating timeseries...")

	wg := &sync.WaitGroup{}

	// create timeseries per guild
	wg.Add(1)
	go func() {
		for guildName, query := range statistics.ByGuild {
			if len(guildName) == 0 {
				continue
			}

			metrics := []Item{
				{"count", uint64(len(query.Characters))},
				{"xp", query.TotalXP},
				{"rp", query.TotalRP},
			}
			for _, metric := range metrics {
				path := timeseries.TimeSeriesPath("guild", guildName, metric.metric)
				err := timeseries.UpdateSeries(path, metric.value, now)
				if err != nil {
					log.Printf("Failed to save TS for %v: %v", guildName, err)
				}
			}
		}
		wg.Done()
	}()

	// Time series per realm
	wg.Add(1)
	go func() {
		for realm, query := range statistics.ByRealm {
			metrics := []Item{
				{"count", uint64(len(query.Characters))},
				{"xp", query.TotalRP},
				{"rp", query.TotalRP},
			}
			for _, metric := range metrics {
				path := timeseries.TimeSeriesPath("realm", realm, metric.metric)
				err := timeseries.UpdateSeries(path, metric.value, now)
				if err != nil {
					log.Printf("failed to save TS for %v: %v", realm, err)
				}
			}
		}
		wg.Done()
	}()

	// time series per character
	wg.Add(1)
	go func() {
		for characterName, character := range statistics.Characters {
			metrics := []Item{
				{"xp", character.Xp},
				{"rp", character.Rp},
			}
			for _, metric := range metrics {
				ts := time.Unix(character.LastUpdated, 0)
				path := timeseries.TimeSeriesPath("character", characterName, metric.metric)
				err := timeseries.UpdateSeries(path, metric.value, ts)
				if err != nil {
					log.Printf("failed to save TS for %v: %v", characterName, err)
				}
			}
		}
		wg.Done()
	}()

	// time series per class
	wg.Add(1)
	go func() {
		for class, query := range statistics.ByClass {
			metrics := []Item{
				{"count", uint64(len(query.Characters))},
				{"xp", query.TotalXP},
				{"rp", query.TotalRP},
			}
			for _, metric := range metrics {
				path := timeseries.TimeSeriesPath("class", class, metric.metric)
				err := timeseries.UpdateSeries(path, metric.value, now)
				if err != nil {
					log.Printf("Failed to save TS for %v: %v", class, err)
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()
	log.Println("Done")
}
