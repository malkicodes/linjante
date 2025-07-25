package generation

import (
	"linjante/words"
	"math/rand/v2"
	"strings"
	"sync"
)

type Sentence struct {
	Sentence             string
	Subject              string
	Verb                 string
	Object               string
	PrepositionalPhrases []string
	Components           []string
}

func PickRandom[T any](slice []T) T {
	return slice[rand.IntN(len(slice))]
}

func GenerateContentWord(contentWords []string) string {
	word := ""

	for word == "" || word == "ni" || word == "mi" || word == "sina" || word == "ona" {
		word = PickRandom(contentWords)
	}

	return word
}

func GenerateNoun(contentWords []string) string {
	if rand.IntN(3) == 1 {
		// Uses a pronoun
		pronoun := rand.IntN(4)

		switch pronoun {
		case 0:
			return "mi"
		case 1:
			return "sina"
		case 2:
			return "ona"
		default:
			return "ni"
		}
	} else {
		word := GenerateContentWord(contentWords)

		if rand.Float32() > 0.5 {
			word = word + " " + PickRandom(contentWords)
		}

		return word
	}
}

func GenerateSentence(wordRoles map[uint8][]string) Sentence {
	contentWords := wordRoles[uint8(words.Content)]
	preverbWords := wordRoles[uint8(words.Preverb)]
	prepositionWords := wordRoles[uint8(words.Preposition)]

	components := make([]string, 0)

	// Subject
	subject := GenerateNoun(contentWords)

	// Verb
	verb := GenerateContentWord(contentWords)

	hasNegation := false

	if rand.IntN(3) == 0 {
		// Use preverb
		preverb := PickRandom(preverbWords)

		if rand.IntN(4) == 0 {
			hasNegation = true
			preverb += " ala"
		}

		verb = preverb + " " + verb
	}

	if (!hasNegation) && rand.IntN(4) == 0 {
		verb += " ala"
	}

	// Object
	object := ""

	if rand.IntN(3) != 0 {
		// Has object
		object = GenerateNoun(contentWords)
	}

	// Form sentence
	sentence := subject + " "

	if subject != "mi" && subject != "sina" {
		sentence += "li "
	}

	sentence += verb

	if object != "" {
		sentence += " e " + object
	}

	components = append(components, subject)
	components = append(components, verb)

	if object != "" {
		components = append(components, object)
	}

	// Prepositions
	prepPhrases := make([]string, 0)
	usedLon := false

	if rand.IntN(3) == 0 {
		prepPhraseCount := rand.IntN(2) + 1
		for len(prepPhrases) < prepPhraseCount {
			preposition := ""

			if !usedLon && rand.IntN(2) == 0 {
				preposition = "lon"
				usedLon = true
			} else {
				for preposition == "" || preposition == "lon" {
					preposition = PickRandom(prepositionWords)
				}
			}

			prepPhrases = append(prepPhrases, preposition+" "+GenerateNoun(contentWords))
		}
	}

	if len(prepPhrases) > 0 {
		prepPhrasesString := strings.Join(prepPhrases, " ")
		sentence += " " + prepPhrasesString

		components = append(components, prepPhrases...)
	}

	return Sentence{
		Sentence:             sentence,
		Subject:              subject,
		Verb:                 verb,
		Object:               object,
		PrepositionalPhrases: prepPhrases,
		Components:           components,
	}
}

func GenerateSentences(count uint8, wordRoles map[uint8][]string) []Sentence {
	results := make([]Sentence, 0, count)
	var wg sync.WaitGroup

	ch := make(chan Sentence, count)

	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ch <- GenerateSentence(wordRoles)
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		results = append(results, v)
	}

	return results
}
