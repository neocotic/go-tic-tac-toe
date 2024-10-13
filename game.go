package tictactoe

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const (
	// MaxSize is the maximum size of a Board
	MaxSize uint8 = math.MaxUint8
	// MinSize is the minimum size of a Board
	MinSize uint8 = 3
)

// Board contains all player turns
type Board [][]Player

// Copy returns a deep copy of Board
func (b Board) Copy() Board {
	if b == nil {
		return nil
	}
	l := len(b)
	c := make(Board, l)
	for row, cols := range b {
		c[row] = make([]Player, l)
		copy(c[row], cols)
	}
	return c
}

// FindEmpty returns all Cells within Board that do not contain a Player
func (b Board) FindEmpty() Cells {
	var cells Cells
	for row, cols := range b {
		for col, player := range cols {
			if player == 0 {
				cells = append(cells, Cell{
					Column: uint8(col),
					Row:    uint8(row),
				})
			}
		}
	}
	return cells
}

// String returns a classic ASCII representation of Board
func (b Board) String() string {
	var (
		sb   strings.Builder
		size = len(b)
	)
	for row, cols := range b {
		sb.WriteString("|")
		for _, player := range cols {
			sb.WriteRune(' ')
			sb.WriteString(player.String())
			sb.WriteString(" |")
		}
		if row < size-1 {
			sb.WriteString("\n|")
			for i := 0; i < size; i++ {
				if i > 0 {
					sb.WriteRune('+')
				}
				sb.WriteString("---")
			}
			sb.WriteString("|\n")
		}
	}
	return sb.String()
}

func (b Board) check() (starter Player, size uint8, maxTurns int, turns []Turn, err error) {
	if l := len(b); l < int(MinSize) {
		err = fmt.Errorf("board must contain at least %d rows", MinSize)
		return
	} else if l > int(MaxSize) {
		err = fmt.Errorf("board must contain at most %d rows", MaxSize)
		return
	} else {
		maxTurns = l * l
		size = uint8(l)
	}

	turnCounter := make(map[Player]uint16, 2)
	for row, cols := range b {
		if l := len(cols); l != int(size) {
			err = fmt.Errorf("board row[%d] must contain %d columns", row, size)
			return
		}

		for col, player := range cols {
			if player == 0 {
				continue
			}
			if player.IsValid() {
				turns = append(turns, Turn{
					Cell: Cell{
						Column: uint8(col),
						Row:    uint8(row),
					},
					Player: player,
				})
				turnCounter[player] = turnCounter[player] + 1
			} else {
				err = fmt.Errorf("board cell[%d,%d] contains unknown player: %d", row, col, player)
				return
			}
		}
	}

	p1Turns, p2Turns := turnCounter[PlayerOne], turnCounter[PlayerTwo]
	diffTurns := int(p1Turns) - int(p2Turns)

	switch {
	case diffTurns == 1:
		starter = PlayerTwo
	case diffTurns > 1:
		err = fmt.Errorf("board contains unfair advantage for player: %d", PlayerOne)
	case diffTurns == -1:
		starter = PlayerOne
	case diffTurns < -1:
		err = fmt.Errorf("board contains unfair advantage for player: %d", PlayerTwo)
	}
	return
}

func newBoard(size uint8) Board {
	board := make(Board, size)
	for i := range board {
		board[i] = make([]Player, size)
	}
	return board
}

type (
	// Cell represents the location of a cell on a Board
	Cell struct {
		// Column is the column of Cell
		Column uint8
		// Row is the row of Cell
		Row uint8
	}

	// Cells contains multiple Board cells
	Cells []Cell
)

// Random returns a random Cell within Cells or an empty Cell if Cells is empty
func (cs Cells) Random() Cell {
	if len(cs) == 0 {
		return Cell{}
	}
	return cs[rand.Intn(len(cs))]
}

