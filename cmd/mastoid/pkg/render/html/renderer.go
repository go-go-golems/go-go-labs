package html

import (
	"fmt"
	"github.com/mattn/go-mastodon"
	"io"
)

type Renderer struct {
	Verbose    bool
	WithHeader bool
}

type RenderOption func(*Renderer)

func WithHeader(withHeader bool) RenderOption {
	return func(r *Renderer) {
		r.WithHeader = withHeader
	}
}

func WithVerbose(verbose bool) RenderOption {
	return func(r *Renderer) {
		r.Verbose = verbose
	}
}

func NewRenderer(options ...RenderOption) *Renderer {
	ret := &Renderer{}
	for _, option := range options {
		option(ret)
	}
	return ret
}

func (r *Renderer) RenderStatus(w io.Writer, status *mastodon.Status) error {
	var content string
	if r.Verbose {
		content = fmt.Sprintf("Status ID: %s<br>Created at: %v<br>Content: %s<br>-----------------<br>",
			status.ID, status.CreatedAt, status.Content)
	} else {
		content = fmt.Sprintf("%s<br>", status.Content)
	}
	_, err := w.Write([]byte(content))
	return err
}

func (r *Renderer) RenderThread(w io.Writer, status *mastodon.Status, context *mastodon.Context) error {
	for _, ancestor := range context.Ancestors {
		if err := r.RenderStatus(w, ancestor); err != nil {
			return err
		}
	}

	if err := r.RenderStatus(w, status); err != nil {
		return err
	}

	for _, descendant := range context.Descendants {
		if err := r.RenderStatus(w, descendant); err != nil {
			return err
		}
	}

	return nil
}
