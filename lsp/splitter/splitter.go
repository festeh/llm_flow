package splitter

import (
	"log"
	"strings"
)

type Splitter int

type SplitFn func(*map[string]string) error

const (
	FimNaive Splitter = iota
	Fim
	Chat
)

func New(model string, override *Splitter) Splitter {
	if override != nil {
		return *override
	}
	switch model {
	case "codestral_latest":
		return FimNaive
	}
	return FimNaive
}

func GetFimNaiveSplitter(text string) SplitFn {
	parts := strings.Split(text, "<|FIM|>")
	var prefix, suffix string
	if len(parts) != 2 {
		log.Println("error splitting text into prefix and suffix")
		prefix = text
	} else {
		prefix = parts[0]
		suffix = parts[1]
	}
	log.Printf("prefix: %s", prefix)
	log.Printf("suffix: %s", suffix)
	return func(data *map[string]string) error {
		(*data)["prefix"] = prefix
		(*data)["suffix"] = suffix
		return nil
	}
}
