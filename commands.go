package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/niudevelop/pokedexcli/internal/pokecache"
)

type Config struct {
	Next      string
	Previous  string
	PokeCache *pokecache.Cache
}
type LocationArea struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []Location `json:"results"`
}

type Location struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(config *Config) error
}

var Commands map[string]cliCommand

func init() {
	Commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Display Pokemon Map",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display Previous Pokemon Map",
			callback:    commandMapB,
		},
	}

}

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(config *Config) error {
	var locationArea LocationArea
	limit := "20"
	fullURL := "https://pokeapi.co/api/v2/location-area/?limit=" + limit
	if config.Next != "" {
		fullURL = config.Next
	}
	if entry, ok := config.PokeCache.Get(fullURL); ok {
		err := json.Unmarshal(entry, &locationArea)
		if err != nil {
			return fmt.Errorf("Error Unmarshal: %s", err)
		}
	} else {
		resp, err := http.Get(fullURL)
		if err != nil {
			return fmt.Errorf("network request error: %v", err)
		}
		defer resp.Body.Close()

		decorder := json.NewDecoder(resp.Body)

		if err := decorder.Decode(&locationArea); err != nil {
			return fmt.Errorf("Decoding Error: %v", err)
		}
		jsonData, err := json.Marshal(locationArea)
		if err != nil {
			return fmt.Errorf("Error Marshal %s", err)
		}
		config.PokeCache.Add(fullURL, jsonData)
	}

	for _, location := range locationArea.Results {
		fmt.Println()
		fmt.Printf("%s", location.Name)
	}
	fmt.Println()
	config.Next = locationArea.Next
	config.Previous = locationArea.Previous

	return nil

}

func commandMapB(config *Config) error {
	var locationArea LocationArea
	limit := "20"
	fullURL := "https://pokeapi.co/api/v2/location-area/?limit=" + limit
	if config.Previous != "" {
		fullURL = config.Previous
	}
	if entry, ok := config.PokeCache.Get(fullURL); ok {
		err := json.Unmarshal(entry, &locationArea)
		if err != nil {
			return fmt.Errorf("Error Unmarshal: %s", err)
		}
	} else {
		resp, err := http.Get(fullURL)
		if err != nil {
			return fmt.Errorf("network request error: %v", err)
		}
		defer resp.Body.Close()

		decorder := json.NewDecoder(resp.Body)
		if err := decorder.Decode(&locationArea); err != nil {
			return fmt.Errorf("Decoding Error: %v", err)
		}
		jsonData, err := json.Marshal(locationArea)
		if err != nil {
			return fmt.Errorf("Error Marshal %s", err)
		}
		config.PokeCache.Add(fullURL, jsonData)
	}
	for _, location := range locationArea.Results {
		fmt.Println()
		fmt.Printf("%s", location.Name)
	}
	fmt.Println()
	config.Next = locationArea.Next
	config.Previous = locationArea.Previous

	return nil
}

func commandHelp(config *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for key, command := range Commands {
		fmt.Printf("%s: %s\n", key, command.description)
	}
	return nil
}

func getCommands() map[string]cliCommand {
	return Commands
}
