package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/neocotic/go-tic-tac-toe"
	"github.com/neocotic/go-tic-tac-toe/internal/bot"
	"os"
	"strconv"
	"strings"
)

type keyMap struct {
	down    key.Binding
	choose  key.Binding
	help    key.Binding
	left    key.Binding
	quit    key.Binding
	restart key.Binding
	right   key.Binding
	up      key.Binding
}

func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{km.help, km.restart, km.quit}
}

func (km keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.up, km.down, km.left, km.right, km.choose}, // First column
		{km.help, km.restart, km.quit},                 // Second column
	}
}

type styles struct {
	board          lipgloss.Style
	cell           lipgloss.Style
	cellError      lipgloss.Style
	cellFocus      lipgloss.Style
	help           lipgloss.Style
	message        lipgloss.Style
	messageDraw    lipgloss.Style
	messageForfeit lipgloss.Style
	messageWin     lipgloss.Style
}

type botTurnMsg struct {
	err    error
	player tictactoe.Player
	state  tictactoe.State
}

type botTurnStartedMsg struct{}

type model struct {
	botTurn          bool
	botTurnChan      chan botTurnMsg
	cursorX, cursorY uint8
	err              error
	forfeit          bool
	game             tictactoe.Game
	gameOver         bool
	help             help.Model
	keys             keyMap
	pack             tictactoe.Pack
	player           tictactoe.Player
	state            tictactoe.State
	styles           styles
	zone             *zone.Manager
	zoneIds          map[string]struct{}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("tic-tac-toe"), m.allowBotTurn())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case botTurnMsg:
		if !m.gameOver {
			if msg.err != nil {
				// Built-in bots should never cause errors to return
				panic(msg.err)
			}
			m.botTurn = false
			m.gameOver = msg.state != tictactoe.StateAwaitingTurn
			m.player = msg.player
			m.state = msg.state
		}
	case botTurnStartedMsg:
		if !m.gameOver {
			m.botTurn = true
			m.err = nil
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.choose):
			if !(m.botTurn || m.gameOver) {
				m.state, m.player, m.err = m.game.Play(tictactoe.Turn{
					Cell: tictactoe.Cell{
						Column: m.cursorX,
						Row:    m.cursorY,
					},
					Player: m.player,
				})
				m.gameOver = m.state != tictactoe.StateAwaitingTurn
				return m, m.allowBotTurn()
			}
		case key.Matches(msg, m.keys.up):
			if !(m.botTurn || m.gameOver) {
				m.err = nil
				if m.cursorY == 0 {
					m.cursorY = m.game.Size() - 1
				} else {
					m.cursorY--
				}
			}
		case key.Matches(msg, m.keys.down):
			if !(m.botTurn || m.gameOver) {
				m.err = nil
				if m.cursorY == m.game.Size()-1 {
					m.cursorY = 0
				} else {
					m.cursorY++
				}
			}
		case key.Matches(msg, m.keys.left):
			if !(m.botTurn || m.gameOver) {
				m.err = nil
				if m.cursorX == 0 {
					m.cursorX = m.game.Size() - 1
				} else {
					m.cursorX--
				}
			}
		case key.Matches(msg, m.keys.right):
			if !(m.botTurn || m.gameOver) {
				m.err = nil
				if m.cursorX == m.game.Size()-1 {
					m.cursorX = 0
				} else {
					m.cursorX++
				}
			}
		case key.Matches(msg, m.keys.restart):
			nm := initModel(m.pack, m.zone)
			return nm, nm.allowBotTurn()
		case key.Matches(msg, m.keys.help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.quit):
			return m, tea.Quit
		}
	case tea.MouseMsg:
		switch msg.Action {
		case tea.MouseActionMotion:
			if !(m.botTurn || m.gameOver) {
				if row, col, found := m.findCellZone(msg); found {
					m.cursorX = col
					m.cursorY = row
				}
			}
		case tea.MouseActionRelease:
			if !(m.botTurn || m.gameOver) && msg.Button == tea.MouseButtonLeft {
				if row, col, found := m.findCellZone(msg); found {
					m.cursorX = col
					m.cursorY = row
					m.state, m.player, m.err = m.game.Play(tictactoe.Turn{
						Cell: tictactoe.Cell{
							Column: m.cursorX,
							Row:    m.cursorY,
						},
						Player: m.player,
					})
					m.gameOver = m.state != tictactoe.StateAwaitingTurn
					return m, m.allowBotTurn()
				}
			}
		default:
			// Do nothing
		}
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
	}
	return m, nil
}

