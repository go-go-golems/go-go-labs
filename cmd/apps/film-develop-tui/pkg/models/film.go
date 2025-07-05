package models

import "time"

// Film represents a film type with its characteristics
type Film struct {
	ID          string
	Name        string
	EIRatings   []int
	Times20C    map[string]map[int]string // dilution -> ei -> time
}

// FilmDatabase contains all supported films
var FilmDatabase = map[string]Film{
	"hp5_plus": {
		ID:        "hp5_plus",
		Name:      "HP5 PLUS",
		EIRatings: []int{200, 400, 800},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				200: "5:00",
				400: "6:30",
				800: "13:30",
			},
			"1_plus_14": {
				200: "7:00",
				400: "11:00",
				800: "19:30",
			},
		},
	},
	"fp4_plus": {
		ID:        "fp4_plus",
		Name:      "FP4 PLUS",
		EIRatings: []int{125},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				125: "4:15",
			},
			"1_plus_14": {
				125: "7:30",
			},
		},
	},
	"delta_100": {
		ID:        "delta_100",
		Name:      "DELTA 100",
		EIRatings: []int{100},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				100: "5:00",
			},
			"1_plus_14": {
				100: "7:30",
			},
		},
	},
	"delta_400": {
		ID:        "delta_400",
		Name:      "DELTA 400",
		EIRatings: []int{200, 400, 800},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				200: "5:30",
				400: "7:00",
				800: "14:00",
			},
			"1_plus_14": {
				200: "8:00",
				400: "12:00",
				800: "20:30",
			},
		},
	},
	"delta_3200": {
		ID:        "delta_3200",
		Name:      "DELTA 3200",
		EIRatings: []int{400, 800, 1600, 3200, 6400},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				400:  "6:00",
				800:  "7:30",
				1600: "10:00",
				3200: "11:00",
				6400: "18:00",
			},
			"1_plus_14": {
				400:  "11:00",
				800:  "13:00",
				1600: "15:30",
				3200: "17:00",
				6400: "23:00",
			},
		},
	},
	"pan_f_plus": {
		ID:        "pan_f_plus",
		Name:      "PAN F PLUS",
		EIRatings: []int{50},
		Times20C: map[string]map[int]string{
			"1_plus_14": {
				50: "4:30",
			},
		},
	},
	"sfx_200": {
		ID:        "sfx_200",
		Name:      "SFX 200",
		EIRatings: []int{200, 400},
		Times20C: map[string]map[int]string{
			"1_plus_9": {
				200: "6:00",
				400: "8:30",
			},
			"1_plus_14": {
				200: "9:00",
				400: "13:30",
			},
		},
	},
}

// GetFilmOptions returns a slice of films in display order
func GetFilmOptions() []Film {
	order := []string{"hp5_plus", "fp4_plus", "delta_100", "delta_400", "delta_3200", "pan_f_plus", "sfx_200"}
	films := make([]Film, 0, len(order))
	for _, id := range order {
		films = append(films, FilmDatabase[id])
	}
	return films
}

// GetDevelopmentTime returns the development time for a film/ei/dilution combination
func (f *Film) GetDevelopmentTime(ei int, dilution string) (time.Duration, error) {
	timeStr, ok := f.Times20C[dilution][ei]
	if !ok {
		return 0, ErrInvalidCombination
	}
	
	duration, err := time.ParseDuration(timeStr)
	if err != nil {
		return 0, err
	}
	
	return duration, nil
}

// HasEI checks if the film supports a specific EI rating
func (f *Film) HasEI(ei int) bool {
	for _, rating := range f.EIRatings {
		if rating == ei {
			return true
		}
	}
	return false
}

// HasDilution checks if the film supports a specific dilution
func (f *Film) HasDilution(dilution string) bool {
	_, ok := f.Times20C[dilution]
	return ok
}
