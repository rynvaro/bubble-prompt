// Package main demonstrates multi-level (kubectl-style) tab completion.
//
// Level 1 – verb:          get / describe / delete / apply / logs
// Level 2 – resource type: pod / deployment / service / configmap / node
// Level 3 – resource name: dynamically generated from the chosen type
package main

import (
	"fmt"
	"log"
	"strings"

	prompt "bubble-prompt"
)

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.WithPrefix("kubectl> "),
		prompt.WithPlaceholder("输入命令，Tab 多级补全，Ctrl+C 退出"),
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
	default:
		fmt.Printf("▶ kubectl %s\n", input)
	}
	return nil
}

// ---- suggestion tables ------------------------------------------------------

var verbs = []prompt.Suggestion{
	{Text: "get", Description: "显示一个或多个资源"},
	{Text: "describe", Description: "显示资源的详细信息"},
	{Text: "delete", Description: "删除资源"},
	{Text: "apply", Description: "通过文件应用配置"},
	{Text: "logs", Description: "打印 Pod 中容器的日志"},
	{Text: "exit", Description: "退出"},
}

var resources = []prompt.Suggestion{
	{Text: "pod", Display: "pod", Description: "运行中的容器组"},
	{Text: "deployment", Display: "deployment", Description: "声明式应用部署"},
	{Text: "service", Display: "service", Description: "网络服务暴露"},
	{Text: "configmap", Display: "configmap", Description: "配置数据"},
	{Text: "node", Display: "node", Description: "集群节点"},
	{Text: "namespace", Display: "namespace", Description: "资源隔离空间"},
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

// ---- completer --------------------------------------------------------------

func completer(d prompt.Document) []prompt.Suggestion {
	// args = words already confirmed (before the cursor's current word)
	before := d.TextBeforeCursor()
	args := strings.Fields(before)
	word := d.CurrentWord()

	// If the cursor is right after a space (current word is empty) the user
	// finished typing the previous word and is starting a new one.
	atWordBoundary := strings.HasSuffix(before, " ")

	switch {
	case len(args) == 0:
		// Nothing typed yet → offer verbs
		return prompt.FilterHasPrefix(verbs, word, true)

	case len(args) == 1 && !atWordBoundary:
		// Typing the first word → filter verbs
		return prompt.FilterHasPrefix(verbs, args[0], true)

	case len(args) == 1 && atWordBoundary:
		// First word done, starting second → offer resource types
		// (only for verbs that take a resource)
		if verbTakesResource(args[0]) {
			return resources
		}
		return nil

	case len(args) == 2 && !atWordBoundary:
		// Typing the resource type → filter resources
		if verbTakesResource(args[0]) {
			return prompt.FilterHasPrefix(resources, args[1], true)
		}
		return nil

	case len(args) == 2 && atWordBoundary:
		// Resource type done, starting third word → offer resource names
		if names, ok := fakeNames[args[1]]; ok {
			return names
		}
		return nil

	case len(args) == 3 && !atWordBoundary:
		// Typing a resource name → filter names
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
