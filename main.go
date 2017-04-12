package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andir/UthgardCommunityHeraldBackend/timeseries"
	radix "github.com/armon/go-radix"
	"github.com/gorilla/mux"
)

func parseDump(filename string) map[string]Character {
	fh, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	var characters map[string]Character
	err = json.Unmarshal(bytes, &characters)
	if err != nil {
		log.Fatal(err)
	}

	return characters
}

// FIXME: replace the global with a context var
var statistics *Statistics
var characterTree *radix.Tree
var guildTree *radix.Tree

func guildInfoEndpoint(wr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	if guild, ok := statistics.ByGuild[vars["guildName"]]; !ok {
		return
	} else {
		wr.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(wr)
		encoder.Encode(guild)
	}
}

func totalRPPlayers(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Set("Content-Type", "application/json; charset=utf-8")

	type TotalRPPlayersJSON struct {
		Name  string
		Guild string
		Realm string
		Class string
		RP    uint64
	}

	players := make([]TotalRPPlayersJSON, len(statistics.SortedByRP))
	for i, p := range statistics.SortedByRP {
		players[i] = TotalRPPlayersJSON{
			Name:  p.Name,
			Guild: p.Guild,
			Realm: p.Realm,
			Class: p.Class,
			RP:    p.Rp,
		}
	}

	encoder := json.NewEncoder(wr)
	encoder.Encode(players)
}

func update() {
	url := "https://www2.uthgard.net/herald/api/dump"
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}

	var characters map[string]*Character
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(bytes, &characters)

	var lastUpdatedInt int64 = 0

	for _, value := range characters {
		if value.LastUpdated > lastUpdatedInt {
			lastUpdatedInt = value.LastUpdated
		}
	}
	go func() {
		ctree := radix.New()
		for name, character := range characters {
			ctree.Insert(strings.ToLower(name), character)
		}
		characterTree = ctree
	}()

	stats := LoadCharacters(characters)

	go func() {
		gtree := radix.New()
		for name := range stats.ByGuild {
			gtree.Insert(strings.ToLower(name), name)
		}
		guildTree = gtree
	}()

	lastUpdated := time.Unix(lastUpdatedInt, 0)
	UpdateTimeseries(stats, lastUpdated)

	UpdateTopLWRP(stats)

	statistics = stats
}

func main() {

	go func() {
		t := time.NewTicker(30 * time.Minute)
		update()
		for range t.C {
			update()
		}
	}()

	//	for guild, data := range statistics.ByGuild {
	//		TimeSeries{}
	//	}

	r := mux.NewRouter()

	endpoints := []struct {
		Endpoint string
		Func     APIFunction
	}{

		{"/toprp", topRPEndpoint},
		{"/topxp", topXPEndpoint},
		{"/rp", totalRPEndpoint},
		{"/xp", totalXPEndpoint},

		{"/toplwxp", topLWRPCharactersEndpoint},
		{"/toplwrp", topLWXPCharactersEndpoint},

		{"/toplwxp/guilds", topLWRPGuildsEndpoint},
		{"/toplwrp/guilds", topLWXPGuildsEndpoint},

		{"/toprp/guilds", topRPGuildsEndpoint},
		{"/topxp/guilds", topXPGuildsEndpoint},

		{"/search/character/{characterName}", searchCharacterEndpoint},
		{"/search/guild/{guildName}", searchGuildEndpoint},

		{"/character/{characterName}", characterEndpoint},
		{"/character/{characterName}/lastwrp", timeSeriesValueSince(timeseries.CharacterRPTimeSeries, "characterName", 7*24*time.Hour)},
		{"/character/{characterName}/history/rp", characterRPHistoryEndpoint},
		{"/character/{characterName}/history/xp", characterXPHistoryEndpoint},

		{"/class/{className}/rp", totalClassRPEndpoint},
		{"/class/{className}/xp", totalClassXPEndpoint},
		{"/class/{className}/toprp", topClassRPEndpoint},
		{"/class/{className}/topxp", topClassXPEndpoint},
		{"/class/{className}/history/rp", classRPHistoryEndpoint},
		{"/class/{className}/history/xp", classXPHistoryEndpoint},
		{"/class/{className}/history/count", classCountHistoryEndpoint},

		{"/realm/{realmName}/rp", totalRealmRPEndpoint},
		{"/realm/{realmName}/xp", totalRealmXPEndpoint},
		{"/realm/{realmName}/toprp", topRealmRPEndpoint},
		{"/realm/{realmName}/topxp", topRealmXPEndpoint},
		{"/realm/{realmName}/history/rp", realmRPHistoryEndpoint},
		{"/realm/{realmName}/history/xp", realmXPHistoryEndpoint},
		{"/realm/{realmName}/history/count", realmCountHistoryEndpoint},

		{"/guild/{guildName}", guildEndpoint},
		{"/guild/{guildName}/rp", totalGuildRPEndpoint},
		{"/guild/{guildName}/xp", totalGuildXPEndpoint},
		{"/guild/{guildName}/toprp", topGuildRPEndpoint},
		{"/guild/{guildName}/topxp", topGuildXPEndpoint},
		{"/guild/{guildName}/lastwrp", timeSeriesValueSince(timeseries.GuildRPTimeSeries, "guildName", 7*24*time.Hour)},
		{"/guild/{guildName}/history/rp", guildRPHistoryEndpoint},
		{"/guild/{guildName}/history/xp", guildXPHistoryEndpoint},
		{"/guild/{guildName}/history/count", guildCountHistoryEndpoint},
	}

	documentation := ""

	for _, endpoint := range endpoints {
		log.Println(endpoint.Endpoint)
		r.Handle(endpoint.Endpoint, apiEndpointWrapper(endpoint.Func))

		documentation += endpoint.Endpoint + "\n"
	}
	r.Handle("/", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		wr.Header().Set("Content-Type", "text/plain")
		wr.Write([]byte(documentation))
	}))
	http.ListenAndServe("127.0.0.1:8081", r)
}
