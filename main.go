package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"

	_ "net/http"
	_ "net/http/pprof"
)

// TODO
// args
// 	* -o to reopen /dev/tty
//  * -t to echo
//  * -p to prompt before each run
// features
//   * track failures and re-run
//   *  use '--' as a separator for arguments

type job struct {
	cmd  string
	args []string
}

type result struct {
	cmd      string
	args     []string
	exitcode int
	output   []byte
	err      error
}

func work(jobs chan (job), results chan (result), wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		cmd := exec.CommandContext(context.TODO(), j.cmd, j.args...)
		output, err := cmd.CombinedOutput()
		results <- result{j.cmd, j.args, cmd.ProcessState.ExitCode(), output, err}
	}
}

var errorOnly bool
var echo bool

func init() {
	flag.BoolVar(&errorOnly, "error-only", false, "")
	flag.BoolVar(&echo, "echo", false, "")
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	const workers = 5
	jobs := make(chan job, workers)
	results := make(chan result, workers)
	var wg sync.WaitGroup
	var resultsWg sync.WaitGroup

	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go work(jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	resultsWg.Add(1)
	go func() {
		for result := range results {
			if !errorOnly || result.exitcode != 0 || result.err != nil {
				fmt.Print(string(result.output))
				if e, ok := result.err.(*exec.ExitError); ok {
					log.Println(errors.Wrapf(e, "exit error on %s %s", result.cmd, result.args))
				}
			}
		}
		resultsWg.Done()
	}()

	for scanner.Scan() {
		t := scanner.Text()
		t = strings.TrimSpace(t)
		args := []string{}
		placeholderFound := false

		for _, c := range flag.Args()[1:len(flag.Args())] {
			if strings.Contains(c, "{}") {
				args = append(args, strings.Replace(c, "{}", t, -1))
				placeholderFound = true
			} else {
				args = append(args, c)
			}
		}

		if !placeholderFound {
			args = append(args, t)
		}

		j := job{flag.Args()[0], args}
		if echo {
			fmt.Printf("enqueuing job %#v\n", j)
		}
		jobs <- j
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	close(jobs)
	resultsWg.Wait()
}
