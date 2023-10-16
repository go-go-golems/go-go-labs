package machinery

import (
	"context"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/pkg"
	"net/http"
)

type HTTPServer struct {
	s        *http.Server
	Playlist *pkg.Playlist
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		s: &http.Server{
			Addr: ":8080",
		},
		Playlist: nil,
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		//fsys := os.DirFS(".")
		//f, err := fsys.Open(indexFileName)
		//if errors.Is(err, fs.ErrNotExist) {
		//	http.Error(w, "index.html not found", http.StatusNotFound)
		//	return
		//}
		//defer func(f fs.File) {
		//	_ = f.Close()
		//}(f)
		//
		//if s.Playlist == nil {
		//	http.Error(w, "playlist not found", http.StatusNotFound)
		//	return
		//}
		//rendered, err := s.Playlist.Render()
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//
		//fileTemplate, err := template.ParseFiles(indexFileName) // assuming index.html is template file
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
		//
		//// Execute the template with the rendered iframe strings
		//err = fileTemplate.Execute(w, rendered)
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
	}))
	return s.s.ListenAndServe()
}

func (s *HTTPServer) HandlePlaylist(playlist *pkg.Playlist) {
	s.Playlist = playlist
}
