package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// error types
type FileErr struct{ Err error }
type TezErr struct{ Err error }

// fetching tezaurs -----
type TezaursResponse struct {
	Code int
	Body io.ReadCloser
}

func FetchTezaursCmd(url string) tea.Cmd {
	return func() tea.Msg {
		res, err := FetchTezaurs(url)

		if err != nil {
			return TezErr{Err: err}
		}

		return TezaursResponse{Code: res.StatusCode, Body: res.Body}
	}
}

func FetchTezaurs(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	return client.Get(url)
}

// fzf -----
type FzfFinished struct {
	ExitCode int
	Err      error
	Word     string
}

const fzfOpts = "--height=100% --prompt='[T]: '"

func OpenFzfCmd() tea.Cmd {
	// temp files aren't ideal, but i don't see a better solution
	file, err := os.CreateTemp(os.TempDir(), "")

	if err != nil {
		return func() tea.Msg {
			return FileErr{err}
		}
	}

	c := exec.Command(
		"bash", "-c",
		fmt.Sprintf("fzf %s < wordlist.txt > %s", fzfOpts, file.Name()),
	)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		bytes, fileError := os.ReadFile(file.Name())

		if fileError != nil {
			return FileErr{Err: fileError}
		}

		defer os.Remove(file.Name())

		return FzfFinished{
			Err:      err,
			Word:     strings.TrimSpace(string(bytes)),
			ExitCode: c.ProcessState.ExitCode(),
		}
	})
}
