package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "runtime"
    "strings"
    "sync"
)

var fileName = flag.String("f", "", "filename containing a newline-separated list of arguments")
var maxGr = flag.Int("n", 1, "max number of concurrent goroutines")
var prefix = flag.String("p", "", "prefix for every argument")
var suffix = flag.String("s", "", "suffix for every argument")

var usage = func() {
    fmt.Fprintf(os.Stderr, "%s runs multiple arguments concurrently for the same command\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "Usage of %s: %s [-n numConcurrent] [-f argFile] [-p prefix] [-s suffix] command [arg1 arg2 arg3...]\n", os.Args[0], os.Args[0])
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    flag.Parse()
    if len(flag.Args()) == 0 {
        fmt.Fprint(os.Stderr, "Not enough arguments\n")
        usage()
        os.Exit(1)
    }

    fullCmd := flag.Args()[0]
    path, err := exec.LookPath(strings.Split(fullCmd, " ")[0])
    if nil != err {
        fmt.Fprint(os.Stderr, err)
        os.Exit(1)
    }

    var allArgs []string
    if len(*fileName) != 0 {
        content, err := ioutil.ReadFile(*fileName)
        if nil != err {
            fmt.Fprint(os.Stderr, err)
            os.Exit(1)
        } else {
            allArgs = append(allArgs, strings.Split(strings.Trim(string(content), "\n"), "\n")...)
        }
    }
    allArgs = append(allArgs, flag.Args()[1:]...)

    runtime.GOMAXPROCS(runtime.NumCPU())
    sem := make(chan int, *maxGr)
    for i := 0; i < *maxGr; i++ {
        sem <- 1
    }
    var wg sync.WaitGroup
    for i := 0; i < len(allArgs); i++ {
        args := strings.Split(allArgs[i], " ")
        if len(*prefix) != 0 {
            args[0] = *prefix + args[0]
        }
        if len(*suffix) != 0 {
            args[len(args) - 1] = args[len(args) - 1] + *suffix
        }
        args = append([]string{fullCmd}, args...)

        cmd := exec.Cmd{Path: path, Args: args}
        <- sem
	wg.Add(1)
        go func(cmd exec.Cmd, i int) {
            out, err := cmd.Output()
            if nil != err {
                fmt.Fprintf(os.Stderr, "%s", err)
            }
            if nil != out {
                fmt.Fprintf(os.Stdout, "%s", out)
            }
	    wg.Done()
            sem <- 1
        }(cmd, i)
    }

    wg.Wait()
}
