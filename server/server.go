package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"malki.codes/linjante/generation"
	"malki.codes/linjante/server/errors"
	"malki.codes/linjante/server/middleware"
	"malki.codes/linjante/words"
)

func RootHandler(w http.ResponseWriter, r *http.Request, g *generation.Generator) {
	response, _ := json.Marshal(map[string]any{
		"message": g.GenerateSentence().Sentence,
		"words":   g.WordCount,
		"up":      true,
	})

	w.WriteHeader(200)
	w.Write(response)
}

func GenerateHandler(w http.ResponseWriter, r *http.Request, g *generation.Generator) {
	countRaw := r.URL.Query().Get("count")

	var count uint8

	if countRaw == "" {
		count = 1
	} else {
		countInt, err := strconv.Atoi(countRaw)
		if err != nil {
			errors.HandleUserError(w, "invalid count")
			return
		} else if countInt > 50 || countInt < 1 {
			errors.HandleUserError(w, "count must be between 1 and 50")
			return
		} else {
			count = uint8(countInt)
		}
	}

	verbose := r.URL.Query().Get("v")

	sentences := g.GenerateSentences(count)

	var data any

	switch verbose {
	case "", "false":
		sentenceList := make([]string, 0, count)

		for _, v := range sentences {
			sentenceList = append(sentenceList, v.Sentence)
		}

		data = sentenceList
	case "true":
		sentenceList := make([]map[string]any, 0, count)

		for _, v := range sentences {
			roles := map[string]any{
				"subject": v.Subject,
				"verb":    v.Verb,
			}

			if v.Object != "" {
				roles["object"] = v.Object
			}

			if len(v.PrepositionalPhrases) > 0 {
				roles["prepositions"] = v.PrepositionalPhrases
			}

			sentenceList = append(sentenceList, map[string]any{
				"sentence":  v.Sentence,
				"compoents": v.Components,
				"roles":     roles,
			})
		}

		data = sentenceList

	default:
		errors.HandleUserError(w, "invalid v")
		return
	}

	response, err := json.Marshal(map[string]any{
		"count":     count,
		"sentences": data,
	})

	if err != nil {
		errors.HandleServerError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(response)
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

func WordsHandler(w http.ResponseWriter, r *http.Request, g *generation.Generator) {
	wordList := make(map[string]map[string]any)

	for _, word := range g.WordList {
		wordList[word.Word] = map[string]any{
			"word": word.Word,
			"role": getRoleNames(word.Roles),
		}
	}

	response, err := json.Marshal(wordList)

	if err != nil {
		errors.HandleServerError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Write(response)
}

func RunServer(port int) error {
	mux := http.NewServeMux()

	generator, err := generation.NewGenerator()
	if err != nil {
		return err
	}

	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()

		if path == "/" {
			RootHandler(w, r, &generator)
		} else {
			errors.HandleNotFoundError(w)
		}
	}))

	mux.Handle("GET /gen", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()

		if path == "/gen" {
			GenerateHandler(w, r, &generator)
		} else {
			errors.HandleNotFoundError(w)
		}
	}))

	mux.Handle("GET /words", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()

		if path == "/gen" {
			WordsHandler(w, r, &generator)
		} else {
			errors.HandleNotFoundError(w)
		}
	}))

	return http.ListenAndServe(":"+strconv.Itoa(port), middleware.LoggerMiddleware(middleware.RateLimitMiddleware(mux)))
}
