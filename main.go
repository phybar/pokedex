package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"encoding/json"
	"net/http"
	"io"
)
// variable for the input command used after it has been cleaned
var command string

// Struct for the cliCommands
type cliCommand struct {
	name		string
	description	string
	callback	func(cfg *Config) error
}

// Struct for the location data from API call
type Location struct {
    Next     *string `json:"next"`
    Previous *string `json:"previous"`
    Results  []struct {
        Name string `json:"name"`
    } `json:"results"`
}

type Config struct {
	next 	*string // URL for the next page in the list
	previous *string // URL for the previous page in the list
}


var commands = map[string]cliCommand{
	"exit": {
		name: 		"exit",
		description:	"Exit the Pokedex",
		callback:	commandExit,
	},
	"help": {
		name:	"help",
		description: "Displays a help message",
		callback: commandHelp,
	},
	"map": {
		name:	"map",
		description: "Displays the next 20 locations from the Pokemon world",
		callback: commandMap,
	},
	"mapb": {
		name:	"mapb",
		description: "Displays the next 20 locations from the Pokemon world",
		callback: commandMapb,
	},
}

func main(){
	scanner := bufio.NewScanner(os.Stdin)

	cfg := &Config{
        next: nil,
        previous: nil,
    }


	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			input := scanner.Text()
			cleanedInput := cleanInput(input)

			// cleanInput returns a slice of strings
			if len(cleanedInput) > 0 {
				command = cleanedInput[0] // Capture first word
				fmt.Printf("Your command was: %s\n", command)

			}

			if cmd, exists := commands[command]; exists {
				err := cmd.callback(cfg)
				if err != nil {
					fmt.Println("Error executing command:", err)
				}
			}
		} else {
			fmt.Println("No input detected")
			break
		}
		}
	}


func commandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex")
	return nil
}
// This function will send a Get request to the pokeApi for locations
// No query is sent, so it should return 20 results
func commandMap(cfg *Config) error {
	baseURL := "https://pokeapi.co/api/v2/location-area/"
    url := baseURL


	if cfg.next != nil {
		url = *cfg.next
	}
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	defer res.Body.Close()

// Takes the response and checks the status code to make sure its complete
	body, err := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Println("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	// Creates new location struct
	var location Location
	// Uses the helper unmarshal function
	err = unmarshalJSON(body, &location)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	cfg.next = location.Next
	cfg.previous = location.Previous 

	// Prints each area/result
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
	return nil

}

func commandMapb(cfg *Config) error {
	if cfg.previous == nil {
		fmt.Println("No previous command")
		return nil
	} 
	
	url := *cfg.previous
	
	
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	defer res.Body.Close()

// Takes the response and checks the status code to make sure its complete
	body, err := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Println("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	// Creates new location struct
	var location Location
	// Uses the helper unmarshal function
	err = unmarshalJSON(body, &location)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	cfg.next = location.Next
	cfg.previous = location.Previous 
	// Prints each result
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
	return nil
}

// Unmarshalling of JSON data recieved - this will take a slice of bytes, and pass it 
// into a struct for the correct data type, returning the struct
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func cleanInput(input string) []string {
	// logic for cleaning the input goes here
	var cleaned []string
	newList := (strings.Split(input, " "))

	// Prcocess each word
	for _, word := range newList {
		lower := strings.ToLower(word)
		trimmed := strings.TrimSpace(lower)
		if trimmed != "" { 
		cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}
