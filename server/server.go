package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"linjante/generation"
	"linjante/words"
)

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
		"message": generation.CreateSentence(wordRoles).Sentence,
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
			sentence := generation.CreateSentence(wordRoles)
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
			sentence := generation.CreateSentence(wordRoles)

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