type (
	// Condition represents a check for a specific winning condition
	Condition interface {
		// FindWinner checks the given Board and returns the Player that wins based on the Condition or zero if there is
		// no winner.
		//
		// The Board provided is not a copy so a Condition must never mutate it or risk corrupting the Game.
		FindWinner(board Board) Player
		// IsWinningTurn checks the given Board and returns whether Turn provided resulted in a win based on the
		// Condition.
		//
		// The Board provided is not a copy so a Condition must never mutate it or risk corrupting the Game.
		IsWinningTurn(board Board, turn Turn) bool
	}

	// Conditions contains multiple winning conditions
	Conditions []Condition
)

// FindWinner checks the given Board and returns the Player that wins based on any of the Conditions or zero if there is
// no winner.
//
// The Board provided is not a copy so if a Condition mutates it they will corrupt the Game.
//
// An ErrConditionInvalid is returned if a Condition returns an invalid non-zero Player.
func (cs Conditions) FindWinner(board Board) (Player, error) {
	for i, c := range cs {
		if player := c.FindWinner(board); !player.IsValidOrZero() {
			return 0, fmtInvalidConditionErr(i, fmt.Sprintf("invalid player: %v", player))
		} else if player > 0 {
			return player, nil
		}
	}
	return 0, nil
}

// IsWinningTurn checks the given Board and returns whether the Turn provided resulted in a win based on any of the
// Conditions.
//
// The Board provided is not a copy so if a Condition mutates it they will corrupt the Game.
func (cs Conditions) IsWinningTurn(board Board, turn Turn) bool {
	for _, c := range cs {
		if won := c.IsWinningTurn(board, turn); won {
			return won
		}
	}
	return false
}

var standardConditions = Conditions{&horizontalCondition{}, &verticalCondition{}, &diagonalCondition{}}

type diagonalCondition struct{}

func (c *diagonalCondition) FindWinner(board Board) Player {
	if winner := c.findWinnerFromLeft(board); winner > 0 {
		return winner
	}
	if winner := c.findWinnerFromRight(board); winner > 0 {
		return winner
	}
	return 0
}

func (c *diagonalCondition) IsWinningTurn(board Board, turn Turn) bool {
	if turn.Row == turn.Column && c.isWinningTurnFromLeft(board, turn) {
		return true
	}
	if turn.Row == uint8(len(board))-turn.Column-1 && c.isWinningTurnFromRight(board, turn) {
		return true
	}
	return false
}

func (c *diagonalCondition) findWinnerFromLeft(board Board) (player Player) {
	for i, cols := range board {
		current := cols[i]
		if current == 0 {
			return 0
		}
		if player == 0 {
			player = current
		} else if current != player {
			return 0
		}
	}
	return
}

func (c *diagonalCondition) findWinnerFromRight(board Board) (player Player) {
	l := len(board)
	for i, cols := range board {
		current := cols[l-i-1]
		if current == 0 {
			return 0
		}
		if player == 0 {
			player = current
		} else if current != player {
			return 0
		}
	}
	return
}

func (c *diagonalCondition) isWinningTurnFromLeft(board Board, turn Turn) bool {
	for i, cols := range board {
		if cols[i] != turn.Player {
			return false
		}
	}
	return true
}

func (c *diagonalCondition) isWinningTurnFromRight(board Board, turn Turn) bool {
	l := len(board)
	for i, cols := range board {
		if cols[l-i-1] != turn.Player {
			return false
		}
	}
	return true
}

type horizontalCondition struct{}

func (c *horizontalCondition) FindWinner(board Board) Player {
	for _, cols := range board {
		if winner := c.findWinnerFromRow(cols); winner > 0 {
			return winner
		}
	}
	return 0
}

func (c *horizontalCondition) IsWinningTurn(board Board, turn Turn) bool {
	for _, current := range board[turn.Row] {
		if current != turn.Player {
			return false
		}
	}
	return true
}

