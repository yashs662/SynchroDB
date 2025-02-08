package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/yashs662/SynchroDB/internal/utils"
	"github.com/yashs662/SynchroDB/pkg/protocol"
)

type ClientSpecificCommand interface {
	Command() []string
	Execute(client *Client, args []string) ClientCommandRegistryResponse
	GetCommandInfo() protocol.CommandDescription
}

type ClientCommandRegistry struct {
	commands map[string]ClientSpecificCommand
}

type ClientCommandRegistryResponse struct {
	Response    string
	ControlFlow ControlFlow
}

func AllCommands() []ClientSpecificCommand {
	return []ClientSpecificCommand{
		&ClearCommand{},
		&ExitCommand{},
		&HelpCommand{},
	}
}

func NewCommandRegistry() *ClientCommandRegistry {
	commandRegistry := &ClientCommandRegistry{commands: make(map[string]ClientSpecificCommand)}

	for _, command := range AllCommands() {
		commandRegistry.Register(command.Command(), command)
	}

	return commandRegistry
}

// list of strings that can all be used as the SAME command
func (r *ClientCommandRegistry) Register(names []string, command ClientSpecificCommand) {
	for _, name := range names {
		r.commands[name] = command
	}
}

func (r *ClientCommandRegistry) Execute(name string, client *Client, args []string) ClientCommandRegistryResponse {
	name = strings.ToUpper(name)
	if command, exists := r.commands[name]; exists {
		return command.Execute(client, args)
	}
	return ClientCommandRegistryResponse{Response: fmt.Sprintf("Command %s not found", name), ControlFlow: NOTFOUND}
}

type ClearCommand struct{}

func (c *ClearCommand) Command() []string {
	return []string{"CLEAR", "CLS"}
}

func (c *ClearCommand) Execute(client *Client, args []string) ClientCommandRegistryResponse {
	fmt.Print("\033[H\033[2J")
	return ClientCommandRegistryResponse{Response: "Screen cleared", ControlFlow: CONTINUE}
}

func (c *ClearCommand) GetCommandInfo() protocol.CommandDescription {
	return protocol.CommandDescription{
		Command:  "",                                                             // Not required for Client Specific Commands
		Name:     fmt.Sprintf("Clear %s Clear", utils.MultipleCommandsDelimiter), // They are Same to allow for cell merging in the help table
		Syntax:   fmt.Sprintf("CLEAR %s CLS", utils.MultipleCommandsDelimiter),
		HelpText: "Clear the screen",
	}
}

type ExitCommand struct{}

func (c *ExitCommand) Command() []string {
	return []string{"EXIT", "QUIT"}
}

func (c *ExitCommand) Execute(client *Client, args []string) ClientCommandRegistryResponse {
	return ClientCommandRegistryResponse{Response: "Bye...", ControlFlow: EXIT}
}

func (c *ExitCommand) GetCommandInfo() protocol.CommandDescription {
	return protocol.CommandDescription{
		Command:  "",                                                           // Not required for Client Specific Commands
		Name:     fmt.Sprintf("Exit %s Exit", utils.MultipleCommandsDelimiter), // They are Same to allow for cell merging in the help table
		Syntax:   fmt.Sprintf("EXIT %s QUIT", utils.MultipleCommandsDelimiter),
		HelpText: "Exit the client",
	}
}

type HelpCommand struct{}

func (c *HelpCommand) Command() []string {
	return []string{"HELP", "?"}
}

func (c *HelpCommand) Execute(client *Client, args []string) ClientCommandRegistryResponse {
	clientHelpCommand := protocol.HelpCommand{}
	response, err := client.SendCommand(clientHelpCommand.GetCommandInfo().Command)

	commandDescriptions := []protocol.CommandDescription{}

	if err != nil {
		color.Red("Error trying to help message from server: %v\n", err)
	} else {
		// the response is json parsed like this:
		// return fmt.Sprintf("%v", commandDescriptions)
		// so we need to parse it back into a slice of CommandDescriptions
		// to print it in a table

		err = json.Unmarshal([]byte(response), &commandDescriptions)
		if err != nil {
			color.Red("Error reading server help message: %v\n", err)
		} else {
			// remove server help command to avoid duplication
			for i, description := range commandDescriptions {
				if description.Name == clientHelpCommand.GetCommandInfo().Name {
					commandDescriptions = append(commandDescriptions[:i], commandDescriptions[i+1:]...)
					break
				}
			}
		}
	}

	// add client specific commands to the list
	for _, command := range AllCommands() {
		description := command.GetCommandInfo()
		commandDescriptions = append(commandDescriptions, description)
	}

	// prepare CLient help table
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Command", "Syntax", "Description"})
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)

	for _, description := range commandDescriptions {
		multipleCommandChecker := fmt.Sprintf(" %s ", utils.MultipleCommandsDelimiter)

		if strings.Contains(description.Command, multipleCommandChecker) {
			splitSyntaxCommands := strings.Split(description.Syntax, multipleCommandChecker)
			splitNames := strings.Split(description.Name, multipleCommandChecker)

			if len(splitSyntaxCommands) != len(splitNames) {
				color.Red("Error: mismatch in alternative commands and names, please report this issue on github. Thanks\n")
				continue
			}

			for i, alternativeSyntax := range splitSyntaxCommands {
				table.Append([]string{
					splitNames[i],
					alternativeSyntax,
					description.HelpText,
				})
			}
		} else {
			table.Append([]string{
				description.Name,
				description.Syntax,
				description.HelpText,
			})
		}
	}

	table.Render()

	return ClientCommandRegistryResponse{Response: tableString.String(), ControlFlow: CONTINUE}
}

func (c *HelpCommand) GetCommandInfo() protocol.CommandDescription {
	return protocol.CommandDescription{
		Command:  "",                                                           // Not required for Client Specific Commands
		Name:     fmt.Sprintf("Help %s Help", utils.MultipleCommandsDelimiter), // They are Same to allow for cell merging in the help table
		Syntax:   fmt.Sprintf("HELP %s ?", utils.MultipleCommandsDelimiter),
		HelpText: "Show this help message",
	}
}
