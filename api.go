package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andir/UthgardCommunityHeraldBackend/timeseries"
	"github.com/gorilla/mux"
)

const MAX_RESULTS = 1000

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

type APIFunction func(wr http.ResponseWriter, req *http.Request) (response, err interface{})

func apiEndpointWrapper(fun APIFunction) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		response, error := fun(wr, req)
		wr.Header().Set("Content-Type", "application/json; encoding=utf-8")
		if error != nil {
			log.Println(error)
			wr.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(wr).Encode(struct {
				Error string
			}{
				Error: "Well this didn't end up as expected :(",
			})
		} else {
			wr.Header().Set("Cache-Control", "max-age=600")
			wr.WriteHeader(http.StatusOK)
			json.NewEncoder(wr).Encode(response)
		}
	})
}

// ---

func characterEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	characterName := strings.ToLower(vars["characterName"])
	char, ok := statistics.Characters[characterName]
	if !ok {
		return nil, "unknown character" // TODO error
	}
	return char, nil
}

// ---

func topRP(wr http.ResponseWriter, query *Query) (interface{}, interface{}) { //players CharactersByRP) {
	type TopRPPLayer struct {
		Name  string
		Guild string
		Rp    uint64
	}
	players := query.SortedByRP
	n := min(len(players), MAX_RESULTS)
	jplayers := make([]TopRPPLayer, n)
	for i, player := range players[:n] {
		jplayers[i] = TopRPPLayer{
			Name:  player.Name,
			Guild: player.Guild,
			Rp:    player.Rp,
		}
	}

	return &jplayers, nil
}

func topXP(wr http.ResponseWriter, query *Query) (interface{}, interface{}) {
	type TopXPPLayer struct {
		Name  string
		Guild string
		Xp    uint64
	}
	players := query.SortedByXP
	n := min(len(players), MAX_RESULTS)
	jplayers := make([]TopXPPLayer, n)
	for i, player := range players[:n] {
		jplayers[i] = TopXPPLayer{
			Name:  player.Name,
			Guild: player.Guild,
			Xp:    player.Xp,
		}
	}

	return &jplayers, nil
}

func univeralTopRPEndpoint(wr http.ResponseWriter, req *http.Request, key string, index map[string]*Query) (interface{}, interface{}) {
	vars := mux.Vars(req)
	val := vars[key]
	stats, ok := index[val]
	if ok {
		return topRP(wr, stats)
	}

	return nil, key + " not found"

}

func universalTopXPEndpoint(wr http.ResponseWriter, req *http.Request, key string, index map[string]*Query) (interface{}, interface{}) {
	vars := mux.Vars(req)
	val := vars[key]
	stats, ok := index[val]
	if ok {
		return topXP(wr, stats)
	}
	return nil, key + "not found"
}

// ---

func topLWRPCharactersEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.LWRPCharacters[:MAX_RESULTS], nil
}
func topLWXPCharactersEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.LWXPCharacters[:MAX_RESULTS], nil
}
func topLWRPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.LWRPGuilds[:MAX_RESULTS], nil
}
func topLWXPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.LWXPGuilds[:MAX_RESULTS], nil
}

// ---

func topRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return topRP(wr, &statistics.Query)
}

func topXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return topXP(wr, &statistics.Query)
}

func totalRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.TotalRP, nil
}

func totalXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return statistics.TotalXP, nil
}

// ---

func topRPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	n := min(len(statistics.TopRPGuilds), MAX_RESULTS)
	return statistics.TopRPGuilds[:n], nil
}

func topXPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	n := min(len(statistics.TopXPGuilds), MAX_RESULTS)
	return statistics.TopXPGuilds[:n], nil
}

// ---

func totalClassRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return univeralTopRPEndpoint(wr, req, "className", statistics.ByClass)
}

func totalClassXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return universalTopXPEndpoint(wr, req, "className", statistics.ByClass)
}

func topClassXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return universalTopXPEndpoint(wr, req, "className", statistics.ByClass)
}

func topClassRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return univeralTopRPEndpoint(wr, req, "className", statistics.ByClass)
}

// ---

func totalRealmRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	realmName := vars["realmName"]
	stats := statistics.ByRealm[realmName]
	return stats.TotalRP, nil
}

func totalRealmXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	realmName := vars["realmName"]
	stats := statistics.ByRealm[realmName]

	return stats.TotalXP, nil
}

func topRealmXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return universalTopXPEndpoint(wr, req, "realmName", statistics.ByRealm)
}

func topRealmRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return univeralTopRPEndpoint(wr, req, "realmName", statistics.ByRealm)
}

// ---

func totalGuildRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	guildName := vars["guildName"]
	stats := statistics.ByGuild[guildName]

	return stats.TotalRP, nil
}

func totalGuildXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	guildName := vars["guildName"]
	stats := statistics.ByGuild[guildName]

	return stats.TotalXP, nil
}

func topGuildXPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return universalTopXPEndpoint(wr, req, "guildName", statistics.ByRealm)
}

func topGuildRPEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return univeralTopRPEndpoint(wr, req, "guildName", statistics.ByGuild)
}

type TSGetter func(key string) *timeseries.TimeSeries

func timeSeriesRenderer(req *http.Request, key string, getter TSGetter, wr http.ResponseWriter) (interface{}, interface{}) {
	val := mux.Vars(req)[key]
	ts := getter(val)
	if ts == nil {
		return nil, "unknown timeseries" // FIXME: return proper error
	}

	return ts, nil
}

func timeSeriesValueSince(getter TSGetter, key string, since time.Duration) APIFunction {
	return func(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
		val := mux.Vars(req)[key]
		ts := getter(val)

		timestamp := time.Now().Add(-1 * since)

		return ts.ValueSince(timestamp), nil
	}
}

func guildRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "guildName", timeseries.GuildRPTimeSeries, wr)
}
func guildXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "guildName", timeseries.GuildXPTimeSeries, wr)
}
func realmRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "realmName", timeseries.RealmRPTimeSeries, wr)
}
func realmXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "realmName", timeseries.RealmXPTimeSeries, wr)
}
func classRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "className", timeseries.ClassRPTimeSeries, wr)
}
func classXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "className", timeseries.ClassXPTimeSeries, wr)
}
func characterRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "characterName", timeseries.CharacterRPTimeSeries, wr)
}
func characterXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "characterName", timeseries.CharacterXPTimeSeries, wr)
}
func guildCountHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "guildName", timeseries.GuildCountTimeSeries, wr)
}
func realmCountHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "realmName", timeseries.RealmCountTimeSeries, wr)
}
func classCountHistoryEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	return timeSeriesRenderer(req, "className", timeseries.ClassCountTimeSeries, wr)
}

func characterRPSinceEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	characterName := mux.Vars(req)["characterName"]
	ts := timeseries.CharacterRPTimeSeries(characterName)

	if ts == nil {
		return nil, "time series not found"
	}
	return ts.ValueSince(time.Date(2017, time.March, 01, 00, 00, 00, 00, time.UTC)), nil
}

// ---

func searchCharacterEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	characterName := strings.ToLower(vars["characterName"])

	characters := make([]*Character, 0)
	n := 0
	characterTree.WalkPrefix(characterName, func(name string, value interface{}) bool {
		if character, ok := value.(*Character); ok {
			characters = append(characters, character)
		}
		n += 1
		if n > MAX_RESULTS {
			return true
		}
		return false
	})

	return characters, nil
}

func searchGuildEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	guildName := strings.ToLower(vars["guildName"])

	log.Printf("guildname: %v", guildName)
	guilds := make([]string, 0)
	n := 0
	guildTree.WalkPrefix(guildName, func(name string, value interface{}) bool {
		if realName, ok := value.(string); ok {
			guilds = append(guilds, realName)
		}
		n += 1
		if n > MAX_RESULTS {
			return true
		}
		return false
	})
	return guilds, nil
}

func guildEndpoint(wr http.ResponseWriter, req *http.Request) (interface{}, interface{}) {
	vars := mux.Vars(req)
	guildName, ok := vars["guildName"]
	if !ok {
		return nil, "missing guildName parameter"
	}
	guild, ok := statistics.ByGuild[guildName]
	if !ok {
		return nil, "guild not found"
	}

	return guild.Characters, nil
}
