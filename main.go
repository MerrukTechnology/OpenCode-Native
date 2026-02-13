package main

import (
	"github.com/MerrukTechnology/OpenCode-Native/cmd"
	"github.com/MerrukTechnology/OpenCode-Native/internal/logging"
)

func main() {
	defer logging.RecoverPanic("main", func() {
		logging.ErrorPersist("Application terminated due to unhandled panic")
	})

	cmd.Execute()
}
