package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
)

var state State

type State struct {
	EnvironmentImageCache string
}

func LoadState() {
	file, err := os.Open(path.Join(workdir, ".containme", "state"))
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	decoder.Decode(&state)
}

func WriteState() {
	data, err := json.Marshal(&state)
	if err != nil {
		return
	}
	ioutil.WriteFile(path.Join(workdir, ".containme", "state"), data, 0644)
}
