package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	version = "1.0.0"
)

func Starter(c chan *exec.Cmd) {
	for {
		select {
		case cmd := <- c:
			c := NewCmd("go", "build", "-o", "build", "test.go")
			time.Sleep(3* time.Second)
			err := c.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			fmt.Printf("Running process: %s", cmd.String())
			err = cmd.Run()
			fmt.Printf("Process exited with: %v\n", err)
		}
	}
}

func Stop(c chan *exec.Cmd) {
	for {
		select {
		case cmd := <- c:
			if !cmd.ProcessState.Exited() {
				if err := cmd.Process.Kill(); err != nil {
					fmt.Printf("Error in killing: %s\n", err)
					os.Exit(1)
				}
			}
			fmt.Printf("Killed process %d\n",  cmd.Process.Pid)
		}
	}
}

func main() {

	fmt.Printf("Running go-watcher %s\n", version)
	startChan := make(chan *exec.Cmd)
	stopChan := make(chan *exec.Cmd)

	go Starter(startChan)
	go Stop(stopChan)
	go watch(startChan, stopChan)

	select {}
}

func NewCmd(c string, args ...string) *exec.Cmd {
	return exec.Command(c, args...)
}

func watch(start chan *exec.Cmd, stop chan *exec.Cmd) {
	cmd := NewCmd("./build")
	fileMap := make(map[string]int64)
	fileName := "/Users/RNA/MyProjects/go-watcher/test.go"
	if len(os.Args) > 2 {
		fileName = os.Args[2]
	}

	start <- cmd
	for {
		res, err := os.Stat(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		lastModTime, ok := fileMap[res.Name()]
		if !ok {
			fileMap[res.Name()] = res.ModTime().Unix()
		}

		time.Sleep(3* time.Second)
		if ok && lastModTime != res.ModTime().Unix() {
			fileMap[res.Name()] = res.ModTime().Unix()
			stop <- cmd
			fmt.Println("Restarting")
			cmd = NewCmd("./build")
			start <- cmd
		}
	}
}
