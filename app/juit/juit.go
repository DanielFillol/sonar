package juit

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
)

// Search calls the JUIT API to search for jurisprudence.
func Search(query string) (*JurisprudenceResponse, error) {
	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		return nil, errors.New("error loading .env file: " + err.Error())
	}

	// Get the required credentials from environment variables.
	user := os.Getenv("JUIT_USER")
	pass := os.Getenv("JUIT_PASS")
	if user == "" || pass == "" {
		return nil, errors.New("missing JUIT_USER or JUIT_PASS in environment")
	}

	// API endpoint with query parameters already embedded.
	url := "https://api.juit.io:8080/v1/data-products/search/jurisprudence?query=" + query + "&search_on=headnote&owner=" + user

	// Create an HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.New("error creating request: " + err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	// Encode credentials in Base64 for Basic Authentication
	auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))

	// Set required headers for Basic Authentication
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.New("error making request: " + err.Error())
	}
	defer res.Body.Close()

	// Read response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("error reading response: " + err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("error on response, current status: " + res.Status)
	}

	// Unmarshal JSON response into the model.
	var result JurisprudenceResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func ReturnAsText(juris []Jurisprudence) *string {
	var r string

	for _, i := range juris {
		var entry string

		// Helper function to append fields only if they're not empty
		appendIfNotEmpty := func(label, value string) {
			if value != "" {
				entry += fmt.Sprintf("**%s:** %s\n\n", label, value)
			}
		}

		entry += "---\n\n" // Markdown separator for readability

		appendIfNotEmpty("Tipo de Documento", i.DocumentType)
		appendIfNotEmpty("Grau", i.Degree)
		appendIfNotEmpty("Classe/Assunto", i.ClassSubject)
		appendIfNotEmpty("️Juiz", i.Judge)
		appendIfNotEmpty("Órgão Julgador", i.JudgmentBody)
		appendIfNotEmpty("Data do Julgamento", i.JudgmentDate)
		appendIfNotEmpty("Data da Publicação", i.PublicationDate)
		appendIfNotEmpty("Ementa", i.Headnote)
		appendIfNotEmpty("Texto Completo", i.FullText)
		appendIfNotEmpty("Referência do Processo", i.LawsuitReference)

		r += entry
	}

	return &r
}
