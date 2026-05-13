package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"

	"github.com/stefanosbou/repokit/internal/runner"
)

type resultMsg struct{ r runner.Result }
type doneMsg struct{}

type repoState struct {
	name    string
	path    string
	done    bool
	status  runner.Status
	message string
	err     error
}

type model struct {
	repos     []repoState
	byName    map[string]int
	spinner   spinner.Model
	ch        <-chan runner.Result
	total     int
	done      int
	nameWidth int
}

func initialModel(tasks []runner.Task, ch <-chan runner.Result) model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	repos := make([]repoState, len(tasks))
	byName := make(map[string]int, len(tasks))
	nameWidth := 0
	for i, t := range tasks {
		repos[i] = repoState{name: t.RepoName, path: t.RepoPath}
		byName[t.RepoName] = i
		if len(t.RepoName) > nameWidth {
			nameWidth = len(t.RepoName)
		}
	}
	return model{
		repos:     repos,
		byName:    byName,
		spinner:   s,
		ch:        ch,
		total:     len(tasks),
		nameWidth: nameWidth,
	}
}

func waitForResult(ch <-chan runner.Result) tea.Cmd {
	return func() tea.Msg {
		r, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return resultMsg{r}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForResult(m.ch))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case resultMsg:
		if idx, ok := m.byName[msg.r.RepoName]; ok {
			m.repos[idx] = repoState{
				name:    msg.r.RepoName,
				path:    msg.r.RepoPath,
				done:    true,
				status:  msg.r.Status,
				message: msg.r.Message,
				err:     msg.r.Err,
			}
			m.done++
		}
		if m.done == m.total {
			return m, tea.Quit
		}
		return m, tea.Batch(m.spinner.Tick, waitForResult(m.ch))

	case doneMsg:
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder
	for _, r := range m.repos {
		if !r.done {
			fmt.Fprintf(&sb, "  %s  %-*s  ···\n", m.spinner.View(), m.nameWidth, r.name)
			continue
		}
		icon := rowIcon(r)
		msg := r.message
		if r.err != nil {
			msg = r.err.Error()
		}
		fmt.Fprintf(&sb, "  %s  %-*s  %s\n", icon, m.nameWidth, r.name, msg)
	}
	return sb.String()
}

func rowIcon(r repoState) string {
	switch {
	case r.err != nil, r.status == runner.StatusError:
		return color.RedString("✗")
	case r.status == runner.StatusSkipped:
		return color.YellowString("⚠")
	default:
		return color.GreenString("✓")
	}
}

// RunWithProgress runs tasks and renders live per-repo status.
// Falls back to calling printFallback per result when stdout is not a TTY.
// Returns all results after completion so callers can compute summaries.
func RunWithProgress(
	ctx context.Context,
	title string,
	tasks []runner.Task,
	parallel int,
	printFallback func(runner.Result),
) ([]runner.Result, error) {
	fmt.Println(title)
	fmt.Println()

	ch := runner.RunAll(ctx, tasks, parallel)
	results := make([]runner.Result, 0, len(tasks))

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		for r := range ch {
			printFallback(r)
			results = append(results, r)
		}
		return results, nil
	}

	m := initialModel(tasks, ch)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm, ok := finalModel.(model)
	if !ok {
		return nil, fmt.Errorf("internal error: unexpected model type %T", finalModel)
	}
	for _, r := range fm.repos {
		results = append(results, runner.Result{
			RepoName: r.name,
			RepoPath: r.path,
			Status:   r.status,
			Message:  r.message,
			Err:      r.err,
		})
	}
	return results, nil
}
