package main

import (
	"github.com/nsf/termbox-go"
)

var DEFAULT_GROUP = "default"

type Layer struct {
	added      int
	Containers []*Container
}

type Container struct {
	X, Y, Width, Height int
	ContainerElements   []*ContainerElement
	currentGroup        string
}

type Drawable interface {
	Draw()
}

type VisibleElement interface {
	getMatrix() []RunePos
}

type ContainerElement struct {
	RuneMatrixPos []RunePos
	Group         string
	Options       int
	Element       VisibleElement
}

type SelectableElement interface {
	getGroup() string
}

type Word struct {
	WordString string
	Width      int
	Fg         termbox.Attribute
	Bg         termbox.Attribute
	WordChan   chan<- string
}

type LineBreakType struct{}

type Table struct {
	Cols      []string
	Rows      []TableRow
	ColWidths []int
	state     []RunePos
	Width     int
}

type RunePos struct {
	X, Y int
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

type TableRow struct {
	Cells    []string
	Selected bool
	Fg       termbox.Attribute
	Bg       termbox.Attribute
}

func NewLayer() *Layer {
	els := make([]*Container, 0)
	return &Layer{
		Containers: els,
	}
}

func newContainerElement(el VisibleElement, group string) *ContainerElement {
	return &ContainerElement{
		Element: el,
		Group:   group,
	}
}

func NewWord(word string, width int, fg termbox.Attribute, bg termbox.Attribute) *Word {
	return &Word{
		WordString: word,
		Width:      width,
		Fg:         fg,
		Bg:         bg,
	}
}

func NewWordDyn(word string, width int, fg termbox.Attribute, bg termbox.Attribute, newword chan<- string) *Word {
	return &Word{
		WordString: word,
		Width:      width,
		Fg:         fg,
		Bg:         bg,
		WordChan:   newword,
	}
}

func NewWordDef(word string, width int) *Word {
	return NewWord(word, width, 0, 0)
}

func LineBreak() *LineBreakType {
	return &LineBreakType{}
}

func NewContainer(x, y, width, height int) *Container {
	return &Container{
		X:                 x,
		Y:                 y,
		Width:             width,
		Height:            height,
		currentGroup:      DEFAULT_GROUP,
		ContainerElements: make([]*ContainerElement, 0),
	}
}

func NewTableRow(fields ...string) *TableRow {
	row := &TableRow{
		Cells: fields,
	}

	return row
}

func (l *Layer) Add(el *Container) {
	l.Containers = append(l.Containers, el)

}
func (c *Container) Add(el VisibleElement) {
	cel := newContainerElement(el, c.currentGroup)
	c.ContainerElements = append(c.ContainerElements, cel)
}

func (c *Container) StartGroup(hash string) {
	c.currentGroup = hash
}

func (c *Container) StopGroup() {
	c.currentGroup = DEFAULT_GROUP
}

func (c *Container) DeleteGroup(hash string) {
	newArray := make([]*ContainerElement, 0)
	for _, v := range c.ContainerElements {
		if v.Group != hash {
			newArray = append(newArray, v)
		}
	}
	c.ContainerElements = newArray
}

func (c *Container) RecalculateRunes() {
	lastRune := NewRunePos(c.X-1, c.Y, ' ', 0, 0) // -1 so we have space between elementsn
	lineBreaksCount := 0
	for _, v := range c.ContainerElements {
		matrix := v.Element.getMatrix()
		if len(matrix) == 0 {
			lineBreaksCount++
			lastRune = NewRunePos(c.X-1, c.Y+lineBreaksCount, ' ', 0, 0)

		} else {
			matrix = addConstant(matrix, lastRune.X+1, lastRune.Y)
			v.RuneMatrixPos = matrix
			lastRune = matrix[len(matrix)-1]
		}
	}
}

func (c *Container) EmptyRunePos() RunePos {
	return NewRunePos(c.X, c.Y, 0, 0, 0)
}

func (c *Container) Draw() {
	c.RecalculateRunes()
	for x := 0; x < c.Width; x++ { // cleanup
		for y := 0; y < c.Height; y++ {
			termbox.SetCell(c.X+x, c.Y+y, ' ', 0, 0)
		}
	}
	for _, v := range c.ContainerElements {
		for _, e := range v.RuneMatrixPos {
			if e.X <= c.Width+c.X && e.Y <= c.Height+c.Y {
				termbox.SetCell(
					e.X,
					e.Y,
					e.Char,
					e.Fg,
					e.Bg,
				)
			}
		}

	}
}

func (l *Layer) Draw() {
	for _, v := range l.Containers {
		v.Draw()
	}
}

func (w *Word) getMatrix() []RunePos {
	matrix := make([]RunePos, w.Width)
	for i := 0; i < w.Width; i++ {
		chru := byte(' ')
		if i < len(w.WordString) {
			chru = w.WordString[i]
		}
		matrix[i] = NewRunePos(i, 0, chru, w.Fg, w.Bg)
	}
	return matrix
}

func (l *LineBreakType) getMatrix() []RunePos {
	runePosMatrix := make([]RunePos, 0)
	return runePosMatrix
}

func NewRunePos(x, y int, ch byte, fg, bg termbox.Attribute) RunePos {
	return RunePos{
		X:    x,
		Y:    y,
		Char: rune(ch),
		Fg:   fg,
		Bg:   bg,
	}
}

func Space() *Word {
	return NewWordDef(" ", 1)
}

type TableCols map[string]int

func (c *Container) NewTable(cols []string, widths []int) *Table {
	return &Table{
		Cols:      cols,
		ColWidths: widths,
	}
}

func (c *Container) NewTableWithHeader(cols []string, widths []int) *Table {
	table := c.NewTable(cols, widths)
	c.AddTableHeader(table)
	return table
}

func (c *Container) AddTableHeader(t *Table) {
	var width int
	for k, v := range t.Cols {
		width = t.ColWidths[k]
		c.Add(NewWordDef(v, width))
		c.Add(Space())

	}
	c.Add(LineBreak())
}

func (c *Container) AddTableRow(t *Table, row *TableRow, hash string) {
	var width int
	c.StartGroup(hash)
	for k, cell := range row.Cells {
		width = t.ColWidths[k]
		c.Add(NewWord(cell, width, row.Fg, row.Bg))

		c.Add(Space())
	}
	c.Add(LineBreak())
	c.StopGroup()
}

func appendRunePosMatrix(m1, m2 []RunePos) []RunePos {
	for _, v := range m2 {
		m1 = append(m1, v)
	}
	return m1
}

func appendRunePos(m []RunePos, p RunePos) []RunePos {
	m = append(m, p)
	return m
}

func addConstant(m []RunePos, x, y int) []RunePos {
	for k, v := range m {
		m[k].X = v.X + x
		m[k].Y = v.Y + y
	}
	return m
}