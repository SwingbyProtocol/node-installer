package bot

import (
	"io"
	"os"
	"os/exec"

	"github.com/apenella/go-ansible/stdoutcallback"
	"github.com/apenella/go-ansible/stdoutcallback/results"
)

type BotExecute struct {
	Write       io.Writer
	ResultsFunc stdoutcallback.StdoutCallbackResultsFunc
}

// Execute takes a command and args and runs it, streaming output to stdout
func (e *BotExecute) Execute(command string, args []string, prefix string) error {
	if e.Write == nil {
		e.Write = os.Stdout
	}
	execDoneChan := make(chan int8)
	defer close(execDoneChan)
	execErrChan := make(chan error)
	defer close(execErrChan)

	cmd := exec.Command(command, args...)
	cmdReader, err := cmd.StdoutPipe()
	defer cmdReader.Close()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		if e.ResultsFunc == nil {
			e.ResultsFunc = results.DefaultStdoutCallbackResults
		}
		err := e.ResultsFunc(prefix, cmdReader, e.Write)
		if err != nil {
			execErrChan <- err
			return
		}
		execDoneChan <- int8(0)
	}()

	select {
	case <-execDoneChan:
	case err := <-execErrChan:
		return err
	}
	err = cmd.Wait()
	return err
}