func (c *horizontalCondition) findWinnerFromRow(cols []Player) (player Player) {
	for _, current := range cols {
		if current == 0 {
			return 0
		}
		if player == 0 {
			player = current
		} else if player != current {
			return 0
		}
	}
	return
}

type verticalCondition struct{}

func (c *verticalCondition) FindWinner(board Board) Player {
	for i := range board {
		if winner := c.findWinnerFromCol(board, uint8(i)); winner > 0 {
			return winner
		}
	}
	return 0
}

func (c *verticalCondition) IsWinningTurn(board Board, turn Turn) bool {
	for _, cols := range board {
		if cols[turn.Column] != turn.Player {
			return false
		}
	}
	return true
}

func (c *verticalCondition) findWinnerFromCol(board Board, col uint8) (player Player) {
	for _, cols := range board {
		current := cols[col]
		if current == 0 {
			return 0
		}
		if player == 0 {
			player = current
		} else if player != current {
			return 0
		}
	}
	return
}

var (
	// ErrBot is returned if a Bot fails to take their turn
	ErrBot = errors.New("bot turn failed")
	// ErrConditionInvalid is returned if a Condition returns invalid information
	ErrConditionInvalid = errors.New("invalid condition")
	// ErrGameOver is returned if attempting to take a turn while not having StateAwaitingTurn
	ErrGameOver = errors.New("game over")
	// ErrOptionInvalid is returned if an Option is passed to Start that has been given an invalid argument
	ErrOptionInvalid = errors.New("invalid option")
	// ErrOutOfBounds is returned if a given row/column is out-of-bounds
	ErrOutOfBounds = errors.New("out of bounds")
	// ErrPlayerNotFound is returned if a given Player cannot be found
	ErrPlayerNotFound = errors.New("player not found")
	// ErrTurnInvalid is returned if attempting to take an invalid turn
	ErrTurnInvalid = errors.New("invalid turn")
)

func fmtBotErr(err error) error {
	return fmt.Errorf("%w: %w", ErrBot, err)
}

func fmtColOutOfBoundsErr(cell Cell, size uint8) error {
	return fmt.Errorf("%w: row[%d]col[%d] is greater than or equal to %d", ErrOutOfBounds, cell.Row, cell.Column, size)
}

func fmtInvalidConditionErr(idx int, reason string) error {
	return fmt.Errorf("%w[%d]: %s", ErrConditionInvalid, idx, reason)
}

func fmtInvalidOptionErr(option string, err error) error {
	return fmt.Errorf("%w[%s]: %w", ErrOptionInvalid, option, err)
}

func fmtInvalidTurnErr(reason string) error {
	return fmt.Errorf("%w: %s", ErrTurnInvalid, reason)
}

func fmtPlayerNotFoundErr(player Player) error {
	return fmt.Errorf("%w: %d", ErrPlayerNotFound, player)
}

func fmtRowOutOfBoundsErr(cell Cell, size uint8) error {
	return fmt.Errorf("%w: row[%d] is greater than or equal to %d", ErrOutOfBounds, cell.Row, size)
}

