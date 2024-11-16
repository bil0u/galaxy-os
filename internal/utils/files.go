package utils

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// `LocalFile` represents a file on the local filesystem
type LocalFile struct {
	Directory string
	Filename  string
}

// `FullPath` returns the full path of a file
func (f LocalFile) FullPath() string {
	return fmt.Sprintf("%s/%s", f.Directory, f.Filename)
}

// `Validate` validates a file configuration
func (f LocalFile) Validate() error {

	if f.Filename == "" {
		return fmt.Errorf("no configuration file provided")
	}

	if f.Directory == "" {
		f.Directory = "."
	}

	return nil
}

// `decodeTOML` decodes a TOML file into a configuration object
func decodeTOML[T any](path string, cfg *T) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config: %w", err)
	}
	slog.Info("File content: ", slog.Any("file", *file))
	defer file.Close()
	if err = toml.NewDecoder(file).Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}
	return nil
}

// `LoadTomlFilesIn` loads a list of files into a configuration object
func LoadTomlFilesIn[T any](cfg *T, files ...LocalFile) error {
	if cfg == nil {
		return fmt.Errorf("no configuration object provided")
	}
	if len(files) == 0 {
		return fmt.Errorf("no files to load")
	}

	errs := ManyErrors{}

	for _, f := range files {
		slog.Info("Loading TOML file: ", slog.Any("file", f.FullPath()))
		filePath := f.FullPath()
		err := decodeTOML(filePath, cfg)
		if errs.Add(err) {
			continue
		}
	}

	return errs.ToError()
}
