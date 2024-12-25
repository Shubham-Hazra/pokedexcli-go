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
	cacheInterval = 15 * time.Second
)

type LocationAreaConfig struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreaResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(args []string) error
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

type cli struct {
	cache  *pokecache.Cache
	LocationAreaConfig *LocationAreaConfig
}

func (c *cli) commandExit(args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func (c *cli) commandHelp(args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, cmd := range getCommands(c) {
		fmt.Printf("%v: %v\n", cmd.name, cmd.description)
	}
	return nil
}

func (c *cli) commandMap(args []string) error {
	if c.LocationAreaConfig.Next == "" {
		fmt.Println("No more pages to load.")
		return nil
	}

	err := fetchData(c.cache, c.LocationAreaConfig.Next, c.LocationAreaConfig)
	if err != nil {
		return err
	}

	for _, loc := range c.LocationAreaConfig.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func (c *cli) commandMapb(args []string) error {
	if c.LocationAreaConfig.Previous == "" {
		fmt.Println("No previous page available.")
		return nil
	}

	err := fetchData(c.cache, c.LocationAreaConfig.Previous, c.LocationAreaConfig)
	if err != nil {
		return err
	}

	for _, loc := range c.LocationAreaConfig.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func (c *cli) commandExplore(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a location area name")
	}

	locationArea := args[0]
	url := fmt.Sprintf("%s/location-area/%s", baseAPIURL, locationArea)
	var response LocationAreaResponse
	err := fetchData(c.cache, url, &response)
	if err != nil {
		return fmt.Errorf("error exploring location area: %w", err)
	}

	fmt.Printf("Exploring %s...\n", locationArea)
	fmt.Println("Found Pokemon:")

	seenPokemon := make(map[string]bool)

	for _, encounter := range response.PokemonEncounters {
		if !seenPokemon[encounter.Pokemon.Name] {
			fmt.Printf(" - %s\n", encounter.Pokemon.Name)
			seenPokemon[encounter.Pokemon.Name] = true
		}
	}

	return nil
}

func getCommands(c *cli) map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    c.commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    c.commandExit,
		},
		"map": {
			name:        "map",
			description: "Show the next page of locations",
			callback:    c.commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Show the previous page of locations",
			callback:    c.commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location area",
			callback:    c.commandExplore,
		},
	}
}

func main() {
	cache := pokecache.NewCache(cacheInterval)
	LocationAreaConfig := LocationAreaConfig{
		Next:     baseAPIURL + "/location-area/?offset=20&limit=20",
		Previous: "",
	}

	c := &cli{
		cache:  &cache,
		LocationAreaConfig: &LocationAreaConfig,
	}

	commands := getCommands(c)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")

		if !scanner.Scan() {
			break
		}
		userInput := scanner.Text()
		inputParts := utils.CleanInput(userInput)

		if len(inputParts) < 1 {
			fmt.Println("Please enter a command.")
			continue
		}

		cmdName := inputParts[0]
		args := inputParts[1:]

		if command, ok := commands[cmdName]; ok {
			if err := command.callback(args); err != nil {
				fmt.Printf("Encountered an error: %v\n", err)
			}
		} else {
			fmt.Printf("Unsupported command: %v\n", cmdName)
		}
	}
}
