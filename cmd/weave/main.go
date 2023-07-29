package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"io"
	"net/http"
	"strings"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func main() {
	cfg := weaviate.Config{
		Host:   "localhost:8080/", // Replace with your endpoint
		Scheme: "http",
		//AuthConfig: auth.ApiKey{Value: "YOUR-WEAVIATE-API-KEY"}, // Replace w/ your Weaviate instance API key
		Headers: map[string]string{
			//"X-HuggingFace-Api-Key": "YOUR-HUGGINGFACE-API-KEY", // Replace with your inference API key
		},
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	if false {
		err = deleteClass(client, "Question")
		if err != nil {
			panic(err)
		}
		createClass(client)
		batchImport(client)
	} else {
		query(client)
	}
}

func deleteClass(client *weaviate.Client, className string) error {
	// delete the class
	if err := client.Schema().ClassDeleter().WithClassName(className).Do(context.Background()); err != nil {
		// Weaviate will return a 400 if the class does not exist, so this is allowed, only return an error if it's not a 400
		if status, ok := err.(*fault.WeaviateClientError); ok && status.StatusCode != http.StatusBadRequest {
			return err
		}
	}

	return nil
}

func query(client *weaviate.Client) {
	fields := []graphql.Field{
		{Name: "question"},
		{Name: "answer"},
		{Name: "category"},
	}

	nearText := client.GraphQL().
		NearTextArgBuilder().
		WithConcepts([]string{"biology"})

	where := filters.Where().
		WithPath([]string{"category"}).
		WithOperator(filters.Equal).
		WithValueText("ANIMALS")

	result, err := client.GraphQL().Get().
		WithClassName("Question").
		WithFields(fields...).
		WithNearText(nearText).
		WithWhere(where).
		WithLimit(2).
		Do(context.Background())
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", buf.String())
}

func createClass(client *weaviate.Client) {
	classObj := &models.Class{
		Class:      "Question",
		Vectorizer: "text2vec-openai",
		ModuleConfig: map[string]interface{}{
			"text2vec-openai": map[string]interface{}{
				"model":        "ada",
				"modelVersion": "002",
				"type":         "text",
			},
		},
	}

	// add the schema
	err := client.Schema().ClassCreator().WithClass(classObj).Do(context.Background())
	if err != nil {
		//panic(err)
		log.Error().Err(err).Msg("error creating class")
	}
}

func batchImport(client *weaviate.Client) {
	// Retrieve the data
	data, err := http.DefaultClient.Get("https://raw.githubusercontent.com/weaviate-tutorials/quickstart/main/data/jeopardy_tiny.json")
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(data.Body)

	// Decode the data
	var items []map[string]string
	if err := json.NewDecoder(data.Body).Decode(&items); err != nil {
		panic(err)
	}

	// convert items into a slice of models.Object
	objects := make([]*models.Object, len(items))
	for i := range items {
		objects[i] = &models.Object{
			Class: "Question",
			Properties: map[string]any{
				"category": items[i]["Category"],
				"question": items[i]["Question"],
				"answer":   items[i]["Answer"],
			},
		}
	}

	// batch write items
	batchRes, err := client.Batch().ObjectsBatcher().WithObjects(objects...).Do(context.Background())
	if err != nil {
		panic(err)
	}
	for _, res := range batchRes {
		if res.Result.Errors != nil {
			panic(res.Result.Errors.Error)
		}
	}
}
