package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

	var charactersFromJSON map[string]Character
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(bytes, &charactersFromJSON)

	var lastUpdatedInt int64 = 0

	characters := make(map[string]*Character)
	for key, value := range charactersFromJSON {
		if value.LastUpdated > lastUpdatedInt {
			lastUpdatedInt = value.LastUpdated
		}
		c := &Character{}
		*c = value
		characters[key] = c
	}

	stats := LoadCharacters(characters)
	lastUpdated := time.Unix(lastUpdatedInt, 0)
	UpdateTimeseries(stats, lastUpdated)

	statistics = stats

}

func main() {
	//	filename := flag.String("filename", "", "Dump to parse")

	//	flag.Parse()

	//	if len(*filename) == 0 {
	//		flag.Usage()
	//		return
	//	}

	//charactersFromJSON := parseDump(*filename)
	//var lastUpdatedInt int64 = 0

	//characters := make(map[string]*Character)
	//for key, value := range charactersFromJSON {
	//	if value.LastUpdated > lastUpdatedInt {
	//		lastUpdatedInt = value.LastUpdated
	//	}
	//	c := &Character{}
	//	*c = value
	//	characters[key] = c
	//}

	//statistics = LoadCharacters(characters)
	//lastUpdated := time.Unix(lastUpdatedInt, 0)
	//UpdateTimeseries(statistics, lastUpdated)

	update()

	go func() {
		t := time.NewTicker(30 * time.Minute)
		for range t.C {
			update()
		}
	}()

	//	for guild, data := range statistics.ByGuild {
	//		TimeSeries{}
	//	}

	r := mux.NewRouter()

	//	realm/<realm>/rp - sum
	//	realm/<realm>/xp - sum
	//	realm/<realm>/topxp - list
	//	realm/<realm>/toprp - list
	//
	//	class
	//	guild
	//
	//	/topxp
	//	/toprp
	//	/xp
	//	/rp

	r.HandleFunc("/toprp", topRPEndpoint)
	r.HandleFunc("/topxp", topXPEndpoint)
	r.HandleFunc("/rp", totalRPEndpoint)
	r.HandleFunc("/xp", totalXPEndpoint)

	r.HandleFunc("/toprp/guilds", topRPGuildsEndpoint)
	r.HandleFunc("/topxp/guilds", topXPGuildsEndpoint)

	//r.HandleFunc("/toprp/classes", topRPClassesEndpoint)
	//r.HandleFunc("/topxp/classes", topXPClassesEndpoint)

	//r.HandleFunc("/toprp/realms", topRPRealmsEndpoint)
	//r.HandleFunc("/topxp/realms", topXPRealmsEndpoint)

	r.HandleFunc("/character/{characterName}", characterEndpoint)
	r.HandleFunc("/character/{characterName}/history/rp", characterRPHistoryEndpoint)
	r.HandleFunc("/character/{characterName}/history/xp", characterXPHistoryEndpoint)

	r.HandleFunc("/class/{className}/rp", totalClassRPEndpoint)
	r.HandleFunc("/class/{className}/xp", totalClassXPEndpoint)
	r.HandleFunc("/class/{className}/toprp", topClassRPEndpoint)
	r.HandleFunc("/class/{className}/topxp", topClassXPEndpoint)
	r.HandleFunc("/class/{className}/history/rp", classRPHistoryEndpoint)
	r.HandleFunc("/class/{className}/history/xp", classXPHistoryEndpoint)

	r.HandleFunc("/realm/{realmName}/rp", totalRealmRPEndpoint)
	r.HandleFunc("/realm/{realmName}/xp", totalRealmXPEndpoint)
	r.HandleFunc("/realm/{realmName}/toprp", topRealmRPEndpoint)
	r.HandleFunc("/realm/{realmName}/topxp", topRealmXPEndpoint)
	r.HandleFunc("/realm/{realmName}/history/rp", realmRPHistoryEndpoint)
	r.HandleFunc("/realm/{realmName}/history/xp", realmXPHistoryEndpoint)

	r.HandleFunc("/guild/{guildName}/rp", totalGuildRPEndpoint)
	r.HandleFunc("/guild/{guildName}/xp", totalGuildXPEndpoint)
	r.HandleFunc("/guild/{guildName}/toprp", topGuildRPEndpoint)
	r.HandleFunc("/guild/{guildName}/topxp", topGuildXPEndpoint)
	r.HandleFunc("/guild/{guildName}/history/rp", guildRPHistoryEndpoint)
	r.HandleFunc("/guild/{guildName}/history/xp", guildXPHistoryEndpoint)

	http.ListenAndServe("127.0.0.1:8081", r)
}
