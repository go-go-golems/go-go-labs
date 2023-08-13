package main_ui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui/playlist"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui/search"
)

type Model struct {
	Search   search.Model
	Playlist playlist.Model
	client   *pkg.Client
}

func NewModel(client *pkg.Client) Model {
	searchModel := search.NewModel(client, nil)
	res := Model{
		client: client,
	}
	searchModel.OnSearchCmd = func(searchTerm string) tea.Cmd {
		return func() tea.Msg {
			return res.SearchBandcamp(searchTerm)
		}
	}
	searchModel.OnSelectEntryCmd = func(result *pkg.Result) tea.Cmd {
		return func() tea.Msg {
			return ui.SelectEntryMsg{Result: result}
		}
	}
	res.Search = searchModel
	return res
}

func (m Model) SearchBandcamp(searchTerm string) tea.Msg {
	resp, err := m.client.Search(context.Background(), searchTerm, pkg.FilterTrack)
	if err != nil {
		return ui.ErrMsg{Err: err}
	}

	return ui.UpdateSearchResultsMsg{Results: resp.Auto.Results}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return m.SearchBandcamp("slono")
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		}

	}

	cmds := make([]tea.Cmd, 0)

	searchModel, cmd := m.Search.Update(msg)
	m.Search = searchModel.(search.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.Search.View()
}
