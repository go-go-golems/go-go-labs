package modem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// StoredCookie represents a cookie that can be serialized to JSON
type StoredCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"httpOnly"`
}

// Client represents the modem HTTP client
type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    map[string]string
	cookies    []*http.Cookie
	username   string
	password   string
}

// NewClient creates a new modem client
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects automatically
				return http.ErrUseLastResponse
			},
		},
		baseURL: baseURL,
		headers: map[string]string{
			"User-Agent":                "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			"Accept-Language":           "en-US,en;q=0.5",
			"Accept-Encoding":           "gzip, deflate",
			"Connection":                "keep-alive",
			"Upgrade-Insecure-Requests": "1",
			"Priority":                  "u=0, i",
		},
	}
}

// SetCredentials sets the username and password for authentication
func (c *Client) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

// getCookieFilePath returns the path to the cookie storage file
func (c *Client) getCookieFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}
	
	configDir := filepath.Join(homeDir, ".config", "poll-modem")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create config directory")
	}
	
	return filepath.Join(configDir, "cookies.json"), nil
}

// saveCookies saves the current cookies to the config file
func (c *Client) saveCookies() error {
	cookieFile, err := c.getCookieFilePath()
	if err != nil {
		return err
	}
	
	var storedCookies []StoredCookie
	for _, cookie := range c.cookies {
		storedCookies = append(storedCookies, StoredCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
		})
	}
	
	data, err := json.MarshalIndent(storedCookies, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal cookies")
	}
	
	if err := os.WriteFile(cookieFile, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write cookie file")
	}
	
	log.Debug().Str("file", cookieFile).Int("cookies", len(storedCookies)).Msg("Saved cookies to file")
	return nil
}

// loadCookies loads cookies from the config file
func (c *Client) loadCookies() error {
	cookieFile, err := c.getCookieFilePath()
	if err != nil {
		return err
	}
	
	// If file doesn't exist, that's okay - no cookies to load
	if _, err := os.Stat(cookieFile); os.IsNotExist(err) {
		log.Debug().Str("file", cookieFile).Msg("Cookie file does not exist, starting with no cookies")
		return nil
	}
	
	data, err := os.ReadFile(cookieFile)
	if err != nil {
		return errors.Wrap(err, "failed to read cookie file")
	}
	
	var storedCookies []StoredCookie
	if err := json.Unmarshal(data, &storedCookies); err != nil {
		return errors.Wrap(err, "failed to unmarshal cookies")
	}
	
	// Convert stored cookies back to http.Cookie and filter out expired ones
	c.cookies = nil
	now := time.Now()
	for _, stored := range storedCookies {
		// Skip expired cookies
		if !stored.Expires.IsZero() && stored.Expires.Before(now) {
			log.Debug().Str("cookie", stored.Name).Time("expired", stored.Expires).Msg("Skipping expired cookie")
			continue
		}
		
		cookie := &http.Cookie{
			Name:     stored.Name,
			Value:    stored.Value,
			Domain:   stored.Domain,
			Path:     stored.Path,
			Expires:  stored.Expires,
			Secure:   stored.Secure,
			HttpOnly: stored.HttpOnly,
		}
		c.cookies = append(c.cookies, cookie)
	}
	
	log.Debug().Str("file", cookieFile).Int("cookies", len(c.cookies)).Msg("Loaded cookies from file")
	return nil
}

// mergeCookies adds new cookies to the existing cookie jar, replacing any existing cookies with the same name
func (c *Client) mergeCookies(newCookies []*http.Cookie) {
	for _, newCookie := range newCookies {
		// Find and replace existing cookie with same name, or append if not found
		found := false
		for i, existingCookie := range c.cookies {
			if existingCookie.Name == newCookie.Name {
				c.cookies[i] = newCookie
				found = true
				break
			}
		}
		if !found {
			c.cookies = append(c.cookies, newCookie)
		}
	}
}

// clearCookies removes all cookies from the jar and optionally deletes the cookie file
func (c *Client) clearCookies(deleteFile bool) error {
	c.cookies = nil
	log.Debug().Msg("Cleared all cookies from memory")
	
	if deleteFile {
		cookieFile, err := c.getCookieFilePath()
		if err != nil {
			return err
		}
		
		if err := os.Remove(cookieFile); err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to delete cookie file")
		}
		
		log.Debug().Str("file", cookieFile).Msg("Deleted cookie file")
	}
	
	return nil
}

// Logout clears the session and removes stored cookies
func (c *Client) Logout() error {
	return c.clearCookies(true)
}

// Login performs authentication with the modem
func (c *Client) Login(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return errors.New("username and password must be set before login")
	}

	// Clear any existing cookies before login
	if err := c.clearCookies(false); err != nil {
		log.Debug().Err(err).Msg("Failed to clear existing cookies")
		// Continue anyway
	}

	// Prepare login data
	data := url.Values{}
	data.Set("locale", "false")
	data.Set("username", c.username)
	data.Set("password", c.password)
	log.Debug().Msgf("Login data: %v", data.Encode())

	// Create login request
	loginURL := c.baseURL + "/check.jst"
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return errors.Wrap(err, "failed to create login request")
	}

	// Set headers for login to match the sniffed request exactly
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.baseURL+"/index.jst")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Origin", c.baseURL)
	
	// Set the standard headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Note: Based on sniffed request, NO cookies are sent with login request

	// Debug logging for login request
	log.Debug().
		Str("method", req.Method).
		Str("url", loginURL).
		Str("username", c.username).
		Str("content_type", req.Header.Get("Content-Type")).
		Str("cookie_header", req.Header.Get("Cookie")).
		Str("origin", req.Header.Get("Origin")).
		Str("referer", req.Header.Get("Referer")).
		Msg("Sending login request")

	// Perform login
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("Login request failed")
		return errors.Wrap(err, "failed to perform login request")
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	// Store response body to file for debugging
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("/tmp/body.%s.html", timestamp)
	if err := os.WriteFile(filename, bodyBytes, 0644); err != nil {
		log.Debug().Err(err).Str("filename", filename).Msg("Failed to write response body to file")
	} else {
		log.Debug().Str("filename", filename).Msg("Response body saved to file")
	}

	// Debug logging for response headers
	log.Debug().
		Interface("response_headers", func() map[string][]string {
			headers := make(map[string][]string)
			for key, values := range resp.Header {
				headers[key] = values
			}
			return headers
		}()).
		Msg("Response headers received")

	// Debug logging for login response
	log.Debug().
		Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Int("body_length", len(bodyString)).
		Int("cookies_count", len(resp.Cookies())).
		Interface("cookies_received", func() []map[string]interface{} {
			var cookies []map[string]interface{}
			for _, cookie := range resp.Cookies() {
				cookies = append(cookies, map[string]interface{}{
					"name":     cookie.Name,
					"value":    cookie.Value,
					"domain":   cookie.Domain,
					"path":     cookie.Path,
					"expires":  cookie.Expires,
					"secure":   cookie.Secure,
					"httpOnly": cookie.HttpOnly,
				})
			}
			return cookies
		}()).
		Msg("Received login response")

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		log.Debug().
			Int("status_code", resp.StatusCode).
			Str("body", bodyString).
			Msg("Login failed with unexpected status")
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// Store cookies from login response
	c.mergeCookies(resp.Cookies())
	
	// If we got a 302 redirect, follow it manually to get additional cookies
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location != "" {
			log.Debug().Str("location", location).Msg("Following 302 redirect manually")
			
			// Build full URL for redirect
			var redirectURL string
			if strings.HasPrefix(location, "http") {
				redirectURL = location
			} else {
				redirectURL = c.baseURL + "/" + strings.TrimPrefix(location, "/")
			}
			
			// Create redirect request
			redirectReq, err := http.NewRequestWithContext(ctx, "GET", redirectURL, nil)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to create redirect request")
				return errors.Wrap(err, "failed to create redirect request")
			}
			
			// Set headers for redirect request
			for key, value := range c.headers {
				redirectReq.Header.Set(key, value)
			}
			redirectReq.Header.Set("Referer", loginURL)
			
			// Add cookies from login response to redirect request
			for _, cookie := range c.cookies {
				redirectReq.AddCookie(cookie)
			}
			
			// Perform redirect request
			redirectResp, err := c.httpClient.Do(redirectReq)
			if err != nil {
				log.Debug().Err(err).Msg("Redirect request failed")
				return errors.Wrap(err, "failed to perform redirect request")
			}
			defer redirectResp.Body.Close()
			
			// Read redirect response body
			redirectBodyBytes, _ := io.ReadAll(redirectResp.Body)
			redirectBodyString := string(redirectBodyBytes)
			
			log.Debug().
				Int("redirect_status_code", redirectResp.StatusCode).
				Str("redirect_status", redirectResp.Status).
				Int("redirect_body_length", len(redirectBodyString)).
				Int("redirect_cookies_count", len(redirectResp.Cookies())).
				Interface("redirect_cookies", func() []map[string]interface{} {
					var cookies []map[string]interface{}
					for _, cookie := range redirectResp.Cookies() {
						cookies = append(cookies, map[string]interface{}{
							"name":     cookie.Name,
							"value":    cookie.Value,
							"domain":   cookie.Domain,
							"path":     cookie.Path,
							"expires":  cookie.Expires,
							"secure":   cookie.Secure,
							"httpOnly": cookie.HttpOnly,
						})
					}
					return cookies
				}()).
				Msg("Received redirect response")
			
			// Merge cookies from redirect response
			c.mergeCookies(redirectResp.Cookies())
		}
	}
	
	log.Debug().
		Int("cookies_stored", len(c.cookies)).
		Interface("stored_cookies", func() []map[string]interface{} {
			var cookies []map[string]interface{}
			for _, cookie := range c.cookies {
				cookies = append(cookies, map[string]interface{}{
					"name":     cookie.Name,
					"value":    cookie.Value,
					"domain":   cookie.Domain,
					"path":     cookie.Path,
					"expires":  cookie.Expires,
					"secure":   cookie.Secure,
					"httpOnly": cookie.HttpOnly,
				})
			}
			return cookies
		}()).
		Msg("Login successful, cookies stored")

	// Save cookies to file for future use
	if err := c.saveCookies(); err != nil {
		log.Debug().Err(err).Msg("Failed to save cookies to file")
		// Don't fail login if we can't save cookies
	}

	return nil
}

// FetchModemInfo fetches and parses modem information
func (c *Client) FetchModemInfo(ctx context.Context) (*ModemInfo, error) {
	// Load saved cookies first
	if err := c.loadCookies(); err != nil {
		log.Debug().Err(err).Msg("Failed to load cookies from file, will proceed without them")
	}
	
	// Try to fetch data first
	info, err := c.fetchModemInfoInternal(ctx)
	if err != nil {
		// Check if it's a forbidden error or logout detection (likely needs authentication)
		if strings.Contains(err.Error(), "403") || 
		   strings.Contains(err.Error(), "forbidden") ||
		   strings.Contains(err.Error(), "logged out") {
			// Try to login and retry
			if c.username != "" && c.password != "" {
				log.Debug().Msg("Authentication error detected, attempting login")
				if loginErr := c.Login(ctx); loginErr != nil {
					return nil, errors.Wrap(loginErr, "authentication failed")
				}
				// Retry after login
				log.Debug().Msg("Login successful, retrying data fetch")
				return c.fetchModemInfoInternal(ctx)
			}
			return nil, errors.New("access forbidden - authentication required")
		}
		return nil, err
	}
	return info, nil
}

// fetchModemInfoInternal performs the actual data fetching
func (c *Client) fetchModemInfoInternal(ctx context.Context) (*ModemInfo, error) {
	connectionStatusURL := c.baseURL + "/network_setup.jst"
	req, err := http.NewRequestWithContext(ctx, "GET", connectionStatusURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Add headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	
	// Set referer for connection status page
	req.Header.Set("Referer", c.baseURL+"/connection_status.jst")

	// Add cookies if we have them
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// Debug logging for connection status request
	log.Debug().
		Str("method", req.Method).
		Str("url", connectionStatusURL).
		Int("cookies_count", len(c.cookies)).
		Interface("cookies_sent", func() []map[string]interface{} {
			var cookies []map[string]interface{}
			for _, cookie := range c.cookies {
				cookies = append(cookies, map[string]interface{}{
					"name":     cookie.Name,
					"value":    cookie.Value,
					"domain":   cookie.Domain,
					"path":     cookie.Path,
					"expires":  cookie.Expires,
					"secure":   cookie.Secure,
					"httpOnly": cookie.HttpOnly,
				})
			}
			return cookies
		}()).
		Str("user_agent", req.Header.Get("User-Agent")).
		Str("referer", req.Header.Get("Referer")).
		Str("cookie_header", req.Header.Get("Cookie")).
		Msg("Sending connection status request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("Connection status request failed")
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	// Read response body for debugging and parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to read response body")
		return nil, errors.Wrap(err, "failed to read response body")
	}
	bodyString := string(bodyBytes)

	// Debug logging for connection status response
	log.Debug().
		Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Int("body_length", len(bodyString)).
		Str("content_type", resp.Header.Get("Content-Type")).
		Int("response_cookies_count", len(resp.Cookies())).
		Interface("response_cookies", func() []map[string]interface{} {
			var cookies []map[string]interface{}
			for _, cookie := range resp.Cookies() {
				cookies = append(cookies, map[string]interface{}{
					"name":     cookie.Name,
					"value":    cookie.Value,
					"domain":   cookie.Domain,
					"path":     cookie.Path,
					"expires":  cookie.Expires,
					"secure":   cookie.Secure,
					"httpOnly": cookie.HttpOnly,
				})
			}
			return cookies
		}()).
		Msg("Received connection status response")

	// Check for logout detection script in response body
	if strings.Contains(bodyString, `alertLoc("Please Login First!")`) || 
	   strings.Contains(bodyString, `location.href="home_loggedout.jst"`) {
		log.Debug().
			Msg("Detected logout script in response - authentication required")
		// Store response body to temporary file for debugging
		timestamp := time.Now().Format("20060102-150405")
		tmpFile := fmt.Sprintf("/tmp/fetch.%s.html", timestamp)
		if err := os.WriteFile(tmpFile, []byte(bodyString), 0644); err != nil {
			log.Debug().Err(err).Str("file", tmpFile).Msg("Failed to write response body to temp file")
		} else {
			log.Debug().Str("file", tmpFile).Msg("Saved response body to temp file")
		}
		return nil, fmt.Errorf("access forbidden (logged out) - authentication required")
	}

	if resp.StatusCode == http.StatusForbidden {
		log.Debug().
			Str("body", bodyString).
			Msg("Access forbidden - authentication may be required")
		return nil, fmt.Errorf("access forbidden (403) - authentication may be required")
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Debug().
			Int("status_code", resp.StatusCode).
			Str("body", bodyString).
			Msg("Unexpected status code received")
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	// Parse HTML from the body string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse HTML document")
		return nil, errors.Wrap(err, "failed to parse HTML")
	}

	log.Debug().Msg("Successfully parsed HTML document, extracting modem info")
	return c.parseModemInfo(doc)
}

// parseModemInfo parses the HTML document and extracts modem information
func (c *Client) parseModemInfo(doc *goquery.Document) (*ModemInfo, error) {
	info := &ModemInfo{
		LastUpdated: time.Now(),
	}

	log.Debug().Msg("Starting to parse modem information")

	// Parse cable modem information
	cableModem, err := c.parseCableModem(doc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse cable modem info")
		return nil, errors.Wrap(err, "failed to parse cable modem info")
	}
	info.CableModem = *cableModem
	log.Debug().
		Str("model", cableModem.Model).
		Str("vendor", cableModem.Vendor).
		Msg("Parsed cable modem info")

	// Parse downstream channels
	downstream, err := c.parseDownstreamChannels(doc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse downstream channels")
		return nil, errors.Wrap(err, "failed to parse downstream channels")
	}
	info.Downstream = downstream
	log.Debug().Int("downstream_channels", len(downstream)).Msg("Parsed downstream channels")

	// Parse upstream channels
	upstream, err := c.parseUpstreamChannels(doc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse upstream channels")
		return nil, errors.Wrap(err, "failed to parse upstream channels")
	}
	info.Upstream = upstream
	log.Debug().Int("upstream_channels", len(upstream)).Msg("Parsed upstream channels")

	// Parse error codewords
	errorCodewords, err := c.parseErrorCodewords(doc)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to parse error codewords")
		return nil, errors.Wrap(err, "failed to parse error codewords")
	}
	info.ErrorCodewords = errorCodewords
	log.Debug().Int("error_channels", len(errorCodewords)).Msg("Parsed error codewords")

	log.Debug().
		Int("total_downstream", len(info.Downstream)).
		Int("total_upstream", len(info.Upstream)).
		Int("total_error_channels", len(info.ErrorCodewords)).
		Msg("Successfully parsed all modem information")

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