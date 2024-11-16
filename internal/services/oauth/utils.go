package oauth

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func writeError(w http.ResponseWriter, text string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(text + ": " + err.Error()))
}

func randStr(n int) string {
	var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

func formatData(data any) []byte {
	var formatted []byte
	formatted, err := json.MarshalIndent(data, "<br />", "&ensp;")
	if err != nil {
		slog.Error("Failed to format data", slog.Any("data", data), slog.Any("error", err))
		return nil
	}
	return formatted
}
