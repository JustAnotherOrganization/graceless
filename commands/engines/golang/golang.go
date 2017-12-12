package golang

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// TODO: call go fmt, vet, etc

// ErrIllegalImport error for passing an illegal import
var ErrIllegalImport = errors.New("Illegal Import")

type (
	// EngineCommand gives a bot access to run something in Go.
	EngineCommand struct {
		*commands.Index
	}
)

// NewEngineCommand returns a new graceless.Command
func NewEngineCommand() *EngineCommand {
	return &EngineCommand{
		Index: &commands.Index{
			CmdName:     "hidden",
			CmdPerms:    []string{"go"},
			CommandType: commands.EngineCommand,
		},
	}
}

// Match matches a string against the go run command.
func (*EngineCommand) Match(str string) (string, bool) {
	affix := "```"
	if strings.HasPrefix(str, affix) && strings.HasSuffix(str, affix) {
		_str := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(str, affix), affix))
		if strings.HasPrefix(_str, "package main") {
			return _str, true
		}
		// TODO: improve functionality to allow us to run functions without
		// package/import name, etc (requires building the rest of the file
		// using a template and go imports, etc).
	}

	return str, false
}

// HelpShort is a dummy function because this command is hidden.
func (*EngineCommand) HelpShort() string {
	return ""
}

// Help is a dummy function because this command is hidden.
func (*EngineCommand) Help() string {
	return ""
}

// Exec performs go operations.
func (g *EngineCommand) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if err := filterImports(cmdStr); err != nil {
		return err
	}

	// Write a file for our go code to be built from.
	// FIXME: this belongs in it's own container with the go runtime.
	fileName := fmt.Sprintf("/tmp/graceless_%s.go", uuid.NewV4().String())
	if err := ioutil.WriteFile(fileName, []byte(cmdStr), 0770); err != nil {
		return errors.Wrap(err, "ioutil.WriteFile")
	}

	// FIXME: see above.
	// TODO: detect go location during startup, disable command (allowing bot
	// to start without it) if not found.
	cmd := exec.Command("/usr/bin/go", "run", fileName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "cmd.Run")
	}

	return acc.SendMessage(fmt.Sprintf("```%v```", out.String()), ev.Origin)
}

func filterImports(body string) error {
	scanner := bufio.NewScanner(strings.NewReader(body))
	isWhiteListed := func(pkg string) bool {
		found := false
		for _, pk := range whitelist {
			if pkg == pk {
				found = true
			}
		}

		return found
	}

	checkingImports := false
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			if strings.HasPrefix(text, "import") {
				if strings.Contains(text, "(") {
					checkingImports = true
					continue
				}
				text := strings.TrimSpace(strings.TrimPrefix(text, "import"))
				if !isWhiteListed(text) {
					return errors.Wrap(ErrIllegalImport, text)
				}
			}
			if checkingImports {
				if strings.HasPrefix(text, ")") {
					checkingImports = false
					continue
				}
				text = strings.TrimSpace(text)
				if !isWhiteListed(text) {
					return errors.Wrap(ErrIllegalImport, text)
				}
			}
		}
	}

	return nil
}
