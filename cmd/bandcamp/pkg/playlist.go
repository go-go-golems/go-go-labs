package pkg

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"text/template"
)

type Track struct {
	BackgroundColor string `json:"background_color"`
	LinkColor       string `json:"link_color"`
	AlbumID         int64  `json:"album_id"`
	Name            string `json:"name"`
	BandName        string `json:"band_name"`
	ItemURLPath     string `json:"item_url_path"`
}

type Playlist struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tracks      []*Track `json:"tracks"`
}

const iframeTmpl = `<iframe
   style="border: 0; width: 100%%; height: 42px; background-color: {{.BackgroundColor}};" 
   src="https://bandcamp.com/EmbeddedPlayer/album={{.AlbumID}}/size=small/bgcol={{.BackgroundColor}}/linkcol={{.LinkColor}}/transparent=true/" seamless>
     <a href="https://bandcamp.com/{{.ItemURLPath}}">{{.Name}} by {{.BandName}}</a>
</iframe>`

func (p *Playlist) Render() (string, error) {
	var out bytes.Buffer

	tmpl, err := template.New("playlist").Parse(iframeTmpl)
	if err != nil {
		return "", err
	}

	for _, track := range p.Tracks {

		// A struct that fulfills the template
		// Apply the data to the template
		if err := tmpl.Execute(&out, track); err != nil {
			return "", err
		}

		// Add a separator between tracks
		if err := out.WriteByte('\n'); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func LoadFromFile(filename string) (*Playlist, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err

	}

	p := &Playlist{}
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Playlist) SaveToFile(filename string) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, b, 0644)
}

// MoveEntryUp moves an entry up in the playlist and returns the new index
func (p *Playlist) MoveEntryUp(index int) int {
	if index == 0 {
		return 0
	}

	p.Tracks[index], p.Tracks[index-1] = p.Tracks[index-1], p.Tracks[index]
	return index - 1
}

func (p *Playlist) MoveEntryDown(index int) int {
	if index == len(p.Tracks)-1 {
		return index
	}

	p.Tracks[index], p.Tracks[index+1] = p.Tracks[index+1], p.Tracks[index]
	return index + 1
}

func (p *Playlist) DeleteEntry(index int) {
	p.Tracks = append(p.Tracks[:index], p.Tracks[index+1:]...)
}

func (p *Playlist) InsertTrack(track *Track, index int) {
	p.Tracks = append(p.Tracks[:index], append([]*Track{track}, p.Tracks[index:]...)...)

}

func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // for linux and unix
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
