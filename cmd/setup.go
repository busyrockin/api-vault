package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/busyrockin/api-vault/core"
	"github.com/busyrockin/api-vault/ui"
	"github.com/spf13/cobra"
)

type setupModel struct {
	db              *core.Database
	step            int
	serviceOptions  []string
	selectedService int
	accountName     string
	apiKey          string
	cursor          int
	err             error
	done            bool
}

var serviceTemplates = []string{
	"OpenAI",
	"Anthropic",
	"Supabase",
	"Stripe",
	"GitHub",
	"Custom",
}

var serviceHints = map[string]struct{ account, key string }{
	"OpenAI":    {"production, development, personal", "Starts with sk-..."},
	"Anthropic": {"production, development, personal", "Starts with sk-ant-..."},
	"Supabase":  {"project name, e.g. my-app", "Found in Project Settings → API"},
	"Stripe":    {"live, test", "Starts with sk_live_ or sk_test_"},
	"GitHub":    {"username or org, e.g. octocat", "Personal access token (ghp_...)"},
	"Custom":    {"any label", "Paste your API key or token"},
}

func newSetupModel(db *core.Database) setupModel {
	return setupModel{
		db:             db,
		step:           0,
		serviceOptions: serviceTemplates,
	}
}

func (m setupModel) Init() tea.Cmd {
	return nil
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			return m.handleEnter()

		case "up", "k":
			if m.step == 0 && m.selectedService > 0 {
				m.selectedService--
			}

		case "down", "j":
			if m.step == 0 && m.selectedService < len(m.serviceOptions)-1 {
				m.selectedService++
			}

		case "backspace":
			if m.step == 1 && len(m.accountName) > 0 {
				m.accountName = m.accountName[:len(m.accountName)-1]
			} else if m.step == 2 && len(m.apiKey) > 0 {
				m.apiKey = m.apiKey[:len(m.apiKey)-1]
			}

		default:
			if len(msg.String()) == 1 {
				if m.step == 1 {
					m.accountName += msg.String()
				} else if m.step == 2 {
					m.apiKey += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m setupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0: // Service selection
		m.step = 1
		return m, nil

	case 1: // Account name
		if m.accountName == "" {
			m.err = fmt.Errorf("account name cannot be empty")
			return m, nil
		}
		m.step = 2
		m.err = nil
		return m, nil

	case 2: // API key
		if m.apiKey == "" {
			m.err = fmt.Errorf("API key cannot be empty")
			return m, nil
		}

		// Save credential
		service := m.serviceOptions[m.selectedService]
		name := fmt.Sprintf("%s-%s", strings.ToLower(service), m.accountName)

		if err := m.db.AddCredential(name, m.apiKey, strings.ToLower(service)); err != nil {
			m.err = fmt.Errorf("failed to save: %w", err)
			return m, nil
		}

		m.done = true
		m.err = nil
		return m, nil
	}

	return m, nil
}

func (m setupModel) View() string {
	if m.done {
		return m.renderSuccess()
	}

	var b strings.Builder

	b.WriteString(ui.TitleStyle.Render("🔐 Agent Vault"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	switch m.step {
	case 0:
		b.WriteString(m.renderServiceSelection())
	case 1:
		b.WriteString(m.renderAccountName())
	case 2:
		b.WriteString(m.renderAPIKey())
	}

	return ui.BoxStyle.Render(b.String())
}

func (m setupModel) renderServiceSelection() string {
	var b strings.Builder

	b.WriteString(ui.SubtitleStyle.Render("Select Service:"))
	b.WriteString("\n\n")

	for i, service := range m.serviceOptions {
		if i == m.selectedService {
			b.WriteString(ui.SelectedStyle.Render("❯ " + service))
		} else {
			b.WriteString(ui.NormalStyle.Render("  " + service))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("[↑↓] Navigate  [Enter] Select  [Esc] Cancel"))

	return b.String()
}

func (m setupModel) renderAccountName() string {
	var b strings.Builder

	service := m.serviceOptions[m.selectedService]
	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Service: %s", service)))
	b.WriteString("\n\n")

	b.WriteString("Account Name: ")
	b.WriteString(ui.Primary.Render(m.accountName + "_"))
	b.WriteString("\n\n")

	hint := serviceHints[service]
	b.WriteString(ui.Muted.Render("Examples: " + hint.account))
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[Type] Enter name  [Enter] Continue  [Esc] Cancel"))

	return b.String()
}

func (m setupModel) renderAPIKey() string {
	var b strings.Builder

	service := m.serviceOptions[m.selectedService]
	name := fmt.Sprintf("%s-%s", strings.ToLower(service), m.accountName)

	b.WriteString(ui.SubtitleStyle.Render(fmt.Sprintf("Credential: %s", name)))
	b.WriteString("\n\n")

	// Mask the API key
	masked := strings.Repeat("*", len(m.apiKey))
	b.WriteString("API Key: ")
	b.WriteString(ui.Primary.Render(masked + "_"))
	b.WriteString("\n\n")

	hint := serviceHints[service]
	b.WriteString(ui.Muted.Render(hint.key))
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("[Type] Enter key  [Enter] Save  [Esc] Cancel"))

	return b.String()
}

func (m setupModel) renderSuccess() string {
	var b strings.Builder

	service := m.serviceOptions[m.selectedService]
	name := fmt.Sprintf("%s-%s", strings.ToLower(service), m.accountName)

	b.WriteString(ui.TitleStyle.Render("✓ Credential Saved"))
	b.WriteString("\n\n")
	b.WriteString(ui.Success.Render(fmt.Sprintf("Successfully saved: %s", name)))
	b.WriteString("\n\n")
	b.WriteString(ui.SubtitleStyle.Render("You can now access this credential with:"))
	b.WriteString("\n")
	b.WriteString(ui.NormalStyle.Render(fmt.Sprintf("  api-vault get %s", name)))
	b.WriteString("\n\n")
	b.WriteString(ui.HelpStyle.Render("Press any key to exit"))

	return ui.BoxStyle.Render(b.String())
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive credential setup wizard",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openVault()
		if err != nil {
			return err
		}
		defer db.Close()

		m := newSetupModel(db)
		p := tea.NewProgram(m)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("setup wizard failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
