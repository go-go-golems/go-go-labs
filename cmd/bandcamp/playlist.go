package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

type PlaylistSection struct {
	LinkColor       string    `json:"linkColor"`
	BackgroundColor string    `json:"backgroundColor"`
	Tracks          []*Result `json:"tracks"`
}

type Playlist struct {
	Sections []*PlaylistSection `json:"sections"`
}

const iframeTmpl = `<iframe
   style="border: 0; width: 100%%; height: 42px; background-color: {{.BackgroundColor}};" 
   src="https://bandcamp.com/EmbeddedPlayer/album={{.Track.AlbumID}}/size=small/bgcol={{.BackgroundColor}}/linkcol={{.LinkColor}}/transparent=true/" seamless>
     <a href="https://bandcamp.com/{{.Track.ItemURLPath}}">{{.Track.Name}} by {{.Track.BandName}}</a>
</iframe>`

func (p *PlaylistSection) Render() (string, error) {
	var out bytes.Buffer

	tmpl, err := template.New("playlist").Parse(iframeTmpl)
	if err != nil {
		return "", err
	}

	for _, track := range p.Tracks {

		// A struct that fulfills the template
		data := struct {
			BackgroundColor string
			LinkColor       string
			Track           *Result
		}{
			BackgroundColor: p.BackgroundColor,
			LinkColor:       p.LinkColor,
			Track:           track,
		}

		// Apply the data to the template
		if err := tmpl.Execute(&out, data); err != nil {
			return "", err
		}

		// Add a separator between tracks
		if err := out.WriteByte('\n'); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func generatePlaylists(jsonData []byte) ([]string, error) {
	var playlists []PlaylistSection
	err := json.Unmarshal(jsonData, &playlists)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, playlist := range playlists {
		for _, track := range playlist.Tracks {
			iframe := fmt.Sprintf(`<iframe style="border: 0; width: 100%%; height: 42px; background-color: %s;" src="https://bandcamp.com/EmbeddedPlayer/album=%d/size=small/bgcol=%s/linkcol=%s/transparent=true/" seamless><a href="https://bandcamp.com/%s">%s by %s</a></iframe>`,
				playlist.BackgroundColor,
				track.AlbumID,
				playlist.BackgroundColor,
				playlist.LinkColor,
				track.ItemURLPath,
				track.AlbumName,
				track.BandName)
			result = append(result, iframe)
		}
	}
	return result, nil
}
