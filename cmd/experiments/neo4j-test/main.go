package main

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	ctx := context.Background()
	// Create a driver instance
	driver, err := neo4j.NewDriverWithContext(
		"neo4j://localhost:7687",
		neo4j.BasicAuth("neo4j", "testtest", ""),
	)
	if err != nil {
		fmt.Println("Error creating driver:", err)
		return
	}
	defer driver.Close(ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		fmt.Println("Error verifying connectivity:", err)
		return
	}

	// Create a session
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	createPerson(session, ctx, "John Doe")
	readPerson(session, ctx, "John Doe")
	updatePersonName(session, ctx, "John Doe", "Jane Doe")
	deletePerson(session, ctx, "Jane Doe")

	fmt.Println("Successfully connected to Neo4j!")
}

func createPerson(session neo4j.SessionWithContext, ctx context.Context, name string) {
	_, err := session.Run(ctx, "CREATE (p:Person {name: $name})", map[string]interface{}{"name": name})
	if err != nil {
		fmt.Println("Error creating person:", err)
	} else {
		fmt.Println("Person created:", name)
	}
}

// Read a person by name
func readPerson(session neo4j.SessionWithContext, ctx context.Context, name string) {
	result, err := session.Run(ctx, "MATCH (p:Person {name: $name}) RETURN p", map[string]interface{}{"name": name})
	if err != nil {
		fmt.Println("Error reading person:", err)
		return
	}

	for result.Next(ctx) {
		record := result.Record()
		personMap := record.AsMap()
		fmt.Printf("Found person: %s\n", personMap["name"])
	}
}

// Update a person's name
func updatePersonName(session neo4j.SessionWithContext, ctx context.Context, oldName string, newName string) {
	_, err := session.Run(ctx, "MATCH (p:Person {name: $oldName}) SET p.name = $newName", map[string]interface{}{"oldName": oldName, "newName": newName})
	if err != nil {
		fmt.Println("Error updating person's name:", err)
	} else {
		fmt.Printf("Updated person's name from %s to %s\n", oldName, newName)
	}
}

// Delete a person by name
func deletePerson(session neo4j.SessionWithContext, ctx context.Context, name string) {
	_, err := session.Run(ctx, "MATCH (p:Person {name: $name}) DELETE p", map[string]interface{}{"name": name})
	if err != nil {
		fmt.Println("Error deleting person:", err)
	} else {
		fmt.Printf("Deleted person: %s\n", name)
	}
}