type (
	// Game represents a single session of the tic-tac-toe game
	Game interface {
		// AllowBotTurn requests a turn from a Bot, where applicable, and plays that Turn.
		//
		// Nothing happens if Game doesn't have StateAwaitingTurn, has no Bot, or it's not the turn of the Bot.
		//
		// An ErrBot is returned if the Bot fails to take their turn or their turn is invalid due to the same
		// constraints as applied to Play.
		AllowBotTurn() (State, Player, error)
		// Board returns a copy of the Board
		Board() Board
		// Conditions returns a copy of the winning conditions for Game
		Conditions() Conditions
		// IsBotTurn returns whether Game has a Bot, and it's their turn.
		//
		// If Game does not have StateAwaitingTurn, false will always be returned.
		IsBotTurn() bool
		// LastTurn returns the last Turn played, where possible
		LastTurn() (Turn, bool)
		// MaxTurns returns the maximum number of turns allowed
		MaxTurns() int
		// Play takes the given Turn and returns the resulting State and Player.
		//
		// When playing against a Bot opponent, it's recommended to simply call AllowBotTurn after each call to Play to
		// ensure that the Bot has an opportunity to take their turn. AllowBotTurn is designed to only request a turn
		// from the Bot when it's appropriate to do so.
		//
		// The resulting Player will vary depending on State:
		//  - For StateAwaitingTurn it's the Player to take the next turn
		//  - For StateDraw it's zero
		//  - For StateWon it's the given Player who's won
		//
		// An error is returned in following cases:
		//  - ErrGameOver if Game doesn't have StateAwaitingTurn
		//  - ErrOutOfBounds if Turn's Cell is out-of-bounds
		//  - ErrPlayerNotFound if Turn's Player is invalid
		//  - ErrTurnInvalid if Turn is invalid (e.g. not turn of Player, Cell taken)
		Play(turn Turn) (State, Player, error)
		// Player returns the current Player, where appropriate.
		//
		// The Player will vary depending on Game's State:
		//  - For StateAwaitingTurn it's the Player to take the next turn
		//  - For StateDraw it's zero
		//  - For StateWon it's the winning Player
		Player() Player
		// PlayerAt returns the Player at the given Cell, where possible.
		//
		// An ErrOutOfBounds is returned if Cell is out-of-bounds.
		PlayerAt(cell Cell) (Player, error)
		// RemainingTurns returns the number of turns remaining
		RemainingTurns() int
		// Size returns the size of the Board
		Size() uint8
		// State returns the current State
		State() State
		// String returns a classic ASCII representation of the Board
		String() string
		// Turns returns a copy of each Turn already played
		Turns() []Turn
	}

	game struct {
		board      Board
		bot        Bot
		conditions Conditions
		maxTurns   int
		player     Player
		size       uint8
		state      State
		turns      []Turn
	}
)

func (g *game) AllowBotTurn() (State, Player, error) {
	if !g.IsBotTurn() {
		return g.state, g.player, nil
	}
	cell, err := g.bot.Turn(g.board, g)
	if err != nil {
		return g.state, g.player, fmtBotErr(err)
	}
	_, _, err = g.play(Turn{Cell: cell, Player: g.player}, true)
	if err != nil {
		err = fmtBotErr(err)
	}
	return g.state, g.player, err
}

func (g *game) Board() Board {
	return g.board.Copy()
}

func (g *game) Conditions() Conditions {
	return g.conditions[:]
}

func (g *game) IsBotTurn() bool {
	return g.state == StateAwaitingTurn && g.bot != nil && g.bot.Player() == g.player
}

func (g *game) LastTurn() (Turn, bool) {
	if l := len(g.turns); l == 0 {
		return Turn{}, false
	} else {
		return g.turns[l-1], true
	}
}

func (g *game) MaxTurns() int {
	return g.maxTurns
}

func (g *game) Play(turn Turn) (State, Player, error) {
	return g.play(turn, false)
}

func (g *game) Player() Player {
	return g.player
}

func (g *game) PlayerAt(cell Cell) (Player, error) {
	if err := g.validateBounds(cell); err != nil {
		return 0, err
	}
	return g.board[cell.Row][cell.Column], nil
}

func (g *game) RemainingTurns() int {
	return g.maxTurns - len(g.turns)
}

func (g *game) Size() uint8 {
	return g.size
}

func (g *game) State() State {
	return g.state
}

func (g *game) String() string {
	return g.board.String()
}

func (g *game) Turns() []Turn {
	return g.turns[:]
}

