package sdk

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Find[T any](ss []T, test func(T) bool) (ret *T) {
	for _, s := range ss {
		if test(s) {
			return &s
		}
	}
	return nil
}

func IndexOf[T comparable](ss []T, s T) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}

func Contains[T comparable](ss []T, s T) bool {
	return IndexOf(ss, s) > -1
}

// loadFromFile loads a configuration file into a struct
func LoadFromFile[T any](path string, cfg *T) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()
	if err = toml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}
	return nil
}

func RemoveWithoutOrder[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
