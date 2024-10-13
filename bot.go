package tictactoe

import "math"

// Bot represents a machine-controlled player of a Game whose sole purpose is to beat a human Player
type Bot interface {
	// Player returns the Player for which the Bot is playing
	Player() Player
	// Turn allows the Bot to check the Board for the best possible turn and returns the Cell representing that turns
	// location on Board.
	//
	// The Board provided is not a copy so a Bot must never mutate it or risk corrupting the Game.
	//
	// Turn is only ever called if there is at least one possible turn to make so a Bot is always able to return
	// something. However, an error may be returned in some cases (e.g. ErrConditionInvalid).
	Turn(board Board, game Game) (Cell, error)
}

type easyBot struct {
	player Player
}

func (b *easyBot) Player() Player {
	return b.player
}

func (b *easyBot) Turn(board Board, _ Game) (Cell, error) {
	return board.FindEmpty().Random(), nil
}

// NewEasyBot returns a new Bot with a very easy difficulty
func NewEasyBot(player Player) Bot {
	return &easyBot{player}
}

type normalBot struct {
	player Player
}

func (b *normalBot) Player() Player {
	return b.player
}

func (b *normalBot) Turn(board Board, game Game) (Cell, error) {
	candidates := board.FindEmpty()
	conditions := game.Conditions()
	for _, candidate := range candidates {
		nextBoard := board.Copy()
		nextBoard[candidate.Row][candidate.Column] = b.player

		turn := Turn{
			Cell:   candidate,
			Player: b.player,
		}
		if conditions.IsWinningTurn(nextBoard, turn) {
			return candidate, nil
		}
	}
	return candidates.Random(), nil
}

// NewNormalBot returns a new Bot with a normal difficulty
func NewNormalBot(player Player) Bot {
	return &normalBot{player}
}

type hardBot struct {
	opponent Player
	player   Player
}

func (b *hardBot) Player() Player {
	return b.player
}

func (b *hardBot) Turn(board Board, game Game) (Cell, error) {
	candidates := board.FindEmpty()
	conditions := game.Conditions()
	var oppWinner *Cell
	for _, candidate := range candidates {
		nextBoard := board.Copy()
		nextBoard[candidate.Row][candidate.Column] = b.player

		turn := Turn{
			Cell:   candidate,
			Player: b.player,
		}
		if conditions.IsWinningTurn(nextBoard, turn) {
			return candidate, nil
		}

		oppCandidates := nextBoard.FindEmpty()
		for _, oppCandidate := range oppCandidates {
			oppBoard := nextBoard.Copy()
			oppBoard[oppCandidate.Row][oppCandidate.Column] = b.opponent

			turn = Turn{
				Cell:   oppCandidate,
				Player: b.opponent,
			}
			if conditions.IsWinningTurn(oppBoard, turn) {
				copyCandidate := oppCandidate
				oppWinner = &copyCandidate
			}
		}
	}

	if oppWinner != nil {
		return *oppWinner, nil
	}
	return candidates.Random(), nil
}

// NewHardBot returns a new Bot with a hard difficulty
func NewHardBot(player Player) Bot {
	return &hardBot{
		opponent: player.Next(),
		player:   player,
	}
}

type (
	impossibleBot struct {
		player Player
	}

	impossibleChoice struct {
		cell         Cell
		depth, value int
	}
)

func (b *impossibleBot) Player() Player {
	return b.player
}

func (b *impossibleBot) Turn(board Board, game Game) (Cell, error) {
	lastTurn, _ := game.LastTurn()
	choice, err := b.minimax(board, game, lastTurn, true, b.player, 0)
	return choice.cell, err
}

func (b *impossibleBot) minimax(board Board, game Game, lastTurn Turn, max bool, player Player, depth int) (impossibleChoice, error) {
	candidates := board.FindEmpty()
	conditions := game.Conditions()
	if winner, err := conditions.FindWinner(board); err != nil {
		return impossibleChoice{}, err
	} else if winner == b.player {
		return impossibleChoice{
			cell:  lastTurn.Cell,
			depth: depth,
			value: (game.MaxTurns() + 1) - depth,
		}, nil
	} else if winner == b.player.Next() {
		return impossibleChoice{
			cell:  lastTurn.Cell,
			depth: depth,
			value: -(game.MaxTurns() + 1) + depth,
		}, nil
	} else if len(candidates) == 0 {
		return impossibleChoice{
			cell:  lastTurn.Cell,
			depth: depth,
			value: 0,
		}, nil
	}

	var choices []impossibleChoice
	for _, candidate := range candidates {
		nextBoard := board.Copy()
		nextBoard[candidate.Row][candidate.Column] = player

		turn := Turn{
			Cell:   candidate,
			Player: player,
		}
		choice, err := b.minimax(nextBoard, game, turn, !max, player.Next(), depth+1)
		if err != nil {
			return choice, err
		}
		choice.cell = turn.Cell
		choices = append(choices, choice)
	}

	minChoice := impossibleChoice{
		depth: depth,
		value: int(math.Pow(float64(game.MaxTurns()+1), 2)),
	}
	maxChoice := impossibleChoice{
		depth: depth,
		value: -int(math.Pow(float64(game.MaxTurns()+1), 2)),
	}

	for _, choice := range choices {
		switch {
		case max && choice.value > maxChoice.value:
			maxChoice = choice
		case !max && choice.value < minChoice.value:
			minChoice = choice
		}
	}

	if max {
		return maxChoice, nil
	}
	return minChoice, nil
}

// NewImpossibleBot returns a new Bot with an impossible-to-beat difficulty
func NewImpossibleBot(player Player) Bot {
	return &impossibleBot{player}
}
