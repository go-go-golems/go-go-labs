package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net/url"
)

func (ei *EmrichenInterpreter) handleURLEncode(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind == yaml.ScalarNode {
		// Simple string encoding
		return makeString(url.QueryEscape(node.Value)), nil
	} else if node.Kind == yaml.MappingNode {
		urlStr, queryParams, err := ei.parseURLEncodeArgs(node)
		if err != nil {
			return nil, err
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL in !URLEncode: %v", err)
		}

		query := parsedURL.Query()
		for k, v := range queryParams {
			query.Set(k, fmt.Sprintf("%s", v))
		}
		parsedURL.RawQuery = query.Encode()

		return makeString(parsedURL.String()), nil
	}

	return nil, errors.New("!URLEncode requires a scalar or mapping node")

}
