package main

import (
	"github.com/rs/zerolog"
	"os"
)

type Address struct {
	Street string
	City   string
}

type Company struct {
	Name    string
	Address Address
}

type Person struct {
	Name    string
	Age     int
	Company Company
}

func (a Address) MarshalZerologObject(e *zerolog.Event) {
	e.Str("street", a.Street).Str("city", a.City)
}

func (c Company) MarshalZerologObject(e *zerolog.Event) {
	e.Str("name", c.Name).Object("address", c.Address)
}

func (p Person) MarshalZerologObject(e *zerolog.Event) {
	e.Str("name", p.Name).Int("age", p.Age).Object("company", p.Company)
}

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	person := Person{
		Name: "John Doe",
		Age:  30,
		Company: Company{
			Name: "ACME Corp",
			Address: Address{
				Street: "123 Main St",
				City:   "Anytown",
			},
		},
	}

	logger.Info().
		Str("event", "user_login").
		Int("user_id", 12345).
		Bool("success", true).
		Float32("response_time", 0.23).
		Object("person", person).
		Msg("User logged in")
}
