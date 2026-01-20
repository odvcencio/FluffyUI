package main

// Game Data Types

type CandyType struct {
	Name     string
	MinPrice int
	MaxPrice int
	Emoji    string
}

type WeaponClass int

const (
	classNeutral WeaponClass = iota
	classBlunt
	classSharp
	classQuick
)

type SpecialItem struct {
	Name           string
	Description    string
	Class          WeaponClass
	UseInCombat    bool
	UseOutOfCombat bool
	Price          int
}

type Weapon struct {
	Name                string
	Class               WeaponClass
	AtkBonus            int
	Cost                int
	StunChance          int
	Tier                int
	RequiresBlackMarket bool
	Description         string
}

type Armor struct {
	Name                string
	DefBonus            int
	SPDMod              int
	Cost                int
	Tier                int
	RequiresBlackMarket bool
	Description         string
}

var CandyTypes = []CandyType{
	{Name: "Gummy Bears", MinPrice: 1, MaxPrice: 8, Emoji: "[G]"},
	{Name: "Chocolate Bar", MinPrice: 5, MaxPrice: 25, Emoji: "[C]"},
	{Name: "Sour Straws", MinPrice: 2, MaxPrice: 15, Emoji: "[S]"},
	{Name: "Lollipops", MinPrice: 1, MaxPrice: 6, Emoji: "[L]"},
	{Name: "Jawbreakers", MinPrice: 3, MaxPrice: 20, Emoji: "[J]"},
	{Name: "Rare Import", MinPrice: 20, MaxPrice: 150, Emoji: "[R]"},
}

var Weapons = []Weapon{
	{Name: "Fists", Class: classNeutral, AtkBonus: 0, Cost: 0, Tier: 0, Description: "Bare knuckles"},
	{Name: "Rubber Band", Class: classQuick, AtkBonus: 5, Cost: 20, Tier: 1, Description: "Snappy and fast"},
	{Name: "Dodgeball", Class: classBlunt, AtkBonus: 5, Cost: 20, Tier: 1, Description: "Bouncy and bruising"},
	{Name: "Textbook", Class: classBlunt, AtkBonus: 5, Cost: 20, Tier: 1, Description: "Solid and heavy"},
	{Name: "Compass", Class: classSharp, AtkBonus: 5, Cost: 20, Tier: 1, Description: "Pointed and precise"},
	{Name: "Jump Rope", Class: classQuick, AtkBonus: 12, Cost: 75, Tier: 2, Description: "Whips quick"},
	{Name: "Lunch Tray", Class: classBlunt, AtkBonus: 12, Cost: 75, Tier: 2, Description: "Wide, hard smack"},
	{Name: "Metal Ruler", Class: classSharp, AtkBonus: 12, Cost: 75, Tier: 2, Description: "Straight edge"},
	{
		Name:                "Turbo Jump Rope",
		Class:               classQuick,
		AtkBonus:            20,
		Cost:                200,
		Tier:                3,
		StunChance:          20,
		RequiresBlackMarket: true,
		Description:         "20% stun chance",
	},
	{
		Name:                "Championship Dodgeball",
		Class:               classBlunt,
		AtkBonus:            20,
		Cost:                200,
		Tier:                3,
		StunChance:          20,
		RequiresBlackMarket: true,
		Description:         "20% stun chance",
	},
	{
		Name:                "Precision Compass",
		Class:               classSharp,
		AtkBonus:            20,
		Cost:                200,
		Tier:                3,
		StunChance:          20,
		RequiresBlackMarket: true,
		Description:         "20% stun chance",
	},
}

var Armors = []Armor{
	{Name: "Hoodie", DefBonus: 0, SPDMod: 0, Cost: 0, Tier: 0, Description: "Everyday comfort"},
	{Name: "Denim Jacket", DefBonus: 5, SPDMod: 0, Cost: 30, Tier: 1, Description: "Light padding"},
	{Name: "Letterman Jacket", DefBonus: 12, SPDMod: -1, Cost: 80, Tier: 2, Description: "Sturdy but stiff"},
	{
		Name:                "Football Pads",
		DefBonus:            20,
		SPDMod:              -2,
		Cost:                180,
		Tier:                3,
		RequiresBlackMarket: true,
		Description:         "Heavy protection",
	},
}

