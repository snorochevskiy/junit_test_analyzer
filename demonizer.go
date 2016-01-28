package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	envVarName  = "_JUTRA_DAEMON"
	envVarValue = "1"
)

func Reborn() (err error) {
	if os.Getenv(envVarName) != envVarValue {
		var path string
		if path, err = filepath.Abs(os.Args[0]); err != nil {
			log.Fatalln(err)
			return
		}
		cmd := exec.Command(path, os.Args[1:]...)
		envVar := fmt.Sprintf("%s=%s", envVarName, envVarValue)
		cmd.Env = append(os.Environ(), envVar)
		if err = cmd.Start(); err != nil {
			log.Fatalln(err)
			return
		}
		os.Exit(0)
	}
	return
}
