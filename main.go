package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		userInput := scanner.Text()
		cleandInput := cleanInput(userInput)

		fmt.Printf("Your command was: %s\n", cleandInput[0])
	}

}