var BlackMarketItems = []SpecialItem{
	{
		Name:           "Heat Sponge",
		Description:    "Reduce heat by 20",
		Class:          classNeutral,
		UseOutOfCombat: true,
		Price:          60,
	},
	{
		Name:           "Trade Voucher",
		Description:    "Next 2 trades: buy -10%, sell +10%",
		Class:          classNeutral,
		UseOutOfCombat: true,
		Price:          90,
	},
	{
		Name:        "Pepper Spray",
		Description: "Deal 18 damage and stun",
		Class:       classSharp,
		UseInCombat: true,
		Price:       70,
	},
}

const (
	startingCash     = 100
	startingDebt     = 500
	startingCapacity = 100
	maxDays          = 30
	hoursPerDay      = 8
	travelHours      = 2
	startingHP       = 100
	maxHP            = 100

	debtInterestRatePercent = 3
	bankInterestPercent     = 1
	bankLimitInitial        = 500
	bankLimitExpanded       = 1500
	combatHeatGain          = 4
	dayHeatDecay            = 5
	endDayHeatReduction     = 20
	statUpgradeCost         = 50
	statUpgradeMax          = 5
	muscleCost              = 150
	intimidationCost        = 200
	stashCost               = 150
	stashCapacity           = 50
	bikeCost                = 200
	informantCost           = 300
	bankExpansionCost       = 250
	blackMarketGearMarkup   = 125
	blackMarketRarePrice    = 120
	baseEventRiskFactor     = 5
	heatEventDivisor        = 6
	travelEventCap          = 55
	loiterEventCap          = 25

	tickSeconds         = 2
	loiterCheckSeconds  = 30
	ticksPerLoiterCheck = loiterCheckSeconds / tickSeconds

	locationCafeteria  = 0
	locationGymnasium  = 1
	locationLibrary    = 2
	locationPlayground = 3
	locationArtRoom    = 4
	locationMusicHall  = 5
)

type Location struct {
	Name        string
	Description string
	RiskLevel   int // 1-5, higher = more teacher patrols
}

var Locations = []Location{
	{Name: "Cafeteria", Description: "Busy during lunch - good prices, moderate risk", RiskLevel: 3},
	{Name: "Gymnasium", Description: "Athletes pay premium for energy", RiskLevel: 2},
	{Name: "Library", Description: "Quiet trades, but librarian watches closely", RiskLevel: 4},
	{Name: "Playground", Description: "High demand, but teachers patrol often", RiskLevel: 5},
	{Name: "Art Room", Description: "Creative kids, unpredictable prices", RiskLevel: 2},
	{Name: "Music Hall", Description: "Band kids have allowance money", RiskLevel: 1},
}

var LocationAdjacency = map[int][]int{
	locationCafeteria:  {locationGymnasium, locationPlayground},
	locationGymnasium:  {locationCafeteria, locationMusicHall},
	locationMusicHall:  {locationGymnasium, locationArtRoom},
	locationArtRoom:    {locationMusicHall, locationLibrary},
	locationLibrary:    {locationArtRoom, locationPlayground},
	locationPlayground: {locationLibrary, locationCafeteria},
}

type Inventory map[string]int

type MarketPrices map[string]int

type MarketStock map[string]int

type LoanTier struct {
	Name            string
	Amount          int
	InterestPercent int
	HeatPenalty     int
	ItemLoss        int
	GoonFight       bool
}

type Loan struct {
	Tier    int
	Balance int
}

var LoanTiers = []LoanTier{
	{Name: "Friendly", Amount: 100, InterestPercent: 5, HeatPenalty: 10, ItemLoss: 0, GoonFight: false},
	{Name: "Standard", Amount: 300, InterestPercent: 8, HeatPenalty: 20, ItemLoss: 1, GoonFight: false},
	{Name: "Desperate", Amount: 500, InterestPercent: 12, HeatPenalty: 30, ItemLoss: 0, GoonFight: true},
}

var backpackTierCosts = []int{100, 250, 500}
var backpackTierGains = []int{25, 50, 100}

type ScheduleStatus int

const (
	statusBlendingIn ScheduleStatus = iota
	statusWrongClass
	statusOffCampus
)

type Combatant struct {
	Name       string
	HP         int
	ATK        int
	DEF        int
	SPD        int
	Weapon     string
	WeaponType WeaponClass
	Armor      string
	WillFlee   bool
	FleeAtPct  int
	FleeChance int
	WeaponDrop int
	ArmorDrop  int
	CashMin    int
	CashMax    int
	CandyMin   int
	CandyMax   int
	RareChance float64
	IsGoon     bool
}

