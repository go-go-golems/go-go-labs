package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/cms/pkg/sqlite"
	"log"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cli",
		Short: "CLI utility for managing a SQLite database",
	}

	var database, schemaFile string

	rootCmd.PersistentFlags().StringVar(&database, "database", "", "SQLite database file")
	rootCmd.PersistentFlags().StringVar(&schemaFile, "schema", "", "YAML DSL schema file")

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Initialize the DB",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := sqlx.Open("sqlite3", database)
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			defer func(db *sqlx.DB) {
				_ = db.Close()
			}(db)

			schema, err := loadSchema(schemaFile)
			if err != nil {
				log.Fatalf("Failed to load schema: %v", err)
			}

			if err := sqlite.CreateSchema(context.Background(), schema, db); err != nil {
				log.Fatalf("Failed to create schema: %v", err)
			}

			fmt.Println("DB created")
		},
	}

	var insertCmd = &cobra.Command{
		Use:   "insert [json/yaml file]",
		Short: "Insert data into the DB",
		Run: func(cmd *cobra.Command, args []string) {
			db, schema, err := openDBAndLoadSchema(database, schemaFile)
			if err != nil {
				log.Fatalf("Failed to open database and load schema: %v", err)
			}
			defer func(db *sqlx.DB) {
				err := db.Close()
				if err != nil {
					log.Fatalf("Failed to close database: %v", err)
				}
			}(db)

			data, err := loadJSON(args[0])
			if err != nil {
				log.Fatalf("Failed to load JSON data: %v", err)
			}

			if _, err := sqlite.InsertData(context.Background(), db, schema, data); err != nil {
				log.Fatalf("Failed to insert data: %v", err)
			}

			fmt.Println("Data inserted")
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all objects in the DB",
		Run: func(cmd *cobra.Command, args []string) {
			db, schema, err := openDBAndLoadSchema(database, schemaFile)
			if err != nil {
				log.Fatalf("Failed to open database and load schema: %v", err)
			}
			defer func(db *sqlx.DB) {
				_ = db.Close()
			}(db)

			offset, _ := cmd.Flags().GetInt("offset")
			limit, _ := cmd.Flags().GetInt("limit")

			paginationRequest := &sqlite.PaginationRequest{
				Offset: offset,
				Limit:  limit,
			}

			response, err := sqlite.ListObjects(context.Background(), db, schema, paginationRequest)
			if err != nil {
				log.Fatalf("Failed to list objects: %v", err)
			}

			fmt.Println("Listing all objects")
			fmt.Println(response)
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [id]",
		Short: "Get an object by ID",
		Run: func(cmd *cobra.Command, args []string) {
			db, schema, err := openDBAndLoadSchema(database, schemaFile)
			if err != nil {
				log.Fatalf("Failed to open database and load schema: %v", err)
			}
			defer func(db *sqlx.DB) {
				_ = db.Close()
			}(db)

			// parse id as int64
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Fatalf("Failed to parse ID: %v", err)
			}

			object, err := sqlite.GetObject(context.Background(), db, schema, id)
			if err != nil {
				log.Fatalf("Failed to get object: %v", err)
			}

			fmt.Println("Getting object by ID")
			fmt.Println(object)
		},
	}

	listCmd.Flags().Int("offset", 0, "Offset for pagination")
	listCmd.Flags().Int("limit", 0, "Limit for pagination")

	rootCmd.AddCommand(createCmd, insertCmd, listCmd, getCmd)

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func loadSchema(file string) (*pkg.Schema, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema pkg.Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

func openDBAndLoadSchema(database, schemaFile string) (*sqlx.DB, *pkg.Schema, error) {
	db, err := sqlx.Open("sqlite3", database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	schema, err := loadSchema(schemaFile)
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return db, schema, nil
}

func loadJSON(filename string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}
