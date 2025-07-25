package server

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"

	"linjante/words"
)

type Sentence struct {
	Sentence             string
	Subject              string
	Verb                 string
	Object               string
	PrepositionalPhrases []string
	Components           []string
}

func pickRandom[T any](slice []T) T {
	return slice[rand.IntN(len(slice))]
}

func getContentWord(contentWords []string) string {
	word := ""

	for word == "" || word == "ni" || word == "mi" || word == "sina" || word == "ona" {
		word = pickRandom(contentWords)
	}

	return word
}

func createNoun(contentWords []string) string {
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
		word := getContentWord(contentWords)

		if rand.Float32() > 0.5 {
			word = word + " " + pickRandom(contentWords)
		}

		return word
	}
}

func createSentence(wordRoles map[uint8][]string) Sentence {
	contentWords := wordRoles[uint8(words.Content)]
	preverbWords := wordRoles[uint8(words.Preverb)]
	prepositionWords := wordRoles[uint8(words.Preposition)]

	components := make([]string, 0)

	// Subject
	subject := createNoun(contentWords)

	// Verb
	verb := getContentWord(contentWords)

	hasNegation := false

	if rand.IntN(3) == 0 {
		// Use preverb
		preverb := pickRandom(preverbWords)

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
		object = createNoun(contentWords)
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
					preposition = pickRandom(prepositionWords)
				}
			}

			prepPhrases = append(prepPhrases, preposition+" "+createNoun(contentWords))
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

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s", r.Method, r.URL.EscapedPath())
	})
}

func HandleUserError(w http.ResponseWriter, err string) {
	w.WriteHeader(400)
	response, _ := json.Marshal(map[string]string{"error": err})

	w.Write(response)
}

func HandleServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	response, _ := json.Marshal(map[string]string{"error": err.Error()})

	w.Write(response)
}

func RootHandler(w http.ResponseWriter, r *http.Request, wordRoles map[uint8][]string) {
	response, _ := json.Marshal(map[string]any{
		"message": createSentence(wordRoles).Sentence,
		"up":      true,
	})

	w.Write(response)
}

func GenerateHandler(w http.ResponseWriter, r *http.Request, wordRoles map[uint8][]string) {
	countRaw := r.URL.Query().Get("count")

	var count uint8

	if countRaw == "" {
		count = 1
	} else {
		countInt, err := strconv.Atoi(countRaw)
		if err != nil {
			HandleUserError(w, "invalid count")
			return
		} else if countInt > 50 || countInt < 1 {
			HandleUserError(w, "count must be between 1 and 50")
			return
		} else {
			count = uint8(countInt)
		}
	}

	verboseRaw := r.URL.Query().Get("v")

	switch verboseRaw {
	case "", "false":
		sentences := make([]string, 0, count)

		for i := uint8(0); i < count; i++ {
			sentence := createSentence(wordRoles)
			sentences = append(sentences, sentence.Sentence)
		}

		response, err := json.Marshal(map[string]any{
			"sentences": sentences,
			"count":     count,
		})

		if err != nil {
			HandleServerError(w, err)
			return
		}

		w.Write(response)
	case "true":
		sentences := make([]map[string]any, 0, count)

		for i := uint8(0); i < count; i++ {
			sentence := createSentence(wordRoles)

			roles := map[string]any{
				"subject": sentence.Subject,
				"verb":    sentence.Verb,
				"object":  sentence.Object,
			}

			if sentence.Object == "" {
				roles["object"] = nil
			}

			if len(sentence.PrepositionalPhrases) > 0 {
				roles["prepositions"] = sentence.PrepositionalPhrases
			}

			sentences = append(sentences, map[string]any{
				"sentence":   sentence.Sentence,
				"components": sentence.Components,
				"roles":      roles,
			})
		}

		response, err := json.Marshal(map[string]any{
			"sentences": sentences,
			"count":     count,
		})

		if err != nil {
			HandleServerError(w, err)
			return
		}

		w.Write(response)
	default:
		HandleUserError(w, "invalid v")
		return
	}
}

func getRoleNames(roles []words.WordRole) []string {
	roleNames := make([]string, 0)

	for _, role := range roles {
		switch role {
		case words.Particle:
			roleNames = append(roleNames, "particle")
		case words.Content:
			roleNames = append(roleNames, "content")
		case words.Preverb:
			roleNames = append(roleNames, "preverb")
		default:
			roleNames = append(roleNames, "unknown")
		}
	}

	return roleNames
}

func WordsHandler(w http.ResponseWriter, r *http.Request, words []words.Word) {
	wordList := make(map[string]map[string]any)

	for _, word := range words {
		wordList[word.Word] = map[string]any{
			"word": word.Word,
			"role": getRoleNames(word.Roles),
		}
	}

	response, err := json.Marshal(wordList)

	if err != nil {
		HandleServerError(w, err)
		return
	}

	w.Write(response)
}

func RunServer(port int, words []words.Word) error {
	mux := http.NewServeMux()

	wordRoles := make(map[uint8][]string)

	for _, word := range words {
		for _, role := range word.Roles {
			wordList, prs := wordRoles[uint8(role)]

			if !prs {
				wordRoles[uint8(role)] = []string{word.Word}
			} else {
				wordRoles[uint8(role)] = append(wordList, word.Word)
			}
		}
	}

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()

		if path == "/" {
			RootHandler(w, r, wordRoles)
		} else {
			w.WriteHeader(404)
			response, _ := json.Marshal(map[string]string{"error": "Not Found"})

			w.Write(response)
		}
	}))

	mux.Handle("/gen", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GenerateHandler(w, r, wordRoles)
	}))

	mux.Handle("/words", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WordsHandler(w, r, words)
	}))

	return http.ListenAndServe(":"+strconv.Itoa(port), LoggerMiddleware(mux))
}
