package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/busyrockin/api-vault/core"
	"github.com/busyrockin/api-vault/ui"
)

type credential struct {
	name    string
	apiType string
	created time.Time
}

type interactiveModel struct {
	db          *core.Database
	credentials []credential
	cursor      int
	filter      string
	viewing     bool
	viewContent string
	adding      bool
	setup       setupModel
	status      string
	err         error
}

func newInteractiveModel(db *core.Database) (interactiveModel, error) {
	m := interactiveModel{
		db: db,
	}

	if err := m.loadCredentials(); err != nil {
		return m, err
	}

	return m, nil
}

func (m *interactiveModel) loadCredentials() error {
	creds, err := m.db.ListCredentials()
	if err != nil {
		return err
	}

	m.credentials = make([]credential, len(creds))
	for i, c := range creds {
		m.credentials[i] = credential{
			name:    c.Name,
			apiType: c.APIType,
			created: c.CreatedAt,
		}
	}

	return nil
}

func (m *interactiveModel) filteredCredentials() []credential {
	if m.filter == "" {
		return m.credentials
	}

	var filtered []credential
	lower := strings.ToLower(m.filter)
	for _, c := range m.credentials {
		if strings.Contains(strings.ToLower(c.name), lower) ||
			strings.Contains(strings.ToLower(c.apiType), lower) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (m interactiveModel) Init() tea.Cmd {
	return nil
}

func (m interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.adding {
		return m.updateAdding(msg)
	}
	if m.viewing {
		return m.updateViewing(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear status on any keypress
		m.status = ""

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			filtered := m.filteredCredentials()
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}

		case "enter":
			filtered := m.filteredCredentials()
			if len(filtered) > 0 {
				cred := filtered[m.cursor]
				key, err := m.db.GetCredential(cred.name)
				if err != nil {
					m.err = err
					return m, nil
				}

				if err := clipboard.WriteAll(key); err != nil {
					m.err = fmt.Errorf("failed to copy to clipboard: %w", err)
				}

				m.viewing = true
				m.viewContent = key
			}

		case "a":
			m.adding = true
			m.setup = newSetupModel(m.db)
			return m, nil

		case "d":
			filtered := m.filteredCredentials()
			if len(filtered) > 0 {
				cred := filtered[m.cursor]
				if err := m.db.DeleteCredential(cred.name); err != nil {
					m.err = err
					return m, nil
				}
				if err := m.loadCredentials(); err != nil {
					m.err = err
					return m, nil
				}
				if m.cursor >= len(m.credentials) && m.cursor > 0 {
					m.cursor--
				}
			}

		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
			}

		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m interactiveModel) updateAdding(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.adding = false
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	updated, _ := m.setup.Update(msg)
	m.setup = updated.(setupModel)

	if m.setup.done {
		m.adding = false
		m.status = "✓ Credential saved"
		if err := m.loadCredentials(); err != nil {
			m.err = err
		}
	}

	return m, nil
}

func (m interactiveModel) updateViewing(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			m.viewing = false
			m.viewContent = ""
			return m, nil
		}
	}
	return m, nil
}

func (m interactiveModel) View() string {
	if m.adding {
		return m.setup.View()
	}
	if m.viewing {
		return m.renderViewing()
	}

	var b strings.Builder

	// Title + tagline
	b.WriteString(ui.TitleStyle.Render("🔐 Agent Vault"))
	b.WriteString("\n")
	b.WriteString(ui.Muted.Render("Keys for your agents"))
	b.WriteString("\n\n")

	// Status message
	if m.status != "" {
		b.WriteString(ui.Success.Render(m.status))
		b.WriteString("\n\n")
	}

	// Error display
	if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Filter display
	if m.filter != "" {
		b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Filter: %s", m.filter)))
		b.WriteString("\n\n")
	}

	// Credentials list
	filtered := m.filteredCredentials()
	if len(filtered) == 0 {
		b.WriteString(ui.NormalStyle.Render("No credentials found"))
	} else {
		for i, cred := range filtered {
			status := m.getStatus(cred.created)
			statusStr := m.formatStatus(status)

			line := fmt.Sprintf("%s  %s  %s", statusStr, cred.name, ui.Muted.Render(cred.apiType))

			if i == m.cursor {
				b.WriteString(ui.SelectedStyle.Render("❯ " + line))
			} else {
				b.WriteString(ui.NormalStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("[↑↓/jk] Navigate  [Enter] Copy  [a] Add  [d] Delete  [Type] Filter  [q] Quit"))

	return ui.BoxStyle.Render(b.String())
}

func (m interactiveModel) renderViewing() string {
	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("🔐 Credential Copied"))
	b.WriteString("\n\n")
	b.WriteString(ui.Success.Render("✓ Copied to clipboard"))
	b.WriteString("\n\n")
	b.WriteString(ui.SubtitleStyle.Render("Preview:"))
	b.WriteString("\n")

	// Show first and last few chars
	preview := m.viewContent
	if len(preview) > 40 {
		preview = preview[:15] + "..." + preview[len(preview)-15:]
	}
	b.WriteString(ui.NormalStyle.Render(preview))

	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[Enter/Esc] Back"))

	return ui.BoxStyle.Render(b.String())
}

func (m interactiveModel) getStatus(created time.Time) string {
	age := time.Since(created)

	if age < 7*24*time.Hour {
		return "recent"
	} else if age < 30*24*time.Hour {
		return "ok"
	} else if age < 90*24*time.Hour {
		return "warning"
	}
	return "old"
}

func (m interactiveModel) formatStatus(status string) string {
	switch status {
	case "recent":
		return ui.StatusRecentStyle.Render("[✓]")
	case "ok":
		return ui.Success.Render("[✓]")
	case "warning":
		return ui.StatusWarningStyle.Render("[⚠]")
	case "old":
		return ui.StatusErrorStyle.Render("[✗]")
	default:
		return "[?]"
	}
}

func runInteractive() error {
	db, err := openVault()
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := newInteractiveModel(db)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run interactive mode: %w", err)
	}

	return nil
}
