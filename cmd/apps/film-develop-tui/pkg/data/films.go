package data

import "time"

// Film represents a film type with its development characteristics
type Film struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	EIRatings   []int             `json:"ei_ratings"`
	Times20C    map[string]Times  `json:"times_20c"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
}

// Times represents development times for different dilutions
type Times map[int]time.Duration

// Chemical represents a chemical with its properties
type Chemical struct {
	Name        string            `json:"name"`
	Dilutions   []string          `json:"dilutions"`
	Default     string            `json:"default"`
	Type        string            `json:"type"`
	Time        time.Duration     `json:"time"`
	Capacity    string            `json:"capacity"`
	RollsPerL   int               `json:"rolls_per_liter"`
}

// TankSize represents tank capacity requirements
type TankSize struct {
	Format string `json:"format"`
	Rolls  int    `json:"rolls"`
	Volume int    `json:"volume_ml"`
}

// FilmDatabase holds all film data
var FilmDatabase = map[string]Film{
	"hp5_plus": {
		ID:          "hp5_plus",
		Name:        "HP5 PLUS",
		EIRatings:   []int{200, 400, 800},
		Description: "Most popular",
		Icon:        "📈",
		Times20C: map[string]Times{
			"1+9": {
				200: 5*time.Minute,
				400: 6*time.Minute + 30*time.Second,
				800: 13*time.Minute + 30*time.Second,
			},
			"1+14": {
				200: 7*time.Minute,
				400: 11*time.Minute,
				800: 19*time.Minute + 30*time.Second,
			},
		},
	},
	"fp4_plus": {
		ID:          "fp4_plus",
		Name:        "FP4 PLUS",
		EIRatings:   []int{125},
		Description: "Fine grain",
		Icon:        "🎯",
		Times20C: map[string]Times{
			"1+9": {
				125: 4*time.Minute + 15*time.Second,
			},
			"1+14": {
				125: 7*time.Minute + 30*time.Second,
			},
		},
	},
	"delta_100": {
		ID:          "delta_100",
		Name:        "DELTA 100",
		EIRatings:   []int{100},
		Description: "Ultra fine",
		Icon:        "🔍",
		Times20C: map[string]Times{
			"1+9": {
				100: 5*time.Minute,
			},
			"1+14": {
				100: 7*time.Minute + 30*time.Second,
			},
		},
	},
	"delta_400": {
		ID:          "delta_400",
		Name:        "DELTA 400",
		EIRatings:   []int{200, 400, 800},
		Description: "Versatile",
		Icon:        "⚖️",
		Times20C: map[string]Times{
			"1+9": {
				200: 5*time.Minute + 30*time.Second,
				400: 7*time.Minute,
				800: 14*time.Minute,
			},
			"1+14": {
				200: 8*time.Minute,
				400: 12*time.Minute,
				800: 20*time.Minute + 30*time.Second,
			},
		},
	},
	"delta_3200": {
		ID:          "delta_3200",
		Name:        "DELTA 3200",
		EIRatings:   []int{400, 800, 1600, 3200, 6400},
		Description: "High speed",
		Icon:        "🌙",
		Times20C: map[string]Times{
			"1+9": {
				400:  6*time.Minute,
				800:  7*time.Minute + 30*time.Second,
				1600: 10*time.Minute,
				3200: 11*time.Minute,
				6400: 18*time.Minute,
			},
			"1+14": {
				400:  11*time.Minute,
				800:  13*time.Minute,
				1600: 15*time.Minute + 30*time.Second,
				3200: 17*time.Minute,
				6400: 23*time.Minute,
			},
		},
	},
	"pan_f_plus": {
		ID:          "pan_f_plus",
		Name:        "PAN F PLUS",
		EIRatings:   []int{50},
		Description: "Finest grain",
		Icon:        "💎",
		Times20C: map[string]Times{
			"1+14": {
				50: 4*time.Minute + 30*time.Second,
			},
		},
	},
	"sfx_200": {
		ID:          "sfx_200",
		Name:        "SFX 200",
		EIRatings:   []int{200, 400},
		Description: "Infrared",
		Icon:        "🔴",
		Times20C: map[string]Times{
			"1+9": {
				200: 6*time.Minute,
				400: 8*time.Minute + 30*time.Second,
			},
			"1+14": {
				200: 9*time.Minute,
				400: 13*time.Minute + 30*time.Second,
			},
		},
	},
}

// ChemicalDatabase holds all chemical data
var ChemicalDatabase = map[string]Chemical{
	"ilfosol_3": {
		Name:      "ILFOSOL 3",
		Dilutions: []string{"1+9", "1+14"},
		Default:   "1+9",
		Type:      "one_shot",
	},
	"ilfostop": {
		Name:      "ILFOSTOP",
		Dilutions: []string{"1+19"},
		Default:   "1+19",
		Type:      "reusable",
		Time:      10 * time.Second,
		Capacity:  "15 rolls per liter",
		RollsPerL: 15,
	},
	"sprint_fixer": {
		Name:      "SPRINT FIXER",
		Dilutions: []string{"1+4"},
		Default:   "1+4",
		Type:      "reusable",
		Time:      2*time.Minute + 30*time.Second,
		Capacity:  "24 rolls per liter",
		RollsPerL: 24,
	},
}

// TankSizes defines tank capacities for different formats and roll counts
var TankSizes = map[string]map[int]int{
	"35mm": {
		1: 300,
		2: 500,
		3: 600,
		4: 700,
		5: 800,
		6: 900,
	},
	"120mm": {
		1: 500,
		2: 700,
		3: 900,
		4: 1000,
		5: 1200,
		6: 1400,
	},
}

// GetFilmList returns an ordered list of films
func GetFilmList() []Film {
	order := []string{"hp5_plus", "fp4_plus", "delta_100", "delta_400", "delta_3200", "pan_f_plus", "sfx_200"}
	var films []Film
	for _, id := range order {
		if film, exists := FilmDatabase[id]; exists {
			films = append(films, film)
		}
	}
	return films
}

// GetTankSize returns the tank size for given format and roll count
func GetTankSize(format string, rolls int) int {
	if sizes, exists := TankSizes[format]; exists {
		if size, exists := sizes[rolls]; exists {
			return size
		}
	}
	return 0
}

// CalculateCustomTankSize calculates tank size for mixed roll counts
func CalculateCustomTankSize(rolls35mm, rolls120mm int) int {
	size35 := GetTankSize("35mm", rolls35mm)
	size120 := GetTankSize("120mm", rolls120mm)
	
	// For mixed batches, we need to calculate based on the larger requirement
	if rolls35mm > 0 && rolls120mm > 0 {
		// Use the higher of the two calculations
		if size35 > size120 {
			return size35
		}
		return size120
	} else if rolls35mm > 0 {
		return size35
	} else if rolls120mm > 0 {
		return size120
	}
	return 0
}
