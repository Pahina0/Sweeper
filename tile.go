package main

import (
	"strconv"
)

const (
	mine  = -1
	empty = 0

	closed  = 0
	opened  = 1
	flagged = 2
)

type tile struct {
	value int
	state int

	row    int
	column int
}

func (t *tile) IsEmpty() bool {
	return t.value == empty
}

func (t *tile) IsMine() bool {
	return t.value == mine
}

func (t *tile) IsOpened() bool {
	return t.state == opened
}

func (t *tile) IsClosed() bool {
	return t.state == closed
}

func (t *tile) IsFlagged() bool {
	return t.state == flagged
}

// opens a mine, returns success, lost
func (t *tile) open() (bool, bool) {
	if t.IsFlagged() {
		return false, false
	}

	if t.IsOpened() {
		return false, false
	}

	t.state = opened

	return true, t.IsMine()
}

func (t *tile) Initialize(r int, c int, isMine bool) {
	t.state = closed
	t.row = r
	t.column = c

	if isMine {
		t.value = mine
	} else {
		t.value = empty
	}
}

func (t *tile) IncrementVal() {
	t.value++
}

func (t *tile) StringRepresentation(showAll bool) string {

	// return strconv.Itoa(t.row)

	if t.IsFlagged() {
		if showAll && !t.IsMine() {
			return "x"
		}
		return "f"
	}

	if t.IsClosed() && !showAll {
		return " "
	}

	switch t.value {
	case mine:
		return "*"
	case empty:
		return " "
	default:
		return strconv.Itoa(t.value)
	}
}
