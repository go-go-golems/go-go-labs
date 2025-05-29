package modem

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// Client represents the modem HTTP client
type Client struct {
	httpClient *http.Client
	url        string
	headers    map[string]string
}

// NewClient creates a new modem client
func NewClient(url string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: url,
		headers: map[string]string{
			"User-Agent":                "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language":           "en-US,en;q=0.5",
			"Accept-Encoding":           "gzip, deflate",
			"Referer":                   "http://192.168.0.1/connection_status.jst",
			"Connection":                "keep-alive",
			"Cookie":                    "DUKSID=jst_sessYGEkaqyW7EEy5WDV6rqsmWA5X9u3I01O; csrfp_token=agtij4ybtx",
			"Upgrade-Insecure-Requests": "1",
			"Priority":                  "u=0, i",
		},
	}
}

// FetchModemInfo fetches and parses modem information
func (c *Client) FetchModemInfo(ctx context.Context) (*ModemInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Add headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTML")
	}

	return c.parseModemInfo(doc)
}

// parseModemInfo parses the HTML document and extracts modem information
func (c *Client) parseModemInfo(doc *goquery.Document) (*ModemInfo, error) {
	info := &ModemInfo{
		LastUpdated: time.Now(),
	}

	// Parse cable modem information
	cableModem, err := c.parseCableModem(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cable modem info")
	}
	info.CableModem = *cableModem

	// Parse downstream channels
	downstream, err := c.parseDownstreamChannels(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse downstream channels")
	}
	info.Downstream = downstream

	// Parse upstream channels
	upstream, err := c.parseUpstreamChannels(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse upstream channels")
	}
	info.Upstream = upstream

	// Parse error codewords
	errorCodewords, err := c.parseErrorCodewords(doc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse error codewords")
	}
	info.ErrorCodewords = errorCodewords

	return info, nil
}

// parseCableModem extracts cable modem hardware information
func (c *Client) parseCableModem(doc *goquery.Document) (*CableModem, error) {
	modem := &CableModem{}

	// Find the Cable Modem section
	doc.Find("div.module.forms").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Find("h2").Text(), "Cable Modem") {
			s.Find("div.form-row").Each(func(j int, row *goquery.Selection) {
				label := strings.TrimSpace(row.Find("span.readonlyLabel").Text())
				value := strings.TrimSpace(row.Find("span.value").Text())

				switch {
				case strings.Contains(label, "HW Version"):
					modem.HWVersion = value
				case strings.Contains(label, "Vendor"):
					modem.Vendor = value
				case strings.Contains(label, "BOOT Version"):
					modem.BOOTVersion = value
				case strings.Contains(label, "Core Version"):
					modem.CoreVersion = value
				case strings.Contains(label, "Model"):
					modem.Model = value
				case strings.Contains(label, "Product Type"):
					modem.ProductType = value
				case strings.Contains(label, "Flash Part"):
					modem.FlashPart = value
				case strings.Contains(label, "Download Version"):
					modem.DownloadVersion = value
				}
			})
		}
	})

	return modem, nil
}

// parseDownstreamChannels extracts downstream channel information
func (c *Client) parseDownstreamChannels(doc *goquery.Document) ([]Channel, error) {
	var channels []Channel

	// Find the downstream table
	doc.Find("div.module.netFlow").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Find("thead td").First().Text(), "Downstream") {
			channels = c.parseChannelTable(s, true)
		}
	})

	return channels, nil
}

// parseUpstreamChannels extracts upstream channel information
func (c *Client) parseUpstreamChannels(doc *goquery.Document) ([]Channel, error) {
	var channels []Channel

	// Find the upstream table
	doc.Find("div.module.netFlow").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Find("thead td").First().Text(), "Upstream") {
			channels = c.parseChannelTable(s, false)
		}
	})

	return channels, nil
}

// parseChannelTable parses a channel table (downstream or upstream)
func (c *Client) parseChannelTable(table *goquery.Selection, isDownstream bool) []Channel {
	var channels []Channel
	var channelIDs []string
	var lockStatuses []string
	var frequencies []string
	var snrs []string
	var powerLevels []string
	var modulations []string
	var symbolRates []string
	var channelTypes []string

	// Parse each row
	table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		rowLabel := strings.TrimSpace(row.Find("th.row-label").Text())
		
		var values []string
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			value := strings.TrimSpace(cell.Find("div.netWidth").Text())
			values = append(values, value)
		})

		switch {
		case strings.Contains(rowLabel, "Channel ID"):
			channelIDs = values
		case strings.Contains(rowLabel, "Lock Status"):
			lockStatuses = values
		case strings.Contains(rowLabel, "Frequency"):
			frequencies = values
		case strings.Contains(rowLabel, "SNR"):
			snrs = values
		case strings.Contains(rowLabel, "Power Level"):
			powerLevels = values
		case strings.Contains(rowLabel, "Modulation"):
			modulations = values
		case strings.Contains(rowLabel, "Symbol Rate"):
			symbolRates = values
		case strings.Contains(rowLabel, "Channel Type"):
			channelTypes = values
		}
	})

	// Build channel objects
	for i := 0; i < len(channelIDs); i++ {
		channel := Channel{
			ChannelID:  getValue(channelIDs, i),
			LockStatus: getValue(lockStatuses, i),
			Frequency:  getValue(frequencies, i),
			PowerLevel: getValue(powerLevels, i),
			Modulation: getValue(modulations, i),
		}

		if isDownstream {
			channel.SNR = getValue(snrs, i)
		} else {
			channel.SymbolRate = getValue(symbolRates, i)
			channel.ChannelType = getValue(channelTypes, i)
		}

		channels = append(channels, channel)
	}

	return channels
}

// parseErrorCodewords extracts error codeword information
func (c *Client) parseErrorCodewords(doc *goquery.Document) ([]ErrorChannel, error) {
	var errorChannels []ErrorChannel

	// Find the error codewords table
	doc.Find("div.module.netFlow").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Find("thead td").Text(), "CM Error Codewords") {
			var channelIDs []string
			var unerrored []string
			var correctable []string
			var uncorrectable []string

			s.Find("tbody tr").Each(func(j int, row *goquery.Selection) {
				rowLabel := strings.TrimSpace(row.Find("th.row-label").Text())
				
				var values []string
				row.Find("td").Each(func(k int, cell *goquery.Selection) {
					value := strings.TrimSpace(cell.Find("div.netWidth").Text())
					values = append(values, value)
				})

				switch {
				case strings.Contains(rowLabel, "Channel ID"):
					channelIDs = values
				case strings.Contains(rowLabel, "Unerrored Codewords"):
					unerrored = values
				case strings.Contains(rowLabel, "Correctable Codewords"):
					correctable = values
				case strings.Contains(rowLabel, "Uncorrectable Codewords"):
					uncorrectable = values
				}
			})

			// Build error channel objects
			for k := 0; k < len(channelIDs); k++ {
				errorChannel := ErrorChannel{
					ChannelID:              getValue(channelIDs, k),
					UnerroredCodewords:     getValue(unerrored, k),
					CorrectableCodewords:   getValue(correctable, k),
					UncorrectableCodewords: getValue(uncorrectable, k),
				}
				errorChannels = append(errorChannels, errorChannel)
			}
		}
	})

	return errorChannels, nil
}

// getValue safely gets a value from a slice at the given index
func getValue(slice []string, index int) string {
	if index < len(slice) {
		return slice[index]
	}
	return ""
} 