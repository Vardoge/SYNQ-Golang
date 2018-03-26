package common

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type ExecObj struct {
	Cmd     *exec.Cmd
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	stdin   io.WriteCloser
	ReadOut []byte
	ReadErr []byte
	RunErr  error
	Err     SynqError
}

func NewExec(command string, args ...string) ExecObj {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	obj := ExecObj{Cmd: cmd}
	// create a default SynqError
	obj.Err = SynqError{
		Name:    "exec_error",
		Url:     "http://docs.synq.fm/api/v1/errors/",
		Message: "An error occured while running your script",
	}
	return obj
}

func (e *ExecObj) Open() error {
	stdin, err := e.Cmd.StdinPipe()
	e.stdin = stdin
	if err != nil {
		return err
	}
	stdout, err := e.Cmd.StdoutPipe()
	e.stdout = stdout
	if err != nil {
		return err
	}

	stderr, err := e.Cmd.StderrPipe()
	e.stderr = stderr
	if err != nil {
		return err
	}
	return nil
}

func (e *ExecObj) Close() {
	e.stderr.Close()
	e.stdout.Close()
}

func (e *ExecObj) Read() error {
	o, err := ioutil.ReadAll(e.stdout)
	if err != nil {
		return err
	}
	e.ReadOut = o
	o2, err := ioutil.ReadAll(e.stderr)
	if err != nil {
		return err
	}
	e.ReadErr = o2
	return nil
}

func (e *ExecObj) Exec(fn func(io.WriteCloser)) {
	if err := e.Open(); err != nil {
		e.RunErr = err
		return
	}
	if err := e.Cmd.Start(); err != nil {
		e.RunErr = err
		return
	}

	fn(e.stdin)

	if err := e.stdin.Close(); err != nil {
		e.RunErr = err
		return
	}

	if err := e.Read(); err != nil {
		e.RunErr = err
		return
	}

	if err := e.Cmd.Wait(); err != nil {
		e.RunErr = err
		return
	}
}

func (e *ExecObj) ErrorMsg() string {
	if e.RunErr == nil {
		return ""
	}
	return e.RunErr.Error()
}

func (e *ExecObj) StatusCode() int {
	if e.ErrorMsg() == "" {
		return 200
	}
	return 400
}

func (e *ExecObj) StatusBody() []byte {
	if e.ErrorMsg() == "" {
		return e.ReadOut
	}
	return e.MarshalError()
}

func (e *ExecObj) MarshalError() (body []byte) {
	// TODO(mastensg): Don't do string matching, but rather something with this:
	// TODO(mastensg): https://golang.org/pkg/syscall/#WaitStatus.ExitStatus
	if e.ErrorMsg() != "exit status 1" {
		log.Println("stdout:", string(e.ReadOut))
		log.Println("stderr:", string(e.ReadErr))
		e.Err.Message = e.ErrorMsg()
	} else {
		jsErr := json.RawMessage(e.ReadErr)
		e.Err.Details = &jsErr
	}
	body, err := json.MarshalIndent(e.Err, "", "    ")
	if err != nil {
		log.Println("error marshaling data ", err.Error())
		return body
	}
	return body
}
