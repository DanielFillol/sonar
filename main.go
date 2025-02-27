package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sonar/app"
	"time"
)

var (
	// The original user prompt.
	prompt = "meu voo da latam atrasou por 4 horas. tenho alguma direito?"
)

func main() {
	// Check if the prompt is empty; if so, collect the prompt from the terminal.
	if prompt == "" {
		fmt.Print("Pergunte ao which: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			prompt = scanner.Text()
		}
	}
	start := time.Now()

	// llm pode ser "deepseek", "gpt-mini" ou "gpt-full"
	err := app.GetPromptResponse("deepseek", prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(time.Since(start))
}