func (g *game) play(turn Turn, allowBotTurn bool) (State, Player, error) {
	if err := g.validateBounds(turn.Cell); err != nil {
		return g.state, g.player, err
	}
	if err := g.validateTurn(turn, allowBotTurn); err != nil {
		return g.state, g.player, err
	}

	g.board[turn.Row][turn.Column] = turn.Player
	g.turns = append(g.turns, turn)

	if g.conditions.IsWinningTurn(g.board, turn) {
		g.state = StateWon
	} else if len(g.turns) >= g.maxTurns {
		g.player = 0
		g.state = StateDraw
	} else {
		g.player = turn.Player.Next()
	}

	return g.state, g.player, nil
}

func (g *game) validateBounds(cell Cell) error {
	if cell.Row >= g.size {
		return fmtRowOutOfBoundsErr(cell, g.size)
	}
	if cell.Column >= g.size {
		return fmtColOutOfBoundsErr(cell, g.size)
	}
	return nil
}

func (g *game) validateTurn(turn Turn, allowBotTurn bool) error {
	if g.state != StateAwaitingTurn {
		return ErrGameOver
	}
	player := turn.Player
	if !player.IsValid() {
		return fmtPlayerNotFoundErr(player)
	}
	if g.bot != nil && g.bot.Player() == player && !allowBotTurn {
		return fmtInvalidTurnErr(fmt.Sprintf("human cannot play turn for bot player[%d]", player))
	}
	if player != g.player {
		return fmtInvalidTurnErr(fmt.Sprintf("player[%d] cannot play turn for player[%d]", player, g.player))
	}
	row, col := turn.Row, turn.Column
	if existing := g.board[row][col]; existing > 0 {
		return fmtInvalidTurnErr(fmt.Sprintf("cell[%d,%d] already taken by player %d", row, col, existing))
	}
	return nil
}

// MustStart is a convenient shorthand for calling Start whilst panicking if it returns an error
func MustStart(opts ...Option) Game {
	if g, err := Start(opts...); err != nil {
		panic(err)
	} else {
		return g
	}
}

// Start returns a new Game, optionally customized by providing options.
//
// An error is returned in following cases:
//   - ErrConditionInvalid if a winning Condition returns an invalid winning Player
//   - ErrOptionInvalid if an Option is passed that was given an invalid argument
func Start(opts ...Option) (Game, error) {
	g := &game{
		maxTurns:   int(MinSize) * int(MinSize),
		conditions: standardConditions[:],
		size:       MinSize,
		state:      StateAwaitingTurn,
	}

	for _, opt := range opts {
		if err := opt(g); err != nil {
			return nil, err
		}
	}

	if g.board == nil {
		g.board = newBoard(g.size)
		if g.player == 0 {
			g.player = PlayerOne
		}
	} else if winner, err := g.conditions.FindWinner(g.board); err != nil {
		return nil, err
	} else if winner > 0 {
		g.state = StateWon
		g.player = winner
	} else if len(g.turns) >= g.maxTurns {
		g.state = StateDraw
		g.player = 0
	} else if g.player == 0 {
		g.player = PlayerOne
	}

	return g, nil
}

type (
	// Option is used to customize Game
	Option func(g *game) error

	// Pack provides a convenient method for bundling more than one Option, typically to represent a game mode
	Pack []Option
)

// WithBoard customizes a Game to use the given Board.
//
// The following game parameters are derived from board if valid:
//   - Player (e.g. starter, winner), where possible
//   - Size
//   - State
//   - Turns
//
// This option takes precedence over size-controlling options and any player-controlling options are only used if a
// "correct" starting player cannot be derived from Board. Any options that register one or more additional winning
// Condition will be honored when checking whether board has been won.
//
// An ErrOptionInvalid is returned by the option if board is invalid. For example;
//   - Length is not within the valid range (i.e. MinSize, MaxSize)
//   - Contains row with number of columns not equaling length of board
//   - Contains cell with an invalid non-zero Player
//   - Either Player has an unfair advantage (i.e. more than one turn ahead of the other)
func WithBoard(board Board) Option {
	return func(g *game) error {
		starter, size, maxTurns, turns, err := board.check()
		if err != nil {
			return fmtInvalidOptionErr("WithBoard", err)
		}
		g.board = board
		g.maxTurns = maxTurns
		g.size = size
		g.turns = turns
		if starter > 0 {
			g.player = starter
		}
		return nil
	}
}

