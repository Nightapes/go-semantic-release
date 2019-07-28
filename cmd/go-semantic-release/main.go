package main

import (
	"github.com/Nightapes/go-semantic-release/cmd/go-semantic-release/commands"
)

var version string

func main() {
	commands.Execute(version)
}
