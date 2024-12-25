package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Shubham-Hazra/pokedexcli/pokecache"
	"github.com/Shubham-Hazra/pokedexcli/utils"
)

const (
	baseAPIURL    = "https://pokeapi.co/api/v2"
	cacheInterval = 1 * time.Minute
)

type Config struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func fetchData(cache *pokecache.Cache, url string, target interface{}) error {
	if cachedData, found := cache.Get(url); found {
		log.Println("Using cached data for:", url)
		return json.Unmarshal(cachedData, target)
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return fmt.Errorf("failed to fetch data, status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	cache.Add(url, body)
	return json.Unmarshal(body, target)
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(commands *map[string]cliCommand) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, cmd := range *commands {
		fmt.Printf("%v: %v\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap(cache *pokecache.Cache, config *Config) error {
	if config.Next == "" {
		fmt.Println("No more pages to load.")
		return nil
	}

	err := fetchData(cache, config.Next, config)
	if err != nil {
		return err
	}

	for _, loc := range config.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func commandMapb(cache *pokecache.Cache, config *Config) error {
	if config.Previous == "" {
		fmt.Println("No previous page available.")
		return nil
	}

	err := fetchData(cache, config.Previous, config)
	if err != nil {
		return err
	}

	for _, loc := range config.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func main() {
	cache := pokecache.NewCache(cacheInterval)
	location_area_config := Config{
		Next:     baseAPIURL + "/location-area/?offset=20&limit=20",
		Previous: "",
	}

	var commands map[string]cliCommand
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    func () error {return commandHelp(&commands)},
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Show the next page of locations",
			callback:    func () error {return commandMap(&cache, &location_area_config)},
		},
		"mapb": {
			name:        "mapb",
			description: "Show the previous page of locations",
			callback:    func () error {return commandMapb(&cache, &location_area_config)},
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()
		cleanedInput := utils.CleanInput(userInput)

		if len(cleanedInput) < 1 {
			fmt.Println("Please enter a command.")
			continue
		}

		cmd := cleanedInput[0]
		if command, ok := commands[cmd]; ok {
			if err := command.callback(); err != nil {
				fmt.Printf("Encountered an error: %v\n", err)
			}
		} else {
			fmt.Printf("Unsupported command: %v\n", cmd)
		}
	}
}