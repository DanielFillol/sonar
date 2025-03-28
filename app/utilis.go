package app

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sonar/app/deepseek"
	"sonar/app/gpt"
	"sonar/app/juit"
	"sonar/app/perplexity"
	"strings"
	"time"
)

var (
	gptRelevantCaseLaw string

	gptSimplePrompt string

	gptClassifier string

	gptAuthors string

	perplexitySearcher string

	perplexityLaw string

	gptSpecialist string
)

func populateVars() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	// Retrieve the prompts from the environment.
	gptRelevantCaseLaw = os.Getenv("gptRelevantCaseLaw")
	gptSimplePrompt = os.Getenv("gptSimplePrompt")
	gptClassifier = os.Getenv("gptClassifier")
	gptAuthors = os.Getenv("gptAuthors")
	perplexitySearcher = os.Getenv("perplexitySearcher")
	perplexityLaw = os.Getenv("perplexityLaw")
	gptSpecialist = os.Getenv("gptSpecialist")

	return nil
}

func createFile(content string) error {
	file, err := os.Create("resposta.md")
	if err != nil {
		return errors.New("Erro ao criar o arquivo:" + err.Error())
	}

	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return errors.New("Erro ao escrever o arquivo:" + err.Error())
	}

	return nil
}

func finalAnswer(llm, system, field, authors, quotes, linkQuotes, laws, linkLaws, prompt, juris string, stream bool) (string, error) {
	log.Println("Gerando texto final...")
	specialistInput := "O ramo do direito é:\n" + field +
		"\nOs doutrinadores relevantes são:\n" + authors +
		"\nA doutrina relevante é:\n" + quotes +
		"\nOs links relevantes são:\n" + linkQuotes +
		"\nAs leis relevantes são:\n" + laws +
		"\nOs links legislativos relevantes são relevantes são:\n" + linkLaws +
		"\nO prompt original do usuário é:\n" + prompt +
		"\nAs Jurisprudências retornadas são:\n" + juris

	var responseBuilder strings.Builder

	if llm == "deepseek" {
		specialist, err := deepseek.Search(system, specialistInput, "deepseek-reasoner")
		if err != nil {
			return "", errors.New("Erro ao processar a resposta final:" + err.Error())
		}

		return specialist.Choices[0].Message.Content, nil
	} else if llm == "gpt-mini" {
		specialist, err := gpt.Search(system, specialistInput, "gpt-4o-mini")
		if err != nil {
			return "", errors.New("Erro ao processar a resposta final:" + err.Error())
		}

		return specialist.Choices[0].Message.Content, nil
	} else if llm == "gpt-full" {
		if stream {
			fmt.Println("\n\033[1;36m[RESPOSTA GERADA]\033[0m\n")

			err := gpt.StreamSearch(
				system,
				specialistInput,
				"gpt-4-1106-preview", // Modelo que suporta streaming
				func(content string) {
					fmt.Print(content)
					responseBuilder.WriteString(content)
				},
			)
			if err != nil {
				return "", errors.New("Erro ao processar a resposta final:" + err.Error())
			}
			fmt.Println("\n\033[1;36m----------------------------------------\033[0m")
			return responseBuilder.String(), err
		} else {
			specialist, err := gpt.Search(system, specialistInput, "chatgpt-4o-latest")
			if err != nil {
				return "", errors.New("Erro ao processar a resposta final:" + err.Error())
			}

			return specialist.Choices[0].Message.Content, nil
		}

	} else {
		return "", errors.New("Erro ao processar a resposta final:" + llm)
	}

}

// GetPromptResponse llm must be:
//   - deepseek
//   - gpt-mini
//   - gpt-full
func GetPromptResponse(llm, prompt string) (*string, error) {
	err := populateVars()
	if err != nil {
		return nil, err
	}

	// Verifies if the prompt is eligible for case law search
	relevant, err := juit.ShouldCallJurisprudencia(gptRelevantCaseLaw, prompt)
	if err != nil {
		return nil, err
	}

	// Get Relevant Case Law
	var promptJuris *string
	var juris *string
	if relevant {
		promptJuris, err = juit.CreateQueryForJurisprudencia(gptSimplePrompt, prompt)
		if err != nil {
			return nil, err
		}

		juris, err = juit.CallAPIjurisprudencia(*promptJuris)
		if err != nil {
			return nil, err
		}
	}

	// Classify the legal field of the prompt.
	field, err := gpt.ClassifyLawField(gptClassifier, prompt)
	if err != nil {
		return nil, err
	}

	// Get the main doctrinal experts for the classified legal field.
	authors, err := gpt.GetRelevantAuthors(gptAuthors, *field)
	if err != nil {
		return nil, err
	}

	// Search for relevant citations from the doctrinal experts.
	if authors == nil || promptJuris == nil {
		return nil, errors.New("autores não foram localizados ou prompt para API de jurisprudência não foi localizado")
	}

	doctrines, err := perplexity.SearchForQuotes(perplexitySearcher, *field, *authors, *promptJuris, prompt)
	if err != nil {
		return nil, err
	}

	// Search for relevant laws in official sites
	laws, err := perplexity.SearchForLaws(perplexityLaw, *field, prompt, *promptJuris)
	if err != nil {
		return nil, err
	}

	// Structure the final answer by integrating all the gathered information.
	answer, err := finalAnswer(llm, gptSpecialist, *field, *authors, doctrines.Response, doctrines.Links, laws.Response, laws.Links, prompt, *juris, false)
	if err != nil {
		return nil, err
	}
	// Exibe a resposta de forma streamada
	fmt.Println("\n\033[1;36m[RESPOSTA GERADA]\033[0m\n")
	for _, line := range strings.Split(answer, "\n") {
		fmt.Print(line + "\n")
		time.Sleep(200 * time.Millisecond) // Ajuste o delay conforme necessário
	}
	fmt.Println("\n\033[1;36m----------------------------------------\033[0m")

	err = createFile(answer)
	if err != nil {
		return nil, err
	}

	log.Println("A resposta foi salva no arquivo resposta.md")
	return &answer, nil
}
