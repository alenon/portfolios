package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Portfolio represents a portfolio in the selector
type Portfolio struct {
	ID          uint
	Name        string
	Description string
	TotalValue  float64
}

// PortfolioSelector is a Bubble Tea model for selecting portfolios
type PortfolioSelector struct {
	portfolios []Portfolio
	cursor     int
	selected   int
	quitting   bool
}

// NewPortfolioSelector creates a new portfolio selector
func NewPortfolioSelector(portfolios []Portfolio) *PortfolioSelector {
	return &PortfolioSelector{
		portfolios: portfolios,
		selected:   -1,
	}
}

func (m *PortfolioSelector) Init() tea.Cmd {
	return nil
}

func (m *PortfolioSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.selected = m.cursor
			m.quitting = true
			return m, tea.Quit

		case "down", "j":
			if m.cursor < len(m.portfolios)-1 {
				m.cursor++
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

func (m *PortfolioSelector) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("13")).
		Bold(true).
		Padding(1, 0)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		Background(lipgloss.Color("235"))

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	s := titleStyle.Render("Select a Portfolio") + "\n\n"

	for i, portfolio := range m.portfolios {
		cursor := " "
		if m.cursor == i {
			cursor = "❯"
		}

		line := fmt.Sprintf("%s [%d] %s - $%.2f",
			cursor,
			portfolio.ID,
			portfolio.Name,
			portfolio.TotalValue,
		)

		if m.cursor == i {
			s += selectedStyle.Render(line) + "\n"
		} else {
			s += normalStyle.Render(line) + "\n"
		}
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • enter: select • q: quit")

	return s
}

// GetSelectedPortfolio returns the selected portfolio
func (m *PortfolioSelector) GetSelectedPortfolio() *Portfolio {
	if m.selected >= 0 && m.selected < len(m.portfolios) {
		return &m.portfolios[m.selected]
	}
	return nil
}

// RunPortfolioSelector runs the interactive portfolio selector
func RunPortfolioSelector(portfolios []Portfolio) (*Portfolio, error) {
	if len(portfolios) == 0 {
		return nil, fmt.Errorf("no portfolios available")
	}

	selector := NewPortfolioSelector(portfolios)
	p := tea.NewProgram(selector)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(*PortfolioSelector); ok {
		return m.GetSelectedPortfolio(), nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
