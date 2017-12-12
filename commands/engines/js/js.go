package js

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/robertkrimen/otto"

	"github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
)

type (
	// EngineCommand gives a bot access to run something in JS.
	EngineCommand struct {
		*commands.Index
	}

	valueMap map[string]interface{}
)

// NewEngineCommand returns a new graceless.Command
func NewEngineCommand() *EngineCommand {
	return &EngineCommand{
		Index: &commands.Index{
			CmdName:     "hidden",
			CmdPerms:    []string{"js"},
			CommandType: commands.EngineCommand,
		},
	}
}

// Match matches a string against the js run command.
func (*EngineCommand) Match(str string) (string, bool) {
	affix := "```"
	if strings.HasPrefix(str, affix) && strings.HasSuffix(str, affix) {
		_str := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(str, affix), affix))
		if strings.HasPrefix(_str, "js") {
			return strings.TrimSpace(strings.TrimPrefix(_str, "js")), true
		}
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

// Exec perms js operations.
func (ec *EngineCommand) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	in, out, script, err := ec.scan(cmdStr)
	if err != nil {
		return err
	}

	if script == "" {
		// TODO: better help?
		return acc.SendMessage(fmt.Sprintf("No script provided"), ev.Origin)
	}

	vm := otto.New()
	for k, v := range in {
		vm.Set(k, v)
	}

	if len(out) > 0 {
		if _, err := vm.Run(script); err != nil {
			return err
		}

		var msg string
		for _, v := range out {
			value, err := vm.Get(v)
			if err != nil {
				return err
			}

			// TODO: matshal this
			msg = fmt.Sprintf("%s\n```%v```", msg, value)
		}

		return acc.SendMessage(msg, ev.Origin)
	}

	value, err := vm.Run(script)
	if err != nil {
		return err
	}

	return acc.SendMessage(fmt.Sprintf("```%v```", value), ev.Origin)
}

func (*EngineCommand) scan(body string) (valueMap, []string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(body))

	_in := false
	_out := false
	in := make(valueMap)
	out := []string{}
	script := []string{}

	splitValue := func(str string) (string, interface{}) {
		fields := strings.Split(str, ":")
		if len(fields) != 2 {
			return str, nil
		}

		trimfunc := func(str string) string {
			str = strings.TrimSpace(str)
			if affix := "`"; strings.HasPrefix(str, affix) && strings.HasSuffix(str, affix) {
				return strings.Trim(str, affix)
			}
			if affix := `"`; strings.HasPrefix(str, affix) && strings.HasSuffix(str, affix) {
				return strings.Trim(str, affix)
			}
			if affix := "'"; strings.HasPrefix(str, affix) && strings.HasSuffix(str, affix) {
				return strings.Trim(str, affix)
			}

			return str
		}

		key := trimfunc(fields[0])
		_value := trimfunc(fields[1])

		if b, err := strconv.ParseBool(_value); err == nil {
			return key, b
		}
		if f, err := strconv.ParseFloat(_value, 64); err == nil {
			return key, f
		}
		if u, err := strconv.ParseUint(_value, 0, 64); err == nil {
			return key, u
		}
		if i, err := strconv.ParseInt(_value, 0, 64); err == nil {
			return key, i
		}

		return key, _value
	}

	for scanner.Scan() {
		text := scanner.Text()

		if text == "in:" || text == "set:" {
			_in = true
			_out = false
			continue
		}

		if text == "out:" || text == "get:" {
			_in = false
			_out = true
			continue
		}

		if text == "script:" {
			_in = false
			_out = false
			continue
		}

		if _in {
			k, v := splitValue(text)
			in[k] = v
			continue
		}

		if _out {
			out = append(out, text)
			continue
		}

		script = append(script, text)
	}

	return in, out, strings.Join(script, "\n"), scanner.Err()
}
