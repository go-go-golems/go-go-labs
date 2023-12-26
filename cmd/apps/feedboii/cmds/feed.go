package cmds

import (
	"context"
	"fmt"
	map_pool "github.com/go-go-golems/clay/pkg/workerpool/map-pool"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type FeedCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FeedCommand)(nil)

func NewFeedCommand() (*FeedCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &FeedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"feed",
			cmds.WithShort("Fetch RSS feed and format as rows"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"feed-url",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"download",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"output-dir",
					parameters.ParameterTypeString,
					parameters.WithDefault("."),
				),
				parameters.NewParameterDefinition(
					"limit",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(0),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

type Download struct {
	URL          string
	FilePath     string
	PodcastTitle string
}

type FeedSettings struct {
	FeedUrls  []string `glazed.parameter:"feed-url"`
	Download  bool     `glazed.parameter:"download"`
	OutputDir string   `glazed.parameter:"output-dir"`
	Limit     int      `glazed.parameter:"limit"`
}

// TODO(manuel, 2023-10-12) Add bubbletea UI for download progress

func (f *FeedCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &FeedSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	fp := gofeed.NewParser()

	err = os.MkdirAll(s.OutputDir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "could not create output directory")
	}

	wg := sync.WaitGroup{}
	downloadPool := map_pool.New[*Download](5)

	if s.Download {
		downloadPool.Start()
		wg.Add(1)

		// start receiving results in order to unblock worker pool
		go func() {
			defer wg.Done()
			for download := range downloadPool.Results() {
				log.Info().
					Str("outputFile", download.FilePath).
					Str("URL", download.URL).
					Msg("Saved enclosure")
			}
		}()
	}

	currentJobs := map[string]interface{}{}
	var mu sync.Mutex

	for _, url := range s.FeedUrls {
		feed, err := fp.ParseURL(url)
		if err != nil {
			return errors.Wrapf(err, "Error fetching the RSS feed from URL: %s", url)
		}

		log.Info().Str("feedTitle", feed.Title).Msg("Processing feed")

		// Convert feed into structured data
		row := types.NewRow()
		row.Set("feedTitle", strings.TrimSpace(feed.Title))
		row.Set("feedLink", strings.TrimSpace(feed.Link))
		row.Set("feedDescription", strings.TrimSpace(feed.Description))

		for i, item := range feed.Items {
			if s.Limit > 0 && i >= s.Limit {
				break
			}

			log.Info().Str("feedTitle", feed.Title).Str("itemTitle", item.Title).Msg("Processing item")

			row.Set("title", strings.TrimSpace(item.Title))
			link := item.Link
			// look for the first enclosure
			for i, enclosure := range item.Enclosures {
				if enclosure.URL != "" {
					row.Set(fmt.Sprintf("enclosure%d", i), enclosure.URL)
				}
				outputFile := filepath.Join(s.OutputDir, makeOutputFileName(feed.Title, item.Title, enclosure.Type))
				// only download first enclosure, for now
				if i == 0 {

					row.Set(fmt.Sprintf("outputFile%d", i), outputFile)
					if s.Download {
						downloadPool.AddJob(func() (*Download, error) {
							// download enclosure.URL
							req, err := http.NewRequestWithContext(ctx, "GET", enclosure.URL, nil)
							if err != nil {
								return nil, err
							}

							// before download, mark as downloading
							mu.Lock()
							currentJobs[enclosure.URL] = true
							mu.Unlock()

							defer func() {
								mu.Lock()
								delete(currentJobs, enclosure.URL)
								mu.Unlock()
							}()

							log.Info().Str("URL", enclosure.URL).Msg("Downloading enclosure")
							// Execute the request
							resp, err := http.DefaultClient.Do(req)
							if err != nil {
								log.Error().Err(err).Str("URL", enclosure.URL).Msg("Error downloading enclosure")
								return nil, err
							}
							defer func(Body io.ReadCloser) {
								_ = Body.Close()
							}(resp.Body)

							// Check for cancellation
							select {
							case <-ctx.Done():
								log.Warn().Str("URL", enclosure.URL).Msg("Context cancelled")
								return nil, ctx.Err()
							default:
								// not cancelled, continue
							}

							log.Info().
								Str("outputFile", outputFile).
								Str("URL", enclosure.URL).
								Msg("Downloaded enclosure")

							// Create the output file
							out, err := os.Create(outputFile)
							if err != nil {
								return nil, err
							}
							defer func(out *os.File) {
								_ = out.Close()
							}(out)

							// Copy the response body to the output file
							_, err = io.Copy(out, resp.Body)
							if err != nil {
								return nil, err
							}

							log.Info().
								Str("outputFile", outputFile).
								Str("URL", enclosure.URL).
								Msg("Saved enclosure")

							return &Download{
								URL:          enclosure.URL,
								FilePath:     outputFile,
								PodcastTitle: item.Title,
							}, nil
						})
					}
				}
			}
			row.Set("link", strings.TrimSpace(link))
			row.Set("description", strings.TrimSpace(item.Description))

			err = gp.AddRow(ctx, row)
			if err != nil {
				return errors.Wrapf(err, "Error processing RSS feed from URL: %s", url)
			}
		}
	}

	downloadPool.Close()
	wg.Wait()

	return nil
}

// makeOutputFileName takes a title and returns a filename by removing special characters
// and replacing spaces with dashes and converting everything to lowercase.
func makeOutputFileName(feedTitle string, title string, contentType string) string {
	// use mp3 per default
	fileEnding := ".mp3"

	switch strings.ToLower(contentType) {
	case "audio/mpeg":
		fileEnding = ".mp3"
	case "audio/x-m4a":
		fileEnding = ".m4a"
	case "audio/mp4":
		fileEnding = ".mp4"
	case "audio/ogg":
		fileEnding = ".ogg"
	case "audio/wav":
		fileEnding = ".wav"
	case "audio/webm":
		fileEnding = ".webm"
	case "audio/flac":
		fileEnding = ".flac"
	case "audio/aac":
		fileEnding = ".aac"
	case "audio/aacp":
		fileEnding = ".aacp"
	case "audio/opus":
		fileEnding = ".opus"
	default:
		log.Warn().Str("content-type", contentType).Msg("Unknown content type, using mp3")
	}

	// remove special characters and spaces, then convert to lowercase
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to compile regex")
	}
	title = fmt.Sprintf("%s--%s", feedTitle, title)
	sanitizedTitle := reg.ReplaceAllString(title, "-")
	sanitizedTitle = strings.ToLower(sanitizedTitle)

	// replace consecutive dashes with single dash
	sanitizedTitle = regexp.MustCompile("-+").ReplaceAllString(sanitizedTitle, "-")

	// trim leading and trailing dashes
	sanitizedTitle = strings.Trim(sanitizedTitle, "-")

	// form the final filename
	fileName := sanitizedTitle + fileEnding

	return fileName
}
