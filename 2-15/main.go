package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

var running []*exec.Cmd

func main() {
	reader := bufio.NewReader(os.Stdin)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for range sig {
			for _, c := range running {
				if c != nil && c.Process != nil {
					_ = c.Process.Kill()
				}
			}
			fmt.Print("\n> ")
		}
	}()

	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println()
			return
		}
		if err != nil {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		execLine(line)
	}
}

//CORE

func execLine(line string) {
	// 1. AND / OR
	tokens := splitByLogical(line)

	executeLogical(tokens)
}

//LOGIC (&& ||)

type token struct {
	op   string // "", "&&", "||"
	line string
}

func splitByLogical(line string) []token {
	var res []token
	cur := ""

	i := 0
	for i < len(line) {
		if i+1 < len(line) && line[i] == '&' && line[i+1] == '&' {
			res = append(res, token{op: "&&", line: strings.TrimSpace(cur)})
			cur = ""
			i += 2
			continue
		}
		if i+1 < len(line) && line[i] == '|' && line[i+1] == '|' {
			res = append(res, token{op: "||", line: strings.TrimSpace(cur)})
			cur = ""
			i += 2
			continue
		}
		cur += string(line[i])
		i++
	}

	res = append(res, token{op: "", line: strings.TrimSpace(cur)})
	return res
}

func executeLogical(tokens []token) {
	var lastOk = true

	for i, t := range tokens {
		if t.line == "" {
			continue
		}

		if i > 0 && tokens[i-1].op == "&&" && !lastOk {
			continue
		}

		if i > 0 && tokens[i-1].op == "||" && lastOk {
			continue
		}

		lastOk = runSegment(t.line)
	}
}

//PIPELINE

func runSegment(line string) bool {
	parts := splitPipe(line)

	if len(parts) > 1 {
		return runPipeline(parts) == nil
	}

	return runCommand(parse(parts[0])) == nil
}

func splitPipe(line string) []string {
	var res []string
	cur := ""

	for i := 0; i < len(line); i++ {
		if line[i] == '|' {
			res = append(res, strings.TrimSpace(cur))
			cur = ""
		} else {
			cur += string(line[i])
		}
	}

	res = append(res, strings.TrimSpace(cur))
	return res
}

func runPipeline(parts []string) error {
	var cmds []*exec.Cmd

	for _, p := range parts {
		args := parse(p)
		if len(args) == 0 {
			return nil
		}

		args, _, _ = splitRedir(args)

		if len(args) == 0 {
			return nil
		}

		cmds = append(cmds, exec.Command(args[0], args[1:]...))
	}

	if len(cmds) == 0 {
		return nil
	}

	running = cmds

	for i := 0; i < len(cmds)-1; i++ {
		out, err := cmds[i].StdoutPipe()
		if err != nil {
			return err
		}
		cmds[i+1].Stdin = out
	}

	cmds[0].Stdin = os.Stdin
	cmds[len(cmds)-1].Stdout = os.Stdout
	cmds[len(cmds)-1].Stderr = os.Stderr

	for _, c := range cmds {
		if err := c.Start(); err != nil {
			return err
		}
	}

	for _, c := range cmds {
		_ = c.Wait()
	}

	running = nil
	return nil
}

//EXEC

func runCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	args, inFile, outFile := splitRedir(args)

	if len(args) == 0 {
		return nil
	}

	cmd := exec.Command(args[0], args[1:]...)

	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdin = f
	} else {
		cmd.Stdin = os.Stdin
	}

	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdout = f
	} else {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = os.Stderr

	running = []*exec.Cmd{cmd}
	err := cmd.Run()
	running = nil

	return err
}

//PARSER

func parse(line string) []string {
	parts := strings.Fields(line)

	for i := range parts {
		if strings.HasPrefix(parts[i], "$") {
			parts[i] = os.Getenv(strings.TrimPrefix(parts[i], "$"))
		}
	}

	return parts
}

//REDIRECTION

func splitRedir(args []string) ([]string, string, string) {
	var out []string
	var inFile, outFile string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case ">":
			if i+1 < len(args) {
				outFile = args[i+1]
				i++
			}
		case "<":
			if i+1 < len(args) {
				inFile = args[i+1]
				i++
			}
		default:
			out = append(out, args[i])
		}
	}

	return out, inFile, outFile
}

//
