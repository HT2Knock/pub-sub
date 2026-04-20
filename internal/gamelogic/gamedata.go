package gamelogic

type Player struct {
	Units    map[int]Unit
	Username string
}

type UnitRank string

const (
	RankInfantry  = "infantry"
	RankCavalry   = "cavalry"
	RankArtillery = "artillery"
)

type Unit struct {
	Rank     UnitRank
	Location Location
	ID       int
}

type ArmyMove struct {
	Player     Player
	ToLocation Location
	Units      []Unit
}

type RecognitionOfWar struct {
	Attacker Player
	Defender Player
}

type Location string

func getAllRanks() map[UnitRank]struct{} {
	return map[UnitRank]struct{}{
		RankInfantry:  {},
		RankCavalry:   {},
		RankArtillery: {},
	}
}

func getAllLocations() map[Location]struct{} {
	return map[Location]struct{}{
		"americas":   {},
		"europe":     {},
		"africa":     {},
		"asia":       {},
		"australia":  {},
		"antarctica": {},
	}
}
