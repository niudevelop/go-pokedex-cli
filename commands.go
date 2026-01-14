package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

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
	callback    func(config *Config, args ...string) error
}

var Pokedex = make(map[string]Pokemon)

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
		"explore": {
			name:        "explore",
			description: "Explore Location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect Pokemon from Pokedex",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show all catched Pokemons",
			callback:    commandPokedex,
		},
	}

}

func commandPokedex(config *Config, args ...string) error {
	fmt.Println("Your Pokedex:")
	for name, _ := range Pokedex {
		fmt.Printf("- %s\n", name)
	}
	return nil
}

func commandExit(config *Config, args ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandInspect(config *Config, args ...string) error {
	if len(args) != 1 || args[0] == "" {
		return fmt.Errorf("Give a pokemon name to inspect\n")
	}
	pokemonName := args[0]
	pokemon, ok := Pokedex[pokemonName]
	if !ok {
		return fmt.Errorf("Pokemon '%s' not found in Pokedex", pokemonName)
	}
	fmt.Printf("\nName: %s", pokemon.Name)
	fmt.Printf("\nHeight: %d", pokemon.Height)
	fmt.Printf("\nWeight: %d", pokemon.Weight)
	fmt.Printf("\nStats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("\n\t-%s: %v", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Printf("\nTypes:")
	for _, types := range pokemon.Types {
		fmt.Printf("\n\t- %s", types.Type.Name)
	}
	fmt.Println()
	return nil
}
func commandCatch(config *Config, args ...string) error {
	var pokemon Pokemon
	if len(args) != 1 || args[0] == "" {
		return fmt.Errorf("Give a pokemon name to catch\n")
	}
	pokemonName := args[0]
	fullURL := "https://pokeapi.co/api/v2/pokemon/" + pokemonName

	if entry, ok := config.PokeCache.Get(fullURL); ok {
		err := json.Unmarshal(entry, &pokemon)
		if err != nil {
			return fmt.Errorf("Error Unmarshal: %s", err)
		}
	} else {
		resp, err := http.Get(fullURL)
		if err != nil {
			return fmt.Errorf("network error: %v", err)
		}
		if resp.StatusCode > 299 {
			return fmt.Errorf("Pokemon '%v' not found", pokemonName)
		}
		defer resp.Body.Close()

		decorder := json.NewDecoder(resp.Body)

		if err := decorder.Decode(&pokemon); err != nil {
			return fmt.Errorf("Decoding Error: %v", err)
		}
		jsonData, err := json.Marshal(pokemon)
		if err != nil {
			return fmt.Errorf("Error Marshal %s", err)
		}
		config.PokeCache.Add(fullURL, jsonData)

	}
	fmt.Printf("\nThrowing a Pokeball at %s...\n", pokemonName)
	catched := tryCatch(pokemon.BaseExperience)
	if !catched {
		fmt.Printf("%s escaped!\n", pokemonName)
		return nil
	}
	fmt.Printf("%s was caught!\n", pokemonName)
	Pokedex[pokemonName] = pokemon

	return nil
}

func tryCatch(baseExp int) bool {
	var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	chance := 1.0 / (1.0 + float64(baseExp)/50.0)
	if chance < 0.05 {
		chance = 0.05
	}
	return rng.Float64() < chance
}

func commandExplore(config *Config, args ...string) error {
	var locationInfo LocationInfo
	if len(args) != 1 || args[0] == "" {
		return fmt.Errorf("Give a location name to explore area\n")
	}
	locationName := args[0]
	fullURL := "https://pokeapi.co/api/v2/location-area/" + locationName
	if entry, ok := config.PokeCache.Get(fullURL); ok {
		err := json.Unmarshal(entry, &locationInfo)
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

		if err := decorder.Decode(&locationInfo); err != nil {
			return fmt.Errorf("Decoding Error: %v", err)
		}
		jsonData, err := json.Marshal(locationInfo)
		if err != nil {
			return fmt.Errorf("Error Marshal %s", err)
		}
		config.PokeCache.Add(fullURL, jsonData)
	}

	for _, pokemon := range locationInfo.PokemonEncounters {
		fmt.Println()
		fmt.Printf("%s", pokemon.Pokemon.Name)
	}
	fmt.Println()
	return nil
}

func commandMap(config *Config, args ...string) error {
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

func commandMapB(config *Config, args ...string) error {
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

func commandHelp(config *Config, args ...string) error {
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

type LocationInfo struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Cries          struct {
		Latest string `json:"latest"`
		Legacy string `json:"legacy"`
	} `json:"cries"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height    int `json:"height"`
	HeldItems []struct {
		Item struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"item"`
		VersionDetails []struct {
			Rarity  int `json:"rarity"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"held_items"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			Order        any `json:"order"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	PastAbilities []struct {
		Abilities []struct {
			Ability  any  `json:"ability"`
			IsHidden bool `json:"is_hidden"`
			Slot     int  `json:"slot"`
		} `json:"abilities"`
		Generation struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"generation"`
	} `json:"past_abilities"`
	PastTypes []any `json:"past_types"`
	Species   struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       any    `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  any    `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      any    `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale any    `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string `json:"front_default"`
				FrontFemale  any    `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string `json:"back_default"`
				BackFemale       any    `json:"back_female"`
				BackShiny        string `json:"back_shiny"`
				BackShinyFemale  any    `json:"back_shiny_female"`
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationIx struct {
				ScarletViolet struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"scarlet-violet"`
			} `json:"generation-ix"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string `json:"back_default"`
						BackFemale       any    `json:"back_female"`
						BackShiny        string `json:"back_shiny"`
						BackShinyFemale  any    `json:"back_shiny_female"`
						FrontDefault     string `json:"front_default"`
						FrontFemale      any    `json:"front_female"`
						FrontShiny       string `json:"front_shiny"`
						FrontShinyFemale any    `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				BrilliantDiamondShiningPearl struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"brilliant-diamond-shining-pearl"`
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}
