/*
Copyright © 2020 Anand Babu Periasamy https://twitter.com/abperiasamy

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/abperiasamy/chess"
)

// Lookup tag pair
func GetTagPair(game *chess.Game, key string) string {
	if game != nil { // Success
		tagPair := game.GetTagPair(key)
		if tagPair != nil {
			return tagPair.Value
		}
	}
	return ""
}

// Start a game from a PGN file
func loadPGN(filename string) *chess.Game {
	pgnDat, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(gConsole.Bold("Unable to read " + gConsole.Red(filename).String() + "."))
		return nil
	}

	pgn, err := chess.PGN(strings.NewReader(string(pgnDat)))
	if err != nil {
		fmt.Println(gConsole.Bold(gConsole.Red(filename)), "is not a valid PGN file.")
		return nil
	}

	game := chess.NewGame(pgn)
	if game == nil {
		fmt.Println("Unable to initialize a new game from " + gConsole.Bold(gConsole.Red(filename)).String() + ".")
		return nil
	}

	tpAnnotator := GetTagPair(game, "Annotator")
	tpWhite := GetTagPair(game, "White")
	tpBlack := GetTagPair(game, "Black")

	// Only load Pinata generated PGN files.
	if tpAnnotator != "pinata" || tpWhite == "" || tpBlack == "" {
		fmt.Println(gConsole.Bold(gConsole.Red(filename)), "is not generated by Pinata.")
		return nil
	}

	// Load previous settings.
	if tpBlack == "Human" { // Human is black.
		gHumanIsBlack = true
		gEngineBinary = tpWhite // Use the same engine as before.
		fmt.Println("You are playing " + gConsole.Bold(gConsole.Yellow("Black")).String() +
			" against " + gConsole.Bold(gConsole.Yellow(gEngineBinary)).String() + ".")
	} else { // Human is white
		gHumanIsBlack = false
		gEngineBinary = tpBlack // Use the same engine as before.
		fmt.Println("You are playing " + gConsole.Bold(gConsole.Yellow("White")).String() +
			" against " + gConsole.Bold(gConsole.Yellow(gEngineBinary)).String() + ".")
	}

	return game
}

// Save the game to a PGN file
func savePGN(game *chess.Game, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Unable to create", gConsole.Bold(gConsole.Red(filename)))
		return err
	}
	defer file.Close()

	// Generate PGN content.
	game.AddTagPair("Annotator", "pinata")
	curTime := time.Now()
	curDate := fmt.Sprintf("%d-%02d-%02d", curTime.Year(), curTime.Month(), curTime.Day())
	game.AddTagPair("Date", curDate)
	game.AddTagPair("Result", game.Outcome().String())
	if humanColor() == chess.White {
		game.AddTagPair("White", "Human")
		game.AddTagPair("Black", gEngineBinary)
	} else {
		game.AddTagPair("White", gEngineBinary)
		game.AddTagPair("Black", "Human")
	}

	// Save the engine name.
	_, err = file.WriteString(game.String() + "\n")
	if err != nil {
		fmt.Println("Unable to save the game to", gConsole.Bold(gConsole.Red(filename)))
		return err
	}

	return nil // Success
}

func drawBoard(game *chess.Game) {
	if !gVisual {
		return // playing blind
	}

	if gHumanIsBlack { // Rotate the board, black facing the human.
		fmt.Print(game.Position().Board().DrawForBlack())
	} else {
		fmt.Print(game.Position().Board().Draw())
	}
}

func isGameOver(game *chess.Game) bool {
	switch game.Outcome() {
	case chess.NoOutcome:
		return false
	case chess.Draw:
		fmt.Println(gConsole.Bold(gConsole.Yellow("Game Draw")).String() +
			" (" + gConsole.Bold(game.Method().String()).String() + ")")
	case chess.WhiteWon:
		fmt.Println(gConsole.Bold(gConsole.Yellow("White Won")).String() +
			" (" + gConsole.Bold(game.Method().String()).String() + ")")
	case chess.BlackWon:
		fmt.Println(gConsole.Bold(gConsole.Yellow("Black Won")).String() +
			" (" + gConsole.Bold(game.Method().String()).String() + ")")
	default:
		panic(game.Outcome()) // should never happen.
	}
	return true // The end.
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func humanColor() chess.Color {
	if gHumanIsBlack {
		return chess.Black
	}
	return chess.White
}

// Readline completion of all the valid moves left.
func validMovesConstructor() func(string) []string {
	return func(string) (moves []string) {
		for _, move := range gGame.Position().ValidMoves() {
			moveSAN := chess.Encoder.Encode(chess.AlgebraicNotation{}, gGame.Position(), move)
			moves = append(moves, moveSAN)
		}
		return moves
	}
}

// Readline completion of all the valid moves left.
func validMoves(game *chess.Game) (moves string) {
	for _, move := range game.Position().ValidMoves() {
		moves += " " + chess.Encoder.Encode(chess.AlgebraicNotation{}, game.Position(), move)
	}
	return moves
}
