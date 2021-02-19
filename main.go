package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// TODO
// args
// 	* -o to reopen /dev/tty
//  * -t to echo
//  * -p to prompt before each run
// features
//   * track failures and re-run
//   *  use '--' as a separator for arguments

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		args := []string{}
		placeholderFound := false

		for _, c := range os.Args[2:len(os.Args)] {
			if c == "{}" {
				args = append(args, scanner.Text())
				placeholderFound = true
			} else {
				args = append(args, c)
			}
		}

		if !placeholderFound {
			args = append(args, scanner.Text())
		}

		cmd := exec.CommandContext(context.TODO(), os.Args[1], args...)

		output, err := cmd.CombinedOutput()
		fmt.Print(string(output))
		if e, ok := err.(*exec.ExitError); ok {
			panic(errors.Wrap(e, "exit error"))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
