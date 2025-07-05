package types

// NewFilmDatabase creates a new film database with all the film data
func NewFilmDatabase() *FilmDatabase {
	return &FilmDatabase{
		Films: map[string]Film{
			"hp5_plus": {
				ID:        "hp5_plus",
				Name:      "HP5 PLUS",
				EIRatings: []int{200, 400, 800},
				Times20C: map[string]map[int]string{
					"1+9":  {200: "5:00", 400: "6:30", 800: "13:30"},
					"1+14": {200: "7:00", 400: "11:00", 800: "19:30"},
				},
				Description: "Most popular",
				Icon:        "📈",
			},
			"fp4_plus": {
				ID:        "fp4_plus",
				Name:      "FP4 PLUS",
				EIRatings: []int{125},
				Times20C: map[string]map[int]string{
					"1+9":  {125: "4:15"},
					"1+14": {125: "7:30"},
				},
				Description: "Fine grain",
				Icon:        "🎯",
			},
			"delta_100": {
				ID:        "delta_100",
				Name:      "DELTA 100",
				EIRatings: []int{100},
				Times20C: map[string]map[int]string{
					"1+9":  {100: "5:00"},
					"1+14": {100: "7:30"},
				},
				Description: "Ultra fine",
				Icon:        "🔍",
			},
			"delta_400": {
				ID:        "delta_400",
				Name:      "DELTA 400",
				EIRatings: []int{200, 400, 800},
				Times20C: map[string]map[int]string{
					"1+9":  {200: "5:30", 400: "7:00", 800: "14:00"},
					"1+14": {200: "8:00", 400: "12:00", 800: "20:30"},
				},
				Description: "Versatile",
				Icon:        "⚖️",
			},
			"delta_3200": {
				ID:        "delta_3200",
				Name:      "DELTA 3200",
				EIRatings: []int{400, 800, 1600, 3200, 6400},
				Times20C: map[string]map[int]string{
					"1+9":  {400: "6:00", 800: "7:30", 1600: "10:00", 3200: "11:00", 6400: "18:00"},
					"1+14": {400: "11:00", 800: "13:00", 1600: "15:30", 3200: "17:00", 6400: "23:00"},
				},
				Description: "High speed",
				Icon:        "🌙",
			},
			"pan_f_plus": {
				ID:        "pan_f_plus",
				Name:      "PAN F PLUS",
				EIRatings: []int{50},
				Times20C: map[string]map[int]string{
					"1+14": {50: "4:30"},
				},
				Description: "Finest grain",
				Icon:        "💎",
			},
			"sfx_200": {
				ID:        "sfx_200",
				Name:      "SFX 200",
				EIRatings: []int{200, 400},
				Times20C: map[string]map[int]string{
					"1+9":  {200: "6:00", 400: "8:30"},
					"1+14": {200: "9:00", 400: "13:30"},
				},
				Description: "Infrared",
				Icon:        "🔴",
			},
		},
	}
}

// NewTankDatabase creates a new tank database with tank size information
func NewTankDatabase() *TankDatabase {
	return &TankDatabase{
		Sizes: map[string]map[int]int{
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
		},
	}
}

// NewChemicalDatabase creates a new chemical database
func NewChemicalDatabase() *ChemicalDatabase {
	return &ChemicalDatabase{
		Chemicals: map[string]Chemical{
			"ilfosol_3": {
				Name:      "ILFOSOL 3",
				Dilutions: []string{"1+9", "1+14"},
				Default:   "1+9",
				Time:      "",
				Type:      "one_shot",
				Capacity:  "",
			},
			"ilfostop": {
				Name:      "ILFOSTOP",
				Dilutions: []string{"1+19"},
				Default:   "1+19",
				Time:      "0:10",
				Type:      "reusable",
				Capacity:  "15 rolls per liter",
			},
			"sprint_fixer": {
				Name:      "SPRINT FIXER",
				Dilutions: []string{"1+4"},
				Default:   "1+4",
				Time:      "2:30",
				Type:      "reusable",
				Capacity:  "24 rolls per liter",
			},
		},
	}
}

// GetFilmOrder returns the films in the order they should be displayed
func GetFilmOrder() []string {
	return []string{
		"hp5_plus",
		"fp4_plus",
		"delta_100",
		"delta_400",
		"delta_3200",
		"pan_f_plus",
		"sfx_200",
	}
} 