func (m model) View() string {
	b := m.styles.board.Render(m.renderBoard())
	var msg string
	if m.forfeit {
		msg = m.styles.messageForfeit.Render(m.renderPlayer() + " FORFEITS!")
	} else if m.gameOver {
		switch m.state {
		case tictactoe.StateDraw:
			msg = m.styles.messageDraw.Render("DRAW!")
		case tictactoe.StateWon:
			msg = m.styles.messageWin.Render(m.renderPlayer() + " WINS!")
		default:
			panic(fmt.Errorf("unexpected final game state: %v", m.state))
		}
	} else if m.botTurn {
		msg = m.styles.message.Render(m.renderPlayer() + " THINKING...")
	} else {
		msg = m.styles.message.Render("READY " + m.renderPlayer())
	}
	h := m.styles.help.Render(m.help.View(m.keys))
	return m.zone.Scan(lipgloss.JoinVertical(lipgloss.Top, b, msg, h))
}

func (m model) allowBotTurn() tea.Cmd {
	if !m.game.IsBotTurn() {
		return nil
	}
	return tea.Batch(startBotTurn(m.botTurnChan, m.game), awaitBotTurn(m.botTurnChan))
}

func (m model) findCellZone(msg tea.MouseMsg) (uint8, uint8, bool) {
	for id := range m.zoneIds {
		if m.zone.Get(id).InBounds(msg) {
			if row, col, err := m.parseCellZoneId(id); err != nil {
				panic(err)
			} else {
				return row, col, true
			}
		}
	}
	return 0, 0, false
}

func (m model) markCellZone(row, col int, value string) string {
	id := fmt.Sprintf("cell:%d %d", col, row)
	m.zoneIds[id] = struct{}{}
	return m.zone.Mark(id, value)
}

func (m model) parseCellZoneId(id string) (uint8, uint8, error) {
	coords, found := strings.CutPrefix(id, "cell:")
	if !found {
		return 0, 0, fmt.Errorf("unexpected cell zone ID: %q", id)
	}

	fields := strings.SplitN(coords, " ", 2)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("malformed cell zone ID: %q", id)
	}

	col, err := strconv.ParseUint(fields[0], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column in cell zone ID: %q", id)
	}

	row, err := strconv.ParseUint(fields[1], 10, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid row in cell zone ID: %q", id)
	}

	return uint8(row), uint8(col), nil
}

