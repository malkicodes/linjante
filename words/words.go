package words

import (
	"log"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const PATH_TO_LINKU_DATASET = "./sona"

type WordRole uint8

const (
	Particle WordRole = iota
	Content
	Preverb
	Preposition
	Pronoun
)

type WordData struct {
	Word           string
	Usage_category string
	Pu_verbatim    map[string]string
}

type Word struct {
	Word  string
	Roles []WordRole
}

type WordDictionary map[string][]string

func getUnique[T comparable](slice []T) []T {
	unique := make(map[T]bool)

	for _, v := range slice {
		_, prs := unique[v]

		if !prs {
			unique[v] = true
		}
	}

	keys := make([]T, 0, len(unique))
	for k := range unique {
		keys = append(keys, k)
	}

	return keys
}

func LoadDictionary() (WordDictionary, error) {
	dictionary := make(WordDictionary)

	contents, err := os.ReadFile(PATH_TO_LINKU_DATASET + "/words/source/definitions.toml")
	if err != nil {
		return nil, err
	}

	var rawDefinitions map[string]string
	err = toml.Unmarshal(contents, &rawDefinitions)

	if err != nil {
		return nil, err
	}

	for k, v := range rawDefinitions {
		dictionary[k] = strings.Split(v, "; ")
	}

	return dictionary, nil
}

func LoadWords() ([]Word, error) {
	words := make([]Word, 0)

	entries, err := os.ReadDir(PATH_TO_LINKU_DATASET + "/words/metadata")
	if err != nil {
		return nil, err
	}

	dictionary, err := LoadDictionary()
	if err != nil {
		return nil, err
	}

	for _, filename := range entries {
		contents, err := os.ReadFile(PATH_TO_LINKU_DATASET + "/words/metadata/" + filename.Name())
		if err != nil {
			return nil, err
		}

		var wordData WordData

		if toml.Unmarshal(contents, &wordData) != nil {
			return nil, err
		}

		// skip obscure words, ale is already in the database so also skip ali
		if wordData.Usage_category == "obscure" || wordData.Word == "ali" {
			continue
		}

		word := wordData.Word

		dictionaryEntry := dictionary[word]

		wordRoles := make([]WordRole, 0)

		if word == "mi" || word == "sina" || word == "ona" || word == "ni" {
			wordRoles = append(wordRoles, Pronoun)
		} else if word == "seme" {
			wordRoles = append(wordRoles, Particle)
		} else if len(dictionaryEntry) != 0 {
			for _, entry := range dictionaryEntry {
				// If any entry does not say that the word is a particle or interjection,
				if strings.HasPrefix(entry, "(particle)") || strings.HasPrefix(entry, "(interjection)") {
					wordRoles = append(wordRoles, Particle)
				} else if strings.HasPrefix(entry, "(preverb)") {
					wordRoles = append(wordRoles, Preverb)
				} else if strings.HasPrefix(entry, "(preposition)") {
					wordRoles = append(wordRoles, Preposition)
				} else {
					wordRoles = append(wordRoles, Content)
				}
			}
		}

		words = append(words, Word{
			Word:  word,
			Roles: getUnique(wordRoles),
		})
	}

	log.Printf("Imported %d words", len(words))

	return words, nil
}
