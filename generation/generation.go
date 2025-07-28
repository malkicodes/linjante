package generation

import (
	"math/rand/v2"
	"sync"

	"malki.codes/linjante/words"
)

type WordRole uint8

type Sentence struct {
	Sentence             string
	Subject              string
	Verb                 string
	Object               string
	PrepositionalPhrases []string
	Components           []string
}

type Generator struct {
	WordList  []words.Word
	WordRoles map[uint8][]string
	WordCount int

	nonLonPrepositions []string
}

func PickRandom[T any](slice []T) T {
	return slice[rand.IntN(len(slice))]
}

func NewGenerator() (Generator, error) {
	wordList, err := words.LoadWords()
	if err != nil {
		return Generator{}, err
	}

	wordRoles := make(map[uint8][]string, 0)

	for _, word := range wordList {
		for _, role := range word.Roles {
			wordList, prs := wordRoles[uint8(role)]

			if !prs {
				wordRoles[uint8(role)] = []string{word.Word}
			} else {
				wordRoles[uint8(role)] = append(wordList, word.Word)
			}
		}
	}

	prepositions := wordRoles[uint8(words.Preposition)]

	nonLon := make([]string, 0)
	i := 0

	for i < len(prepositions) {
		if prepositions[i] == "lon" {
			break
		}
		nonLon = append(nonLon, prepositions[i])
		i++
	}

	nonLon = append(nonLon, prepositions[i+1:]...)

	return Generator{
		WordList:           wordList,
		WordCount:          len(wordList),
		WordRoles:          wordRoles,
		nonLonPrepositions: nonLon,
	}, nil
}

func (g *Generator) GenerateContentWord() string {
	return PickRandom(g.WordRoles[uint8(words.Content)])
}

func (g *Generator) GenerateNoun() string {
	if rand.IntN(3) == 1 {
		// Uses a pronoun
		return PickRandom(g.WordRoles[uint8(words.Pronoun)])
	} else {
		word := g.GenerateContentWord()

		if rand.Float32() > 0.5 {
			word = word + " " + g.GenerateContentWord()
		}

		return word
	}
}

func (g *Generator) GenerateVerb() string {
	preverbWords := g.WordRoles[uint8(words.Preverb)]
	verb := g.GenerateContentWord()

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

	return verb
}

func (g *Generator) GenerateSentence() Sentence {
	components := make([]string, 0)

	// Subject
	subject := g.GenerateNoun()

	// Verb
	verb := g.GenerateVerb()

	// Object
	object := ""

	if rand.IntN(3) != 0 {
		// Has object
		object = g.GenerateNoun()
	}

	// Prepositions
	prepPhrases := make([]string, 0)
	usedLon := false

	if rand.IntN(3) == 0 {
		prepPhraseCount := rand.IntN(2) + 1
		for len(prepPhrases) < prepPhraseCount {
			preposition := PickRandom(g.nonLonPrepositions)

			if !usedLon && rand.IntN(2) == 0 {
				preposition = "lon"
				usedLon = true
			}

			prepPhrases = append(prepPhrases, preposition+" "+g.GenerateNoun())
		}
	}

	// Form sentence
	sentence := subject + " "

	if !(subject == "mi" || subject == "sina") {
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

	for _, prepPhrase := range prepPhrases {
		sentence += " " + prepPhrase
		components = append(components, prepPhrase)
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

func (g *Generator) GenerateSentences(count uint8) []Sentence {
	results := make([]Sentence, 0, count)
	var wg sync.WaitGroup

	ch := make(chan Sentence, count)

	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- g.GenerateSentence()
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
