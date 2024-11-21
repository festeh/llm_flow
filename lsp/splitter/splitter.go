package splitter

import (
	"github.com/charmbracelet/log"
	"github.com/festeh/llm_flow/lsp/constants"
	"strings"
)

type Splitter int

type SplitFn func(*map[string]interface{}) error

type PrefixSuffix struct {
  Prefix string
  Suffix string
}

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
	parts := strings.Split(text, constants.FIM_TOKEN)
	var prefix, suffix string
	if len(parts) != 2 {
		log.Error("error splitting text into prefix and suffix")
		prefix = text
	} else {
		prefix = parts[0]
		suffix = parts[1]
	}
	log.Debug("", "prefix", prefix)
	log.Debug("", "suffix", suffix)
	return func(data *map[string]interface{}) error {
		(*data)["prompt"] = prefix
		(*data)["suffix"] = suffix
		return nil
	}
}
