package app

import tea "github.com/charmbracelet/bubbletea"

const (
	TAB_ID = "ruppi_tab_id_"
)

type updateURL string
type newTabMsg string
type changeTabMsg int

func updateURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return updateURL(url)
	}
}

func createNewTabCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return newTabMsg(url)
	}
}

func createChangeTabCmd(tabId int) tea.Cmd {
	return func() tea.Msg {
		return changeTabMsg(tabId)
	}
}
