package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

// ---

func characterEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	characterName := strings.ToLower(vars["characterName"])
	char, ok := statistics.Characters[characterName]
	if !ok {
		return // TODO error
	}
	json.NewEncoder(wr).Encode(char)
}

// ---

func topRP(wr http.ResponseWriter, query *Query) { //players CharactersByRP) {
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

	json.NewEncoder(wr).Encode(&jplayers)
}

func topXP(wr http.ResponseWriter, query *Query) {
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

	json.NewEncoder(wr).Encode(&jplayers)
}

func univeralTopRPEndpoint(wr http.ResponseWriter, req *http.Request, key string, index map[string]Query) {
	vars := mux.Vars(req)
	val := vars[key]
	stats, ok := index[val]
	if ok {
		topRP(wr, &stats)
	}

}

func universalTopXPEndpoint(wr http.ResponseWriter, req *http.Request, key string, index map[string]Query) {
	vars := mux.Vars(req)
	val := vars[key]
	stats, ok := index[val]
	if ok {
		topXP(wr, &stats)
	}
}

func topRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	topRP(wr, &statistics.Query)
}

func topXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	topXP(wr, &statistics.Query)
}

func totalRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	json.NewEncoder(wr).Encode(statistics.TotalRP)
}

func totalXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	json.NewEncoder(wr).Encode(statistics.TotalXP)
}

// ---

func topRPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) {
	n := min(len(statistics.TopRPGuilds), MAX_RESULTS)
	json.NewEncoder(wr).Encode(statistics.TopRPGuilds[:n])
}

func topXPGuildsEndpoint(wr http.ResponseWriter, req *http.Request) {
	n := min(len(statistics.TopXPGuilds), MAX_RESULTS)
	json.NewEncoder(wr).Encode(statistics.TopXPGuilds[:n])
}

// ---

func totalClassRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	className := vars["className"]
	stats := statistics.ByClass[className]

	json.NewEncoder(wr).Encode(stats.TotalRP)
}

func totalClassXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	className := vars["className"]
	stats := statistics.ByClass[className]

	json.NewEncoder(wr).Encode(stats.TotalXP)
}

func topClassXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	universalTopXPEndpoint(wr, req, "className", statistics.ByClass)
}

func topClassRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	univeralTopRPEndpoint(wr, req, "className", statistics.ByClass)
}

// ---

func totalRealmRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	realmName := vars["realmName"]
	stats := statistics.ByRealm[realmName]

	json.NewEncoder(wr).Encode(stats.TotalRP)
}

func totalRealmXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	realmName := vars["realmName"]
	stats := statistics.ByRealm[realmName]

	json.NewEncoder(wr).Encode(stats.TotalXP)
}

func topRealmXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	universalTopXPEndpoint(wr, req, "realmName", statistics.ByRealm)
}

func topRealmRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	univeralTopRPEndpoint(wr, req, "realmName", statistics.ByRealm)
}

// ---

func totalGuildRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	guildName := vars["guildName"]
	stats := statistics.ByGuild[guildName]

	json.NewEncoder(wr).Encode(stats.TotalRP)
}

func totalGuildXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	guildName := vars["guildName"]
	stats := statistics.ByGuild[guildName]

	json.NewEncoder(wr).Encode(stats.TotalXP)
}

func topGuildXPEndpoint(wr http.ResponseWriter, req *http.Request) {
	universalTopXPEndpoint(wr, req, "guildName", statistics.ByRealm)
}

func topGuildRPEndpoint(wr http.ResponseWriter, req *http.Request) {
	univeralTopRPEndpoint(wr, req, "guildName", statistics.ByGuild)
}

type TSGetter func(key string) *timeseries.TimeSeries

func timeSeriesRenderer(key string, getter TSGetter, wr http.ResponseWriter) {
	ts := getter(key)
	if ts == nil {
		return // FIXME: return proper error
	}
	bytes, err := ts.Serialize()
	if err != nil {
		log.Println(err)
		return
	}
	wr.Write(bytes)
}

func guildRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["guildName"], timeseries.GuildRPTimeSeries, wr)
}
func guildXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["guildName"], timeseries.GuildXPTimeSeries, wr)
}
func realmRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["realmName"], timeseries.RealmRPTimeSeries, wr)
}
func realmXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["realmName"], timeseries.RealmXPTimeSeries, wr)
}
func classRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["className"], timeseries.ClassRPTimeSeries, wr)
}
func classXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["className"], timeseries.ClassXPTimeSeries, wr)
}
func characterRPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["characterName"], timeseries.CharacterRPTimeSeries, wr)
}
func characterXPHistoryEndpoint(wr http.ResponseWriter, req *http.Request) {
	timeSeriesRenderer(mux.Vars(req)["characterName"], timeseries.CharacterXPTimeSeries, wr)
}
