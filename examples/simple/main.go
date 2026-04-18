// Package main demonstrates multi-level (kubectl-style) tab completion with
// cursor-following popup position.
//
// Level 1 – verb:          get / describe / delete / apply / logs
// Level 2 – resource type: pod / deployment / service / configmap / node
// Level 3 – resource name: live-style names per type
// Level 4 – flags:         --namespace / --output / --selector / ...
package main

import (
	"fmt"
	"log"
	"strings"

	prompt "github.com/rynvaro/bubble-prompt"
)

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.WithPrefix("kubectl> "),
		prompt.WithPlaceholder("type a command, Tab for multi-level completion, Ctrl+C to quit"),
		// popup follows the cursor instead of staying fixed at the prefix
		prompt.WithCompletionPosition(prompt.CompletionAtCursor),
	)
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func executor(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	args := strings.Fields(input)
	switch args[0] {
	case "exit", "quit":
		fmt.Println("Goodbye!")
		return prompt.ErrExit
	default:
		fmt.Printf("▶ kubectl %s\n", input)
	}
	return nil
}

// ---- suggestion tables ------------------------------------------------------

var verbs = []prompt.Suggestion{
	{Text: "get", Description: "display one or more resources"},
	{Text: "describe", Description: "show detailed info of a resource"},
	{Text: "delete", Description: "delete a resource"},
	{Text: "apply", Description: "apply configuration from a file"},
	{Text: "logs", Description: "print logs of a container in a pod"},
	{Text: "exit", Description: "exit the prompt"},
}

var resources = []prompt.Suggestion{
	{Text: "pod", Display: "pod", Description: "running container group"},
	{Text: "deployment", Display: "deployment", Description: "declarative app deployment"},
	{Text: "service", Display: "service", Description: "network service exposure"},
	{Text: "configmap", Display: "configmap", Description: "configuration data"},
	{Text: "node", Display: "node", Description: "cluster node"},
	{Text: "namespace", Display: "namespace", Description: "resource isolation scope"},
}

// fakeNames pretends to be the list of live resources for each type.
var fakeNames = map[string][]prompt.Suggestion{
	"pod": {
		{Text: "nginx-7d8b9c-xk2pq", Description: "Running"},
		{Text: "redis-5f6d8b-mnjlp", Description: "Running"},
		{Text: "worker-abc12", Description: "Pending"},
	},
	"deployment": {
		{Text: "nginx", Description: "3/3 ready"},
		{Text: "redis", Description: "1/1 ready"},
	},
	"service": {
		{Text: "kubernetes", Description: "ClusterIP"},
		{Text: "nginx-svc", Description: "LoadBalancer"},
	},
	"configmap": {
		{Text: "kube-dns", Description: "system"},
		{Text: "app-config", Description: "default"},
	},
	"node": {
		{Text: "node-1", Description: "Ready"},
		{Text: "node-2", Description: "Ready"},
	},
	"namespace": {
		{Text: "default", Description: ""},
		{Text: "kube-system", Description: ""},
		{Text: "production", Description: ""},
	},
}

// flags available after a resource name for each verb.
var verbFlags = map[string][]prompt.Suggestion{
	"get": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  specify namespace"},
		{Text: "--output", Display: "--output", Description: "-o  output format (json|yaml|wide)"},
		{Text: "--selector", Display: "--selector", Description: "-l  label selector"},
		{Text: "--all-namespaces", Display: "--all-namespaces", Description: "-A  all namespaces"},
		{Text: "--watch", Display: "--watch", Description: "-w  watch for changes"},
		{Text: "--show-labels", Display: "--show-labels", Description: "    show all labels"},
	},
	"describe": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  specify namespace"},
		{Text: "--selector", Display: "--selector", Description: "-l  label selector"},
	},
	"delete": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  specify namespace"},
		{Text: "--force", Display: "--force", Description: "    force delete"},
		{Text: "--grace-period", Display: "--grace-period", Description: "    graceful exit wait seconds"},
		{Text: "--selector", Display: "--selector", Description: "-l  label selector"},
	},
	"logs": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  specify namespace"},
		{Text: "--follow", Display: "--follow", Description: "-f  stream logs"},
		{Text: "--tail", Display: "--tail", Description: "    last N lines"},
		{Text: "--previous", Display: "--previous", Description: "-p  previous container instance"},
		{Text: "--container", Display: "--container", Description: "-c  specify container name"},
	},
	"apply": {
		{Text: "--filename", Display: "--filename", Description: "-f  file path or URL"},
		{Text: "--namespace", Display: "--namespace", Description: "-n  specify namespace"},
		{Text: "--dry-run", Display: "--dry-run", Description: "    dry run, no actual changes"},
		{Text: "--recursive", Display: "--recursive", Description: "-R  process directories recursively"},
	},
}

// ---- completer --------------------------------------------------------------

func completer(d prompt.Document) []prompt.Suggestion {
	before := d.TextBeforeCursor()
	args := strings.Fields(before)
	word := d.CurrentWord()
	atWordBoundary := strings.HasSuffix(before, " ")

	// Count how many non-flag args we've seen so far.
	// Flags (--xxx) can appear anywhere after level-3 and are always offered
	// when the current word starts with "-".
	nonFlagArgs := make([]string, 0, len(args))
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			nonFlagArgs = append(nonFlagArgs, a)
		}
	}

	// If the user is typing a flag, offer flags for the current verb.
	if strings.HasPrefix(word, "-") && len(nonFlagArgs) >= 1 {
		if flags, ok := verbFlags[nonFlagArgs[0]]; ok {
			return prompt.FilterHasPrefix(flags, word, true)
		}
		return nil
	}
	// If we're at a word boundary and there are already ≥3 non-flag args,
	// offer flags as a next-step hint.
	if atWordBoundary && len(nonFlagArgs) >= 3 {
		if flags, ok := verbFlags[nonFlagArgs[0]]; ok {
			return flags
		}
		return nil
	}

	switch {
	case len(args) == 0:
		return prompt.FilterHasPrefix(verbs, word, true)

	case len(args) == 1 && !atWordBoundary:
		return prompt.FilterHasPrefix(verbs, args[0], true)

	case len(args) == 1 && atWordBoundary:
		if verbTakesResource(args[0]) {
			return resources
		}
		return nil

	case len(args) == 2 && !atWordBoundary:
		if verbTakesResource(args[0]) {
			return prompt.FilterHasPrefix(resources, args[1], true)
		}
		return nil

	case len(args) == 2 && atWordBoundary:
		if names, ok := fakeNames[args[1]]; ok {
			return names
		}
		return nil

	case len(args) == 3 && !atWordBoundary:
		if names, ok := fakeNames[args[1]]; ok {
			return prompt.FilterHasPrefix(names, args[2], true)
		}
		return nil
	}

	return nil
}

func verbTakesResource(verb string) bool {
	switch verb {
	case "get", "describe", "delete", "logs":
		return true
	}
	return false
}
