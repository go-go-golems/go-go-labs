package types

// Film represents a film type with its properties
type Film struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	EIRatings   []int                     `json:"ei_ratings"`
	Times20C    map[string]map[int]string `json:"times_20c"`
	Description string                    `json:"description"`
	Icon        string                    `json:"icon"`
}

// FilmDatabase contains all available films
type FilmDatabase struct {
	Films map[string]Film `json:"films"`
}

// GetFilms returns all films as a slice
func (fd *FilmDatabase) GetFilms() []Film {
	films := make([]Film, 0, len(fd.Films))
	for _, film := range fd.Films {
		films = append(films, film)
	}
	return films
}

// GetFilmByID returns a film by its ID
func (fd *FilmDatabase) GetFilmByID(id string) (Film, bool) {
	film, ok := fd.Films[id]
	return film, ok
}