// WithBot customizes a Game to play against the given Bot.
//
// This option is ignored if preceded by another bot-controlling option (e.g. WithNormalBot).
//
// An ErrOptionInvalid is returned by the option if bot has an invalid Player.
func WithBot(bot Bot) Option {
	return withBot(bot, "WithBot")
}

// WithCondition customizes a Game to include an additional winning Condition
func WithCondition(condition Condition) Option {
	return func(g *game) error {
		g.conditions = append(g.conditions, condition)
		return nil
	}
}

// WithConditions customizes a Game to include additional winning Conditions
func WithConditions(conditions Conditions) Option {
	return func(g *game) error {
		g.conditions = append(g.conditions, conditions...)
		return nil
	}
}

// WithEasyBot is a convenient shorthand for WithBot(NewEasyBot(player)).
//
// This option is ignored if preceded by another bot-controlling option (e.g. WithNormalBot).
//
// An ErrOptionInvalid is returned by the option if player is invalid.
func WithEasyBot(player Player) Option {
	return withBot(NewEasyBot(player), "WithEasyBot")
}

// WithHardBot is a convenient shorthand for WithBot(NewHardBot(player)).
//
// This option is ignored if preceded by another bot-controlling option (e.g. WithNormalBot).
//
// An ErrOptionInvalid is returned by the option if player is invalid.
func WithHardBot(player Player) Option {
	return withBot(NewHardBot(player), "WithHardBot")
}

// WithImpossibleBot is a convenient shorthand for WithBot(NewImpossibleBot(player)).
//
// This option is ignored if preceded by another bot-controlling option (e.g. WithNormalBot).
//
// An ErrOptionInvalid is returned by the option if player is invalid.
func WithImpossibleBot(player Player) Option {
	return withBot(NewImpossibleBot(player), "WithImpossibleBot")
}

// WithNormalBot is a convenient shorthand for WithBot(NewNormalBot(player)).
//
// This option is ignored if preceded by another bot-controlling option (e.g. WithHardBot).
//
// An ErrOptionInvalid is returned by the option if player is invalid.
func WithNormalBot(player Player) Option {
	return withBot(NewNormalBot(player), "WithNormalBot")
}