type CombatState struct {
	Enemy      Combatant
	EnemyHP    int
	Defending  bool
	Log        []string
	Shield     int
	SpeedBuff  int
	SpeedTurns int
	EnemyStun  int
}

type EventOption struct {
	Key    rune
	Label  string
	Action func()
}

var enemyBully = Combatant{
	Name:       "Playground Bully",
	HP:         30,
	ATK:        12,
	DEF:        5,
	SPD:        8,
	Weapon:     "Dodgeball",
	WeaponType: classBlunt,
	Armor:      "Denim Jacket",
	WillFlee:   false,
	FleeAtPct:  0,
	FleeChance: 0,
	WeaponDrop: 15,
	ArmorDrop:  10,
	CashMin:    10,
	CashMax:    30,
	CandyMin:   1,
	CandyMax:   3,
}

var enemyHallMonitor = Combatant{
	Name:       "Hall Monitor",
	HP:         35,
	ATK:        8,
	DEF:        12,
	SPD:        10,
	Weapon:     "Metal Ruler",
	WeaponType: classSharp,
	Armor:      "Letterman Jacket",
	WillFlee:   false,
	FleeAtPct:  0,
	FleeChance: 0,
	WeaponDrop: 20,
	ArmorDrop:  15,
	CashMin:    5,
	CashMax:    20,
	CandyMin:   0,
	CandyMax:   2,
}

var enemyGoon = Combatant{
	Name:       "Tony's Goon",
	HP:         45,
	ATK:        18,
	DEF:        10,
	SPD:        12,
	Weapon:     "Jump Rope",
	WeaponType: classQuick,
	Armor:      "Football Pads",
	WillFlee:   false,
	FleeAtPct:  0,
	FleeChance: 0,
	WeaponDrop: 35,
	ArmorDrop:  20,
	CashMin:    20,
	CashMax:    40,
	CandyMin:   0,
	CandyMax:   2,
	RareChance: 0.5,
	IsGoon:     true,
}

var enemyRivalTrader = Combatant{
	Name:       "Rival Trader",
	HP:         32,
	ATK:        10,
	DEF:        8,
	SPD:        15,
	Weapon:     "Rubber Band",
	WeaponType: classQuick,
	Armor:      "Denim Jacket",
	WillFlee:   true,
	FleeAtPct:  40,
	FleeChance: 55,
	WeaponDrop: 30,
	ArmorDrop:  35,
	CashMin:    15,
	CashMax:    55,
	CandyMin:   2,
	CandyMax:   5,
	RareChance: 0.2,
}

type CraftedItem struct {
	Name           string
	Ingredients    map[string]int
	Description    string
	Class          WeaponClass
	UseInCombat    bool
	UseOutOfCombat bool
}

var CraftedItems = []CraftedItem{
	{
		Name:        "Sugar Rush",
		Ingredients: map[string]int{"Gummy Bears": 2, "Sour Straws": 1},
		Description: "+15 SPD for 3 turns",
		Class:       classQuick,
		UseInCombat: true,
	},
	{
		Name:        "Jaw Lockdown",
		Ingredients: map[string]int{"Jawbreakers": 2},
		Description: "Enemy skips 1 turn",
		Class:       classBlunt,
		UseInCombat: true,
	},
	{
		Name:        "Chocolate Shield",
		Ingredients: map[string]int{"Chocolate Bar": 3},
		Description: "Block next 30 damage",
		Class:       classBlunt,
		UseInCombat: true,
	},
	{
		Name:        "Sour Bomb",
		Ingredients: map[string]int{"Sour Straws": 2, "Lollipops": 1},
		Description: "Deal 25 damage, ignores DEF",
		Class:       classSharp,
		UseInCombat: true,
	},
	{
		Name:           "Rare Remedy",
		Ingredients:    map[string]int{"Rare Import": 1, "Gummy Bears": 1},
		Description:    "Heal 50 HP",
		Class:          classNeutral,
		UseInCombat:    true,
		UseOutOfCombat: true,
	},
	{
		Name:           "Lucky Charm",
		Ingredients:    map[string]int{"Rare Import": 1, "Lollipops": 2},
		Description:    "Next trade: buy -10%, sell +10%",
		Class:          classNeutral,
		UseOutOfCombat: true,
	},
	{
		Name:        "Invisibility Pop",
		Ingredients: map[string]int{"Rare Import": 2},
		Description: "Auto-flee any combat",
		Class:       classNeutral,
		UseInCombat: true,
	},
}
