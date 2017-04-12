package main

import (
	"log"
	"sort"
	"time"

	"github.com/andir/UthgardCommunityHeraldBackend/timeseries"
)

func UpdateTopLWRP(stats *Statistics) {
	lw := time.Now().Add(-24 * 8 * time.Hour)
	log.Println("Calculating Character LWRP")

	stats.LWRPCharacters = make(CharactersByLWRP, 0)
	stats.LWXPCharacters = make(CharactersByLWXP, 0)

	for _, char := range stats.Characters {
		ts := timeseries.CharacterRPTimeSeries(char.Name)
		lwrp := ts.ValueSince(lw)
		ts = timeseries.CharacterXPTimeSeries(char.Name)
		lwxp := ts.ValueSince(lw)
		char.LastWeekRp = lwrp
		char.LastWeekXp = lwxp
		if lwxp > 10000000000 {
			log.Printf("XP %v: %v", char.Name, lwxp)
		}

		stats.LWRPCharacters = append(stats.LWRPCharacters, char)
		stats.LWXPCharacters = append(stats.LWXPCharacters, char)
	}
	sort.Sort(&stats.LWRPCharacters)
	sort.Sort(&stats.LWXPCharacters)
	log.Println("Done")

	log.Println("calculating Guild LWRP/XP")

	stats.LWRPGuilds = make(GuildsByLWRP, 0, 1000)
	stats.LWXPGuilds = make(GuildsByLWXP, 0, 1000)

	for guildName, guild := range stats.Guilds {
		if guildName == "" {
			continue
		}
		ts := timeseries.GuildRPTimeSeries(guildName)
		lwrp := ts.ValueSince(lw)
		ts = timeseries.GuildXPTimeSeries(guildName)
		lwxp := ts.ValueSince(lw)

		guild.LWRP = lwrp
		guild.LWXP = lwxp

		stats.LWRPGuilds = append(stats.LWRPGuilds, guild)
		stats.LWXPGuilds = append(stats.LWXPGuilds, guild)
		stats.ByGuild[guildName].LWRP = lwrp
		stats.ByGuild[guildName].LWXP = lwxp
	}
	sort.Sort(&stats.LWRPGuilds)
	sort.Sort(&stats.LWXPGuilds)
	log.Println("Done")

	log.Println("calculating Class LWRP/XP")
	for className, query := range stats.ByClass {
		ts := timeseries.ClassRPTimeSeries(className)
		lwrp := ts.ValueSince(lw)
		ts = timeseries.ClassXPTimeSeries(className)
		lwxp := ts.ValueSince(lw)

		query.LWRP = lwrp
		query.LWXP = lwxp
	}
	log.Println("Done")

	log.Println("calculating realm LWRP/XP")
	for realmName, query := range stats.ByRealm {
		ts := timeseries.RealmRPTimeSeries(realmName)
		lwrp := ts.ValueSince(lw)
		ts = timeseries.RealmXPTimeSeries(realmName)
		lwxp := ts.ValueSince(lw)
		query.LWRP = lwrp
		query.LWXP = lwxp
	}
	log.Println("Done")

}
