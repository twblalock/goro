package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "runtime"
    "strings"
)

var cmd = flag.String("c", "", "command to run")
var fileName = flag.String("f", "", "filename containing a newline-separated list of arguments")
var maxGr = flag.Int("n", 1, "max number of goroutines")
var prefix = flag.String("p", "", "prefix for every argument")
var suffix = flag.String("s", "", "suffix for every argument")

var usage = func() {
    fmt.Fprintf(os.Stderr, "%s runs multiple arguments concurrently for the same command\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "Usage of %s: %s -c command arg1 arg2 arg3...\n", os.Args[0], os.Args[0])
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    flag.Parse()

    if len(*cmd) == 0 {
        usage()
	os.Exit(1)
    }

    path, err := exec.LookPath(strings.Split(*cmd, " ")[0])
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
            allArgs = strings.Split(strings.Trim(string(content), "\n"), "\n")
        }
    } else {
        allArgs = flag.Args()
    }

    runtime.GOMAXPROCS(runtime.NumCPU())
    sem := make(chan int, *maxGr)
    quit := make(chan int) // Used when all goroutines are finsihed
    for i := 0; i < *maxGr; i++ {
        sem <- 1
    }

    for i := 0; i < len(allArgs); i++ {
        args := strings.Split(allArgs[i], " ")
        if len(*prefix) != 0 {
	    args[0] = *prefix + args[0]
        }
        if len(*suffix) != 0 {
            args[len(args) - 1] = args[len(args) - 1] + *suffix
        }
	args = append([]string{ *cmd }, args...)

        cmd := exec.Cmd{Path: path, Args: args}
	<- sem
        go func(cmd exec.Cmd, i int) {
	    out, err := cmd.Output()
	    if nil != err {
                fmt.Fprintf(os.Stderr, "%s", err)
            }
            if nil != out {
                fmt.Fprintf(os.Stdout, "%s", out)
            }
            sem <- 1
	    if i == len(allArgs) - 1 {
                quit <- 1
            }
        }(cmd, i)
    }

    <- quit
}
