package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/niudevelop/pokedexcli/internal/pokecache"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := Config{}
	config.PokeCache = pokecache.NewCache(5 * time.Minute)

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		userInput := scanner.Text()
		cleandInput := cleanInput(userInput)

		command, ok := getCommands()[cleandInput[0]]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		var argument string
		if len(cleandInput) == 2 {
			argument = cleandInput[1]
		}
		err := command.callback(&config, argument)
		if err != nil {
			fmt.Printf("Error: %v", err)
			continue
		}

		// fmt.Printf("Your command was: %s\n", cleandInput[0])
	}

}
