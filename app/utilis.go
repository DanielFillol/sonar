package app

import (
	"errors"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sonar/app/deepseek"
	"sonar/app/gpt"
	"sonar/app/juit"
	"sonar/app/perplexity"
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

func finalAnswer(llm, system, field, authors, quotes, linkQuotes, laws, linkLaws, prompt, juris string) (string, error) {
	log.Println("Gerando texto final...")
	specialistInput := "O ramo do direito é:\n" + field +
		"\nOs doutrinadores relevantes são:\n" + authors +
		"\nA doutrina relevante é:\n" + quotes +
		"\nOs links relevantes são:\n" + linkQuotes +
		"\nAs leis relevantes são:\n" + laws +
		"\nOs links legislativos relevantes são relevantes são:\n" + linkLaws +
		"\nO prompt original do usuário é:\n" + prompt +
		"\nAs Jurisprudências retornadas são:\n" + juris

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
		specialist, err := gpt.Search(system, specialistInput, "chatgpt-4o-latest")
		if err != nil {
			return "", errors.New("Erro ao processar a resposta final:" + err.Error())
		}

		return specialist.Choices[0].Message.Content, nil
	} else {
		return "", errors.New("Erro ao processar a resposta final:" + llm)
	}

}

// GetPromptResponse llm must be:
//   - deepseek
//   - gpt-mini
//   - gpt-full
func GetPromptResponse(llm, prompt string) error {
	err := populateVars()
	if err != nil {
		return err
	}

	// Verifies if the prompt is eligible for case law search
	relevant, err := juit.ShouldCallJurisprudencia(gptRelevantCaseLaw, prompt)
	if err != nil {
		return err
	}

	// Get Relevant Case Law
	var promptJuris *string
	var juris *string
	if relevant {
		promptJuris, err = juit.CreateQueryForJurisprudencia(gptSimplePrompt, prompt)
		if err != nil {
			return err
		}

		juris, err = juit.CallAPIjurisprudencia(*promptJuris)
		if err != nil {
			return err
		}
	}

	// Classify the legal field of the prompt.
	field, err := gpt.ClassifyLawField(gptClassifier, prompt)
	if err != nil {
		return err
	}

	// Get the main doctrinal experts for the classified legal field.
	authors, err := gpt.GetRelevantAuthors(gptAuthors, *field)
	if err != nil {
		return err
	}

	// Search for relevant citations from the doctrinal experts.
	if authors == nil || promptJuris == nil {
		return errors.New("autores não foram localizados ou prompt para API de jurisprudência não foi localizado")
	}

	doctrines, err := perplexity.SearchForQuotes(perplexitySearcher, *field, *authors, *promptJuris, prompt)
	if err != nil {
		return err
	}

	// Search for relevant laws in official sites
	laws, err := perplexity.SearchForLaws(perplexityLaw, *field, prompt, *promptJuris)
	if err != nil {
		return err
	}

	// Structure the final answer by integrating all the gathered information.
	answer, err := finalAnswer(llm, gptSpecialist, *field, *authors, doctrines.Response, doctrines.Links, laws.Response, laws.Links, prompt, *juris)
	if err != nil {
		return err
	}

	err = createFile(answer)
	if err != nil {
		return err
	}

	log.Println("A resposta foi salva no arquivo resposta.md")
	return nil
}
