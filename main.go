package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "chat",
		Short: "Get answers using OpenAI's GPT",
		Long:  "Get answers to your queries using OpenAI's GPT-3.5-turbo model via CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Please provide a query.")
				return
			}
			query := args[0]
			answer, err := getChatResponse(query)
			if err != nil {
				fmt.Printf("Error getting response: %v\n", err)
				return
			}
			fmt.Println("Answer:", answer)
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getChatResponse(query string) (string, error) {
	apiKey := os.Getenv("OPEN_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPEN_API_KEY not set in .env file")
	}
	fmt.Println("Using API Key:", apiKey)
	endpoint := "https://api.openai.com/v1/chat/completions"
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": query},
		},
		"max_tokens": 50,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("Response Body:", string(body))

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", fmt.Errorf("you have exceeded your API quota, please check your plan and billing details")
		}
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format: %s", body)
	}

	answer, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: %s", body)
	}

	return answer, nil
}
