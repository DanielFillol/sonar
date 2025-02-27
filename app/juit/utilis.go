package juit

import (
	"log"
	"sonar/app/gpt"
	"strconv"
	"time"
)

// JurisprudenceResponse represents the structure of the API response.
type JurisprudenceResponse struct {
	Total         int                 `json:"total"`
	Size          int                 `json:"size"`
	NextPageToken string              `json:"next_page_token"`
	SearchInfo    SearchInfo          `json:"search_info"`
	Items         []JurisprudenceItem `json:"items"`
}

// SearchInfo holds metadata about the search.
type SearchInfo struct {
	SearchID        string `json:"search_id"`
	ElapsedTimeInMs int    `json:"elapsed_time_in_ms"`
}

// JurisprudenceItem represents a single jurisprudence record.
type JurisprudenceItem struct {
	ID                   string   `json:"id"`
	JuitID               string   `json:"juit_id"`
	Title                string   `json:"title"`
	Headnote             string   `json:"headnote"`
	FullText             *string  `json:"full_text"` // Nullable field.
	CnjUniqueNumber      string   `json:"cnj_unique_number"`
	OrderDate            string   `json:"order_date"`
	JudgmentDate         string   `json:"judgment_date"`
	PublicationDate      string   `json:"publication_date"`
	ReleaseDate          *string  `json:"release_date"`   // Nullable.
	SignatureDate        *string  `json:"signature_date"` // Nullable.
	CourtCode            string   `json:"court_code"`
	Degree               string   `json:"degree"`
	ProcessOriginState   *string  `json:"process_origin_state"` // Nullable.
	District             string   `json:"district"`
	DocumentMatterList   []string `json:"document_matter_list"`
	ProcessClassNameList []string `json:"process_class_name_list"`
	JudgmentBody         string   `json:"judgment_body"`
	Trier                string   `json:"trier"`
	DocumentType         string   `json:"document_type"`
	JusticeType          string   `json:"justice_type"`
	RimorURL             string   `json:"rimor_url"`
}

// formatDate converts a timestamp string (RFC3339) into "DD/MM/YYYY" format.
func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "" // Handle the error appropriately (e.g., logging)
	}
	return t.Format("02/01/2006") // DD/MM/YYYY format
}

type Jurisprudence struct {
	DocumentType     string
	Degree           string
	ClassSubject     string
	Judge            string
	JudgmentBody     string
	PublicationDate  string
	JudgmentDate     string
	Headnote         string
	FullText         string
	LawsuitReference string
}

func (j *JurisprudenceResponse) GetJurisprudence() []Jurisprudence {
	var r []Jurisprudence

	for _, i := range j.Items {
		var classSubject string
		for k, cS := range i.ProcessClassNameList {
			classSubject += cS
			if k != len(i.ProcessClassNameList)-1 {
				classSubject += " / "
			}
		}

		// Handle nil FullText safely
		fullText := ""
		if i.FullText != nil {
			fullText = *i.FullText
		}

		r = append(r, Jurisprudence{
			DocumentType:     i.DocumentType,
			Degree:           i.Degree,
			ClassSubject:     classSubject,
			Judge:            i.Trier,
			JudgmentBody:     i.JudgmentBody,
			PublicationDate:  formatDate(i.PublicationDate),
			JudgmentDate:     formatDate(i.JudgmentDate),
			Headnote:         i.Headnote,
			FullText:         fullText,
			LawsuitReference: i.CnjUniqueNumber,
		})
	}

	return r
}

func ShouldCallJurisprudencia(system, query string) (bool, error) {
	log.Println("Verificando eligibilidade para jurisprudência")
	relevant, err := gpt.Search(system, query, "gpt-4o-mini")
	if err != nil {
		return false, err
	}

	log.Println("Elegibilidade para jurisprudência: " + relevant.Choices[0].Message.Content)
	if relevant.Choices[0].Message.Content == "Sim" {
		return true, nil
	} else {
		return false, nil
	}
}

func CreateQueryForJurisprudencia(system, query string) (*string, error) {
	log.Println("Criando o prompt ideal para pesquisar a jurisprudência")

	responseKeyWords, err := gpt.Search(system, "transforme esse prompt: "+query+" em uma excelente query para API.", "gpt-4o-mini")
	if err != nil {
		return nil, err
	}

	log.Println("Prompt para jurisprudência: " + responseKeyWords.Choices[0].Message.Content)
	return &responseKeyWords.Choices[0].Message.Content, nil
}

func CallAPIjurisprudencia(query string) (*string, error) {
	log.Println("Consultando jurisprudências")

	jt, err := Search(query)
	if err != nil {
		return nil, err
	}

	jurisInit := jt.GetJurisprudence()
	log.Println("Jurisprudências encontradas: " + strconv.Itoa(len(jurisInit)))

	r := ReturnAsText(jurisInit)
	return r, nil
}
