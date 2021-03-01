package main

import (
	"bufio"
	"context"
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
	cmd    string
	args   []string
	output []byte
	err    error
}

func work(jobs chan (job), results chan (result), wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		cmd := exec.CommandContext(context.TODO(), j.cmd, j.args...)
		output, err := cmd.CombinedOutput()
		results <- result{j.cmd, j.args, output, err}
	}
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
			fmt.Print(string(result.output))
			if e, ok := result.err.(*exec.ExitError); ok {
				log.Println(errors.Wrapf(e, "exit error on %s %s", result.cmd, result.args))
			}
		}
		resultsWg.Done()
	}()

	for scanner.Scan() {
		args := []string{}
		placeholderFound := false

		for _, c := range os.Args[2:len(os.Args)] {
			if strings.Contains(c, "{}") {
				args = append(args, strings.Replace(c, "{}", scanner.Text(), -1))
				placeholderFound = true
			} else {
				args = append(args, c)
			}
		}

		if !placeholderFound {
			args = append(args, scanner.Text())
		}

		jobs <- job{os.Args[1], args}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	close(jobs)
	resultsWg.Wait()
}
