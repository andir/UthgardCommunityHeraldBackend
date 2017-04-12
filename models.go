package main

import "sort"

type CharactersByRP []*Character

func (s *CharactersByRP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *CharactersByRP) Less(a, b int) bool { return (*s)[a].Rp > (*s)[b].Rp }
func (s *CharactersByRP) Len() int           { return len(*s) }

type CharactersByXP []*Character

func (s *CharactersByXP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *CharactersByXP) Less(a, b int) bool { return (*s)[a].Xp > (*s)[b].Xp }
func (s *CharactersByXP) Len() int           { return len(*s) }

type GuildsByXP []*Guild

func (s *GuildsByXP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *GuildsByXP) Less(a, b int) bool { return (*s)[a].XP > (*s)[b].XP }
func (s *GuildsByXP) Len() int           { return len(*s) }

type GuildsByRP []*Guild

func (s *GuildsByRP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *GuildsByRP) Less(a, b int) bool { return (*s)[a].RP > (*s)[b].RP }
func (s *GuildsByRP) Len() int           { return len(*s) }

type GuildsByLWRP []*Guild

func (s *GuildsByLWRP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *GuildsByLWRP) Less(a, b int) bool { return (*s)[a].LWRP > (*s)[b].LWRP }
func (s *GuildsByLWRP) Len() int           { return len(*s) }

type GuildsByLWXP []*Guild

func (s *GuildsByLWXP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *GuildsByLWXP) Less(a, b int) bool { return (*s)[a].LWXP > (*s)[b].LWXP }
func (s *GuildsByLWXP) Len() int           { return len(*s) }

type CharactersByLWRP []*Character

func (s *CharactersByLWRP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *CharactersByLWRP) Less(a, b int) bool { return (*s)[a].LastWeekRp > (*s)[b].LastWeekRp }
func (s *CharactersByLWRP) Len() int           { return len(*s) }

type CharactersByLWXP []*Character

func (s *CharactersByLWXP) Swap(a, b int)      { (*s)[b], (*s)[a] = (*s)[a], (*s)[b] }
func (s *CharactersByLWXP) Less(a, b int) bool { return (*s)[a].LastWeekXp > (*s)[b].LastWeekXp }
func (s *CharactersByLWXP) Len() int           { return len(*s) }

type Query struct {
	Characters map[string]*Character
	SortedByRP CharactersByRP
	SortedByXP CharactersByXP
	TotalRP    uint64
	TotalXP    uint64
	LWRP       int64
	LWXP       int64
	//	SortedByLWRP []*Character
}

func (q *Query) FromCharacters(characters map[string]*Character) {
	charmap := make(map[string]*Character)

	chars := make([]*Character, 0)
	for key, val := range characters {
		charmap[key] = val
		chars = append(chars, val)
	}
	q.Characters = charmap
	q.SortedByRP = make(CharactersByRP, len(chars))
	q.SortedByXP = make(CharactersByXP, len(chars))
	q.TotalRP = 0
	q.TotalXP = 0
	for i, c := range chars {
		q.SortedByRP[i] = c
		q.SortedByXP[i] = c
		q.TotalRP += c.Rp
		q.TotalXP += c.Xp
	}
	sort.Sort(&q.SortedByRP)
	sort.Sort(&q.SortedByXP)

}

type Guild struct {
	Name string
	RP   uint64
	XP   uint64

	LWRP int64
	LWXP int64
}

type Statistics struct {
	Query
	ByRealm map[string]*Query // key: realm
	ByClass map[string]*Query // key: class
	ByGuild map[string]*Query // key: guild name

	TopRPGuilds GuildsByRP
	TopXPGuilds GuildsByXP

	LWRPGuilds GuildsByLWRP
	LWXPGuilds GuildsByLWXP

	LWRPCharacters CharactersByLWRP
	LWXPCharacters CharactersByLWXP

	Guilds map[string]*Guild
}

type Character struct {
	Name             string
	Guild            string
	Race             string
	Class            string
	Realm            string
	Xp               uint64
	Rp               uint64
	Level            int
	RealmRank        int
	XpPercentOfLevel float32
	RpPercentOfLevel float32
	LastWeekRp       int64
	LastWeekXp       int64
	LastUpdated      int64
}

func LoadCharacters(characters map[string]*Character) *Statistics {
	s := &Statistics{
		ByRealm:     make(map[string]*Query),
		ByClass:     make(map[string]*Query),
		ByGuild:     make(map[string]*Query),
		TopRPGuilds: make(GuildsByRP, 0),
		TopXPGuilds: make(GuildsByXP, 0),
		Guilds:      make(map[string]*Guild),
	}

	s.Query.FromCharacters(characters)

	byRealm := make(map[string]map[string]*Character)
	byClass := make(map[string]map[string]*Character)
	byGuild := make(map[string]map[string]*Character)

	for name, char := range characters {
		realmName := char.Realm
		realm, ok := byRealm[realmName]
		if !ok {
			realm = make(map[string]*Character)
			byRealm[realmName] = realm
		}

		className := char.Class
		class, ok := byClass[className]
		if !ok {
			class = make(map[string]*Character)
			byClass[className] = class
		}

		guildName := char.Guild
		guild, ok := byGuild[guildName]
		if !ok {
			guild = make(map[string]*Character)
			byGuild[guildName] = guild
		}

		realm[name] = char
		class[name] = char
		guild[name] = char
	}
	for realm, characters := range byRealm {
		q := &Query{}
		q.FromCharacters(characters)
		s.ByRealm[realm] = q
	}

	for class, characters := range byClass {
		q := &Query{}
		q.FromCharacters(characters)
		s.ByClass[class] = q
	}

	for guild, characters := range byGuild {
		q := &Query{}
		q.FromCharacters(characters)
		s.ByGuild[guild] = q
	}

	for guild, query := range s.ByGuild {
		g := &Guild{
			Name: guild,
			RP:   query.TotalRP,
			XP:   query.TotalXP,
		}
		s.TopRPGuilds = append(s.TopRPGuilds, g)
		s.TopXPGuilds = append(s.TopXPGuilds, g)
		s.Guilds[guild] = g
	}

	sort.Sort(&s.TopRPGuilds)
	sort.Sort(&s.TopXPGuilds)

	return s
}