// WithPack customizes a Game by applying the given Pack
func WithPack(pack Pack) Option {
	return func(g *game) error {
		for _, opt := range pack {
			if err := opt(g); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithRandomSize customizes a Game to create a Board with a random size.
//
// This option is ignored if preceded by another size-controlling option (e.g. WithSize) or if WithBoard is also used.
func WithRandomSize() Option {
	return func(g *game) error {
		if g.board != nil {
			return nil
		}
		size := rand.Intn(int(MaxSize)-int(MinSize)) + int(MinSize)
		g.maxTurns = size * size
		g.size = uint8(size)
		return nil
	}
}

// WithRandomStarterPlayer customizes a Game to start with a random Player.
//
// This option is ignored if preceded by another player-controlling option (e.g. WithStarterPlayer) or if WithBoard is
// also used and a "correct" starting player was derived.
func WithRandomStarterPlayer() Option {
	return func(g *game) error {
		if g.player == 0 {
			g.player = Player(rand.Intn(2) + 1)
		}
		return nil
	}
}

// WithSize customizes a Game to create a Board with the given size.
//
// This option is ignored if preceded by another size-controlling option (e.g. WithRandomStarterPlayer) or if WithBoard
// is also used.
//
// An ErrOptionInvalid is returned by the option if size is not within the valid range (i.e. MinSize, MaxSize).
func WithSize(size uint8) Option {
	return func(g *game) error {
		if g.board != nil {
			return nil
		}
		if size < MinSize {
			return fmtInvalidOptionErr("WithSize", fmt.Errorf("size must be at least: %d", MinSize))
		}
		if size > MaxSize {
			return fmtInvalidOptionErr("WithSize", fmt.Errorf("size must be at most: %d", MaxSize))
		}
		g.maxTurns = int(size) * int(size)
		g.size = size
		return nil
	}
}

// WithStarterPlayer customizes a Game to start with the given Player.
//
// This option is ignored if preceded by another player-controlling option (e.g. WithRandomStarterPlayer) or if
// WithBoard is also used and a "correct" starting player was derived.
//
// An ErrOptionInvalid is returned by the option if player is invalid.
func WithStarterPlayer(player Player) Option {
	return func(g *game) error {
		if g.player > 0 {
			return nil
		}
		if !player.IsValid() {
			return fmtInvalidOptionErr("WithStarterPlayer", fmtPlayerNotFoundErr(player))
		}
		g.player = player
		return nil
	}
}

func withBot(bot Bot, option string) Option {
	return func(g *game) error {
		if g.bot != nil {
			return nil
		}
		if player := bot.Player(); !player.IsValid() {
			return fmtInvalidOptionErr(option, fmtPlayerNotFoundErr(player))
		}
		g.bot = bot
		return nil
	}
}

// Player represents a player of a Game
type Player uint8

const (
	// PlayerOne represents the first player (X)
	PlayerOne Player = iota + 1
	// PlayerTwo represents the second player (O)
	PlayerTwo
)

// IsValid returns whether Player is valid.
//
// As zero is used to denote a non-existent Player, it is not considered a valid Player. To allow zero use IsValidOrZero
// instead.
func (p Player) IsValid() bool {
	switch p {
	case PlayerOne, PlayerTwo:
		return true
	default:
		return false
	}
}

// IsValidOrZero returns whether Player is valid or zero, with the latter used to denote a non-existent Player.
func (p Player) IsValidOrZero() bool {
	return p == 0 || p.IsValid()
}

// Next returns the next logical Player.
//
// PlayerOne is returned for PlayerTwo and vice versa. Any other (invalid) value will return itself.
func (p Player) Next() Player {
	switch p {
	case PlayerOne:
		return PlayerTwo
	case PlayerTwo:
		return PlayerOne
	default:
		return p
	}
}

// String returns a simple string representation of Player.
//
// "X" is returned for PlayerOne and "O" is returned for PlayerTwo. An empty space (" ") is returned for zero (used to
// denote a non-existent Player), otherwise, a question mark ("?") is returned to represent an unknown Player.
func (p Player) String() string {
	switch p {
	case 0:
		return " "
	case PlayerOne:
		return "X"
	case PlayerTwo:
		return "O"
	default:
		return "?"
	}
}

// Players returns valid Player values
func Players() []Player {
	return []Player{PlayerOne, PlayerTwo}
}

// Turn represents a turn that is either to be taken or has already been taken
type Turn struct {
	// Cell is the location of the cell on the Board
	Cell
	// Player is the Player
	Player Player
}

// State represents the state of a Game
type State uint8

const (
	// StateAwaitingTurn represents the state in which the game is waiting for a player to take a turn
	StateAwaitingTurn State = iota
	// StateDraw represents the state in which the game is over and there is no winner
	StateDraw
	// StateWon represents the state in which the game is over with a clear winner
	StateWon
)

// IsValid returns whether State is valid
func (s State) IsValid() bool {
	switch s {
	case StateAwaitingTurn, StateDraw, StateWon:
		return true
	default:
		return false
	}
}

// String returns a string representation of State
func (s State) String() string {
	switch s {
	case StateAwaitingTurn:
		return "Awaiting Turn"
	case StateDraw:
		return "Draw"
	case StateWon:
		return "Won"
	default:
		return fmt.Sprintf("Unknown State (%d)", s)
	}
}

// States returns valid State values
func States() []State {
	return []State{StateAwaitingTurn, StateDraw, StateWon}
}
