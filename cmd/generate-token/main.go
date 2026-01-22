package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"golang.org/x/term"
)

type Config struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

func main() {
	// Ensure terminal is in normal mode (not raw mode)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.GetState(int(os.Stdin.Fd()))
		if err == nil {
			defer term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}

	fmt.Println("=== Zerodha Kite Connect - Access Token Generator ===\n")

	// Read config
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	kc := kiteconnect.New(config.APIKey)

	// Step 1: Display login URL
	fmt.Println("STEP 1: Login to Kite Connect")
	fmt.Println("Copy and paste this URL in your browser:\n")
	fmt.Println(kc.GetLoginURL())
	fmt.Println()

	// Step 2: Get request token
	fmt.Println("STEP 2: After logging in, you'll be redirected to your redirect URL")
	fmt.Println("Copy the 'request_token' parameter from the URL")
	fmt.Print("\nEnter request_token: ")

	// Use bufio for reliable input reading
	reader := bufio.NewReader(os.Stdin)
	requestToken, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}

	// Clean the input - remove all whitespace and control characters
	requestToken = strings.TrimSpace(requestToken)
	requestToken = strings.ReplaceAll(requestToken, "\r", "")
	requestToken = strings.ReplaceAll(requestToken, "\n", "")

	if requestToken == "" {
		log.Fatal("Request token cannot be empty")
	}

	// Step 3: Generate session
	fmt.Println("\nGenerating access token...")
	data, err := kc.GenerateSession(requestToken, config.APISecret)
	if err != nil {
		log.Fatalf("Error generating session: %v\n\nPlease ensure:\n1. Request token is copied correctly\n2. Token hasn't expired (valid for ~2 minutes)\n3. You haven't used this token before", err)
	}

	// Step 4: Display results
	fmt.Println("\n=== ✅ SUCCESS! ===")
	fmt.Printf("\nYour Access Token:\n%s\n", data.AccessToken)
	fmt.Printf("\nUser ID: %s\n", data.UserID)
	fmt.Printf("User Name: %s\n", data.UserName)
	fmt.Println("\n⚠️  IMPORTANT NOTES:")
	fmt.Println("1. Copy the access_token above and paste it in your config.json file")
	fmt.Println("2. Access tokens expire at 6:00 AM IST the next day")
	fmt.Println("3. You'll need to generate a new token daily before market opens")
	fmt.Println("4. Each request_token can only be used ONCE")
	fmt.Println("\n💡 TIP: Add this to your daily pre-market routine!\n")
}

func loadConfig() (*Config, error) {
	// Get the path of the current source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get current file path")
	}

	// Go up two levels from cmd/generate-token/main.go to find config.json
	configPath := filepath.Join(filepath.Dir(filename), "../../config.json")

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config.json at %s: %w\n\nPlease ensure config.json exists in the project root", configPath, err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config.json: %w", err)
	}

	if config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("api_key or api_secret missing in config.json")
	}

	return &config, nil
}