func (m model) renderBoard() string {
	board := m.game.Board()
	size := m.game.Size()
	rows := make([]string, size)
	for row, cols := range board {
		cells := make([]string, size)
		for col, player := range cols {
			var style lipgloss.Style
			if row == int(m.cursorY) && col == int(m.cursorX) {
				if m.err != nil {
					style = m.styles.cellError
				} else {
					style = m.styles.cellFocus
				}
			} else {
				style = m.styles.cell
			}
			cells[col] = m.markCellZone(row, col, style.Render(player.String()))
		}
		rows[row] = lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) renderPlayer() string {
	switch m.player {
	case tictactoe.PlayerOne:
		return "PLAYER ONE"
	case tictactoe.PlayerTwo:
		return "PLAYER TWO"
	default:
		// Should never happen
		return "PLAYER UNKNOWN"
	}
}

func initModel(pack tictactoe.Pack, zm *zone.Manager) model {
	g := tictactoe.MustStart(tictactoe.WithPack(pack))
	p, s := g.Player(), g.State()

	km := keyMap{
		choose: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		down: key.NewBinding(
			key.WithKeys("down", "s"),
			key.WithHelp("↓/s", "move down"),
		),
		help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		left: key.NewBinding(
			key.WithKeys("left", "a"),
			key.WithHelp("←/a", "move left"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		right: key.NewBinding(
			key.WithKeys("right", "d"),
			key.WithHelp("→/d", "move right"),
		),
		up: key.NewBinding(
			key.WithKeys("up", "w"),
			key.WithHelp("↑/w", "move up"),
		),
	}

	cst := lipgloss.NewStyle().
		Width(15).
		Height(5).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Background(lipgloss.ANSIColor(15)).
		Foreground(lipgloss.ANSIColor(0)).
		Border(lipgloss.OuterHalfBlockBorder(), true, true, true, true).
		BorderForeground(lipgloss.ANSIColor(15))
	mst := lipgloss.NewStyle().
		Width((cst.GetHorizontalFrameSize()+cst.GetWidth())*int(g.Size())).
		Height(1).
		Margin(0, 1, 1, 1).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).
		Background(lipgloss.ANSIColor(15)).
		Foreground(lipgloss.ANSIColor(0))

	st := styles{
		board: lipgloss.NewStyle().
			Margin(1, 1, 0, 1),
		cell: cst,
		cellError: cst.
			Background(lipgloss.ANSIColor(9)).
			Foreground(lipgloss.ANSIColor(1)),
		cellFocus: cst.
			Background(lipgloss.ANSIColor(33)).
			Foreground(lipgloss.ANSIColor(4)),
		help: lipgloss.NewStyle().
			Margin(0, 1),
		message: mst,
		messageDraw: mst.
			Background(lipgloss.ANSIColor(3)).
			Foreground(lipgloss.ANSIColor(11)),
		messageForfeit: mst.
			Background(lipgloss.ANSIColor(9)).
			Foreground(lipgloss.ANSIColor(1)),
		messageWin: mst.
			Background(lipgloss.ANSIColor(10)).
			Foreground(lipgloss.ANSIColor(22)),
	}

	h := help.New()
	h.Styles.FullKey.Bold(true)
	h.Styles.ShortKey.Bold(true)

	return model{
		botTurnChan: make(chan botTurnMsg),
		game:        g,
		help:        h,
		keys:        km,
		pack:        pack,
		player:      p,
		state:       s,
		styles:      st,
		zone:        zm,
		zoneIds:     make(map[string]struct{}),
	}
}

const (
	flagNameBot     = "bot"
	flagNameHelp    = "help"
	flagNameNoMouse = "no-mouse"
	flagNamePlayer  = "player"
	flagNameSize    = "size"

	flagInvalidReasonBotMaxSizeExceeded = "bot max board size exceeded"
	flagInvalidReasonOutOfRange         = "value out of range"
	flagInvalidReasonParse              = "parse error"
)

func awaitBotTurn(ch chan botTurnMsg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func handleInvalidFlag(name string, value any, reason string) {
	fmt.Printf(`invalid value "%v" for flag -%s: %s
`, value, name, reason)
	flag.Usage()
	os.Exit(2)
}

func startBotTurn(ch chan botTurnMsg, game tictactoe.Game) tea.Cmd {
	return func() tea.Msg {
		go func() {
			state, player, err := game.AllowBotTurn()
			ch <- botTurnMsg{
				err:    err,
				player: player,
				state:  state,
			}
		}()

		return botTurnStartedMsg{}
	}
}

func main() {
	var (
		botFlag               string
		helpFlag, noMouseFlag bool
		playerFlag, sizeFlag  uint
	)

	flag.StringVar(&botFlag, flagNameBot, "", `enable bot opponent with difficulty (e.g. "normal")`)
	flag.BoolVar(&helpFlag, flagNameHelp, false, "print help")
	flag.BoolVar(&noMouseFlag, flagNameNoMouse, false, "disable mouse support")
	flag.UintVar(&playerFlag, flagNamePlayer, 1, "starter player")
	flag.UintVar(&sizeFlag, flagNameSize, 3, "size of board")
	flag.Parse()

	if helpFlag {
		flag.Usage()
		return
	}

	var player tictactoe.Player
	if playerFlag == 0 || playerFlag > 2 {
		handleInvalidFlag(flagNamePlayer, playerFlag, flagInvalidReasonOutOfRange)
	} else {
		player = tictactoe.Player(playerFlag)
	}

	var size uint8
	if sizeFlag < uint(tictactoe.MinSize) || sizeFlag > uint(tictactoe.MaxSize) {
		handleInvalidFlag(flagNameSize, sizeFlag, flagInvalidReasonOutOfRange)
	} else {
		size = uint8(sizeFlag)
	}

	var maxSize uint8
	pack := tictactoe.Pack{tictactoe.WithSize(size), tictactoe.WithStarterPlayer(player)}
	switch botFlag {
	case "":
		// Do nothing
	case bot.NameEasy:
		pack = append(pack, tictactoe.WithEasyBot(tictactoe.PlayerTwo))
		maxSize = bot.MaxSizeEasy
	case bot.NameNormal:
		pack = append(pack, tictactoe.WithNormalBot(tictactoe.PlayerTwo))
		maxSize = bot.MaxSizeNormal
	case bot.NameHard:
		pack = append(pack, tictactoe.WithHardBot(tictactoe.PlayerTwo))
		maxSize = bot.MaxSizeHard
	case bot.NameImpossible:
		pack = append(pack, tictactoe.WithImpossibleBot(tictactoe.PlayerTwo))
		maxSize = bot.MaxSizeImpossible
	default:
		handleInvalidFlag(flagNameBot, botFlag, flagInvalidReasonParse)
	}

	if botFlag != "" && maxSize > 0 && size > maxSize {
		handleInvalidFlag(flagNameSize, sizeFlag, fmt.Sprintf("%q %s (%v)", botFlag, flagInvalidReasonBotMaxSizeExceeded, maxSize))
	}

	zm := zone.New()
	zm.SetEnabled(!noMouseFlag)
	defer zm.Close()

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if !noMouseFlag {
		opts = append(opts, tea.WithMouseAllMotion())
	}

	p := tea.NewProgram(initModel(pack, zm), opts...)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
