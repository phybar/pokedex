package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
	"encoding/json"
	"net/http"
	"io"
	"time"
	"github.com/phybar/pokedexcli/pokecache"
	"reflect"
)
// variable for the input command used after it has been cleaned
var command string

// Struct for the cliCommands
type cliCommand struct {
	name		string
	description	string
	callback    interface{} // 
}

type cliCommandFunc func(cfg *Config) error
type cliCommandWithCacheFunc func(cfg *Config, cache *pokecache.Cache) error
// type cliCommandExplore func(cfg *Config, input string) error

// Struct for the location data from API call
type Location struct {
    Next     *string `json:"next"`
    Previous *string `json:"previous"`
    Results  []struct {
        Name string `json:"name"`
    } `json:"results"`
}

type LocationArea struct {
    PokemonEncounters []struct {
        Pokemon struct {
            Name string `json:"name"`
        } `json:"pokemon"`
    } `json:"pokemon_encounters"`
}

type Config struct {
	next 	*string // URL for the next page in the list
	previous *string // URL for the previous page in the list
	params []string // Store command parameters
}


var commands = map[string]cliCommand{
	"exit": {
		name: 		"exit",
		description:	"Exit the Pokedex",
		callback:	cliCommandFunc(commandExit),
	},
	"help": {
		name:	"help",
		description: "Displays a help message",
		callback: cliCommandFunc(commandHelp),
	},
	"map": {
		name:	"map",
		description: "Displays the next 20 locations from the Pokemon world",
		callback: cliCommandWithCacheFunc(commandMap),
	},
	"mapb": {
		name:	"mapb",
		description: "Displays the next 20 locations from the Pokemon world",
		callback: cliCommandWithCacheFunc(commandMapb),
	},
	"explore": {
		name: "explore",
		description: "Lists all Pokemon withing an area",
		callback: cliCommandWithCacheFunc(commandExplore),
	},
}


func main(){
	scanner := bufio.NewScanner(os.Stdin)

	// Initialise the cache
	var cache = pokecache.NewCache(5 * time.Minute)


	cfg := &Config{
        next: nil,
        previous: nil,
		params: []string{},
    }


	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			input := scanner.Text()
			cleanedInput := cleanInput(input)
			

			// cleanInput returns a slice of strings
			if len(cleanedInput) > 0 {
				command = cleanedInput[0] // Capture first word
				if len(cleanedInput) > 1 {
					cfg.params = cleanedInput[1:]
				} else {
					cfg.params = []string{}
				}
				fmt.Printf("Your command was: %s\n", command)

			

			cmd, exists := commands[command]
			if !exists {
				fmt.Println("Command not found")
				return
			}
			
			var err error

			switch fn := cmd.callback.(type) {
			case cliCommandWithCacheFunc:
				err = fn(cfg, cache)
			case cliCommandFunc:
				err = fn(cfg)
			// case cliCommandExplore:
			// 	err = fn(cfg, input)
			default:
				fmt.Printf("Unrecognized command type: %s\n", reflect.TypeOf(cmd.callback))
			}

			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		
		
				}
			}
		}	
}


func commandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nmap: Displays the next 20 map locations\nmapb: Displays the previous 20 map locations\nexit: Exit the Pokedex")
	return nil
}
// This function will send a Get request to the pokeApi for locations
// No query is sent, so it should return 20 results
func commandMap(cfg *Config, cache *pokecache.Cache) error {
	baseURL := "https://pokeapi.co/api/v2/location-area/"
    url := baseURL
	// Creates new location struct
	var location Location
	var body []byte
	

	if cfg.next != nil {
		url = *cfg.next
	}
	
	entry, exists := cache.Get(url)
	if exists {
		body = entry
	} else {

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s", err)
		fmt.Println("Reached error block")
		fmt.Println("Using body:", string(body))
		return err
	}
	defer res.Body.Close()

// Takes the response and checks the status code to make sure its complete
	body, err := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	
	// Uses the helper unmarshal function
	err = unmarshalJSON(body, &location)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	cache.Add(url, body)

}

	cfg.next = location.Next
	cfg.previous = location.Previous 

	// Prints each area/result
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
	return nil

}

func commandMapb(cfg *Config, cache *pokecache.Cache) error {
	if cfg.previous == nil {
		fmt.Println("No previous command")
		return nil
	} 
	var location Location
	var body []byte
	
	url := *cfg.previous
	
	entry, exists := cache.Get(url)
	if exists {
		body = entry
	} else {
	
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s", err)
		fmt.Println("Reached error block")
		fmt.Println("Using body:", string(body))
		return err
	}
	defer res.Body.Close()

// Takes the response and checks the status code to make sure its complete
	body, err := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	// Creates new location struct
	
	// Uses the helper unmarshal function
	err = unmarshalJSON(body, &location)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	cache.Add(url, body)

}

	cfg.next = location.Next
	cfg.previous = location.Previous 
	// Prints each result
	for _, result := range location.Results {
		fmt.Println(result.Name)
	}
	return nil
}

func commandExplore(cfg *Config, cache *pokecache.Cache) error {
	// This command shows a list of all pokemon within the area of the last map command
	// It uses a cache
	// It will have to take a name as an input, the use that in the API call...
	if len(cfg.params) < 1 {
		fmt.Println("Please provide a location name")
		return nil
	}

	locationExplore := cfg.params[0]
	
	baseURL := "https://pokeapi.co/api/v2/location-area/"
    url := baseURL + locationExplore
	// Creates new location struct
	var locationArea LocationArea
	var body []byte
	


	
	entry, exists := cache.Get(url)
	if exists {
		body = entry
	} else {

	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %s", err)
		fmt.Println("Reached error block")
		fmt.Println("Using body:", string(body))
		return err
	}
	defer res.Body.Close()

// Takes the response and checks the status code to make sure its complete
	body, err := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	
	// Uses the helper unmarshal function
	err = unmarshalJSON(body, &locationArea)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	cache.Add(url, body)

}


	fmt.Println("Found Pokemon")
	// Prints each Pokemon in the area
	for _, encounter := range locationArea.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
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
