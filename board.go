package main

import (
	"math/rand/v2"
	"strings"
)

type board struct {
	rowCount    int
	columnCount int

	mineCount int

	mines [][]tile
}

func generateBoard(rows int, cols int, mineCount int) board {
	mines := make([][]tile, rows)
	for i := range mines {
		mines[i] = make([]tile, cols)
	}

	for i := 0; i < mineCount; i++ {
		mines[i/cols][i%cols].Initialize(i/cols, i%cols, true)
	}

	for i := mineCount; i < rows*cols; i++ {
		mines[i/cols][i%cols].Initialize(i/cols, i%cols, false)
	}

	rand.Shuffle(rows*cols, func(i, j int) {
		mines[i/cols][i%cols].value, mines[j/cols][j%cols].value =
			mines[j/cols][j%cols].value, mines[i/cols][i%cols].value
	})

	return board{
		rowCount:    rows,
		columnCount: cols,

		mineCount: mineCount,

		mines: mines,
	}
}

func (b *board) updateStartBoard(row int, col int) {
	if !b.mines[row][col].IsMine() {
		return
	}

	// makes it go to the nth blank space available
	offset := rand.IntN(b.rowCount*b.columnCount-b.mineCount) + 1

	start := row*b.columnCount + col
	for i := 1; i <= b.rowCount*b.columnCount; i++ {
		newRow := (start + i) % (b.rowCount * b.columnCount) / b.columnCount
		newColumn := ((start + i) % (b.rowCount * b.columnCount)) % b.columnCount

		if !b.mines[newRow][newColumn].IsMine() {
			offset--

			if offset > 0 {
				continue
			}

			b.mines[start/b.columnCount][start%b.columnCount].value,
				b.mines[newRow][newColumn].value =
				b.mines[newRow][newColumn].value,
				b.mines[start/b.columnCount][start%b.columnCount].value

			break
		}
	}
}

func (b *board) calculateSurrounding() {
	for i, row := range b.mines {
		for j, col := range row {

			if col.IsMine() {
				continue
			}

			for k := -1; k <= 1; k++ {
				for l := -1; l <= 1; l++ {
					if i+k < 0 || i+k >= b.rowCount || l+j < 0 || l+j >= b.columnCount {
						continue
					}

					if b.mines[i+k][l+j].IsMine() {
						b.mines[i][j].IncrementVal()
					}
				}
			}
		}
	}
}

func (g *game) checkRenderColor(toRender string, tiles ...*tile) string {
	if g.lost {
		return styleLost.Render(toRender)
	}

	if g.won {
		return styleWon.Render(toRender)
	}

	for _, tile := range tiles {
		if g.selectedCol == tile.column && g.selectedRow == tile.row {
			return styleSelected.Render(toRender)
		}
	}

	for _, tile := range tiles {
		if tile.IsClosed() {
			return styleUnopened.Render(toRender)
		}
	}

	for _, tile := range tiles {
		if tile.IsFlagged() {
			return styleFlagged.Render(toRender)
		}
	}

	return toRender
}

func (g *game) drawRow(
	s *strings.Builder,
	rowAware bool,
	row int,
	left string,
	middle string,
	right string,
	spacing func(int) string) {
	canCheckNextRow := rowAware && row+1 < g.board.rowCount
	var canCheckNextColumn bool

	// left side
	toCheck := []*tile{&g.board.mines[row][0]}

	if canCheckNextRow {
		toCheck = append(toCheck, &g.board.mines[row+1][0])
	}
	s.WriteString(g.checkRenderColor(left, toCheck...))

	for i := 0; i < g.board.columnCount-1; i++ {

		canCheckNextColumn = i+1 < g.board.columnCount

		// starts adding the spacing
		toCheck = []*tile{&g.board.mines[row][i]}

		if canCheckNextRow {
			toCheck = append(toCheck, &g.board.mines[row+1][i])
		}

		s.WriteString(g.checkRenderColor(spacing(i), toCheck...))

		// starts adding intersection
		toCheck = []*tile{&g.board.mines[row][i]}

		if canCheckNextColumn {
			toCheck = append(toCheck, &g.board.mines[row][i+1])
			if canCheckNextRow {
				toCheck = append(toCheck, &g.board.mines[row+1][i+1])
			}
		}

		if canCheckNextRow {
			toCheck = append(toCheck, &g.board.mines[row+1][i])
		}

		s.WriteString(g.checkRenderColor(middle, toCheck...))
		// s.WriteString(middle)

	}

	// renders right side
	toCheck = []*tile{&g.board.mines[row][g.board.columnCount-1]}

	if canCheckNextRow {
		toCheck = append(toCheck, &g.board.mines[row+1][g.board.columnCount-1])
	}
	s.WriteString(g.checkRenderColor(spacing(g.board.columnCount-1)+right, toCheck...))
	s.WriteString("\n")
}

func (g *game) drawTop(s *strings.Builder) {
	g.drawRow(s, false, 0, "┌", "┬", "┐", func(i int) string { return "───" })
}

func (g *game) drawMiddle(s *strings.Builder, row int) {
	g.drawRow(s, true, row, "├", "┼", "┤", func(_ int) string { return "───" })
}

func (g *game) drawBottom(s *strings.Builder) {
	g.drawRow(s, false, g.board.columnCount-1, "└", "┴", "┘", func(_ int) string { return "───" })
}

func (g *game) drawItems(s *strings.Builder, row int, showAll bool) {
	g.drawRow(s, false, row, "│", "│", "│", func(i int) string {

		// return strconv.Itoa(row) + "," + strconv.Itoa(i)
		return " " + g.board.mines[row][i].StringRepresentation(showAll) + " "
	})

}
