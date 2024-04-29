package process

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

type ANSITable struct {
	Sub   *ANSITable
	Data  []byte
	Bound [2]int // rune index
}

// split ansi table given index(bounds), left: [:index] right: [index:]
func (a *ANSITable) Split(index int) (*ANSITable, *ANSITable) {
	if index <= a.Bound[0] {
		return nil, a
	}
	if index >= a.Bound[1] {
		return a, nil
	}
	left := a
	_r := *a
	right := &_r
	left.Bound[1] = index
	right.Bound[0] = index
	if left.Sub != nil {
		left.Sub, right.Sub = left.Sub.Split(index)
	}
	return left, right
}

func (a *ANSITable) AddStyle(data []byte, boundLeft int) {
	if a.Bound[1] < boundLeft {
		panic(errors.New("bound right bigger than current table"))
	}
	if a.Bound[0] > boundLeft {
		boundLeft = a.Bound[0]
	}
	if a.Sub == nil {
		a.Sub = &ANSITable{
			Data:  data,
			Bound: [2]int{boundLeft, a.Bound[1]},
		}
	} else {
		if a.Sub.Bound[0] <= boundLeft {
			a.Sub.AddStyle(data, boundLeft)
		} else {
			a.Sub = &ANSITable{
				Data:  data,
				Bound: [2]int{boundLeft, a.Bound[1]},
				Sub:   a.Sub,
			}
		}
	}
}

// implement `BoundsStruct` for search
func (a *ANSITable) getBounds() [2]int {
	return a.Bound
}

type ANSITableList struct {
	L []BoundsStruct
}

var EMPTY_ANSITABLELIST = make([]BoundsStruct, 0)

// get a slice of ansi table, it will find all tables between `startIndex` and `endIndex`
func (a *ANSITableList) GetSlice(startIndex, endIndex int) []BoundsStruct {
	if len(a.L) == 0 {
		return a.L
	}
	var start, end int
	temp := Search(a.L, startIndex)
	// len == 1 means index within a specific table
	// len == 2 means index between two tables, and for `startIndex` we only need the tables after `startIndex`
	// temp[1] == -1 means already at the front of tablelist and no matchs
	if len(temp) == 1 {
		start = temp[0]
	} else if len(temp) == 2 {
		if temp[1] == -1 {
			return EMPTY_ANSITABLELIST
		} else {
			start = temp[1]
		}
	}

	temp = Search(a.L, endIndex)
	// len == 1 means index within a specific table
	// len == 2 means index between two tables, and for `endIndex` we only need the tables before `endIndex`
	// temp[1] == -1 means already at the front of tablelist and no matchs
	if len(temp) == 1 {
		end = temp[0]
	} else if len(temp) == 2 {
		if temp[0] == -1 {
			return EMPTY_ANSITABLELIST
		} else {
			end = temp[0]
		}
	}

	// get slice of tablelist between start and end
	return a.L[start : end+1]
}

func (a *ANSITableList) SetStyle(style []byte, startIndex, endIndex int) {
	if len(a.L) == 0 {
		var t BoundsStruct = &ANSITable{
			Sub:   nil,
			Data:  style,
			Bound: [2]int{startIndex, endIndex},
		}
		a.L = slices.Insert(a.L, 0, t)
		return
	}

	var start, end int
	temp := Search(a.L, startIndex)
	if len(temp) == 1 {
		start = temp[0]
	} else if len(temp) == 2 {
		if temp[1] == -1 {
			var t BoundsStruct = &ANSITable{
				Sub:   nil,
				Data:  style,
				Bound: [2]int{a.L[temp[0]].getBounds()[1], endIndex},
			}
			a.L = slices.Insert(a.L, len(a.L), t)
			return
		} else {
			t := &ANSITable{
				Sub:  nil,
				Data: style,
			}
			t.Bound[1] = a.L[temp[1]].getBounds()[0]
			if temp[0] != -1 {
				t.Bound[0] = a.L[temp[0]].getBounds()[1]
			}
			var tt BoundsStruct = t
			a.L = slices.Insert(a.L, temp[1], tt)
			start = temp[1] + 1
		}
	}

	temp = Search(a.L, endIndex)
	if len(temp) == 1 {
		end = temp[0]
		t := a.L[end].(*ANSITable)
		// if endIndex bigger than bound[1], split
		if endIndex < t.Bound[1] {
			left, right := t.Split(endIndex + 1)
			if right != nil {
				var r BoundsStruct = right
				a.L = slices.Insert(a.L, end+1, r)
				a.L[end] = left
			}
		}
	} else if len(temp) == 2 {
		if temp[0] == -1 {
			var t BoundsStruct = &ANSITable{
				Sub:   nil,
				Data:  style,
				Bound: [2]int{startIndex, a.L[temp[1]].getBounds()[0]},
			}
			a.L = slices.Insert(a.L, 0, t)
			return
		} else {
			t := &ANSITable{
				Sub:  nil,
				Data: style,
			}
			var tt BoundsStruct = t
			t.Bound[0] = a.L[temp[0]].getBounds()[1]
			if temp[1] != -1 {
				t.Bound[1] = a.L[temp[1]].getBounds()[0]
				a.L = slices.Insert(a.L, temp[1], tt)
			} else {
				t.Bound[1] = endIndex
				a.L = slices.Insert(a.L, len(a.L), tt)
			}
			end = temp[0]
		}
	}

	if start == end {
		a.L[start].(*ANSITable).AddStyle(style, startIndex)
	} else if start < end {
		length := end + 1 - start
		index := start
		var last *ANSITable
		for i := 0; i < length; i++ {
			at := a.L[index].(*ANSITable)
			at.AddStyle(style, startIndex)
			if last != nil {
				lastEnd := last.Bound[1]
				thisStart := at.Bound[0]
				if lastEnd < thisStart {
					var t BoundsStruct = &ANSITable{
						Sub:   nil,
						Data:  style,
						Bound: [2]int{lastEnd, thisStart},
					}
					a.L = slices.Insert(a.L, index, t)
					index++
				}
			}
			last = at
			index++
		}
	}
}

type ANSIQueueItem struct {
	data       []byte
	startIndex int
}

// NOTE: Planning to make these rune process function available to be set from outside
const TAB_RUNE = '\t'

var TAB_BYTES = []byte{32, 32, 32, 32}

func processRune(r rune, writer *strings.Builder) int {
	if r == TAB_RUNE {
		writer.Write(TAB_BYTES)
		return len(TAB_BYTES)
	} else {
		writer.WriteRune(r)
		return 1
	}
}

// transform queue into ansi table which contains all ansi sequences from start to end
func queueToTable(queue []*ANSIQueueItem, endIndex int) *ANSITable {
	first := queue[0]
	root := &ANSITable{
		Bound: [2]int{
			first.startIndex,
			endIndex,
		},
		Data: first.data,
	}

	// add to sub
	temp := root
	for _, v := range queue[1:] {
		temp.Sub = &ANSITable{
			Bound: [2]int{
				v.startIndex,
				endIndex,
			},
			Data: v.data,
		}
		temp = temp.Sub
	}
	return root
}

// split `string with ansi` into `ansi sequences` and `raw string`, Ps: tab will be replaced by 4 spaces
func Extract(s string) (*ANSITableList, string) {
	// preserve normal string
	var normalString strings.Builder
	normalString.Grow(len(s))

	// preserve ansi string and position
	tables := make([]BoundsStruct, 0)
	ansiQueue := make([]*ANSIQueueItem, 0)
	ansi := false
	// NOTE: do not use `for i := range string` index since it's not i+=1 but i+=byte_len
	// solution: transform s into []rune or use custom variable for index
	i := 0
	var ansiItem *ANSIQueueItem = nil
	for _, v := range s {
		// meet `esc` char
		if v == ESCAPE_SEQUENCE {
			// enable ansi mode until meet 'm'
			ansi = true
			// NOTE: using utf8 rune function
			// but maybe just byte(v) is enough since ansi only contains rune of one byte?
			byteData := []byte{}
			byteData = utf8.AppendRune(byteData, v)
			ansiItem = &ANSIQueueItem{
				startIndex: i,
				data:       slices.Clip(byteData),
			}
		} else {
			// in ansi sequence content mode
			if ansi {
				ansiItem.data = utf8.AppendRune(ansiItem.data, v)
				// end of an ansi sequence. terminate
				if IsEscEnd(v) {
					ansi = false
					// clip cap
					ansiItem.data = slices.Clip(ansiItem.data)
					// filter SGR(function named `m`) and push into queue
					if IsSGR(ansiItem.data) {
						ansiQueue = append(ansiQueue, ansiItem)
						// ends all ansi SGR sequences in queue and create ansi table
						if IsEndOfSGR(ansiItem.data) {
							// skip if ansi queue only contains "[0m", which means no SGR actually working
							if len(ansiQueue) > 1 {
								table := queueToTable(ansiQueue[:len(ansiQueue)-1], i)
								tables = append(tables, table)
							}
							// reset queue
							ansiQueue = make([]*ANSIQueueItem, 0)
						}
					}
					// reset item
					ansiItem = nil
				}
			} else {
				// normal content
				i += processRune(v, &normalString)
			}
		}
	}
	return &ANSITableList{
		L: slices.Clip(tables),
	}, normalString.String()
}

func Render(atl *ANSITableList, _s string, startIndex int) []byte {
	s := []rune(_s)
	at := atl.GetSlice(startIndex, startIndex+len(s))
	if len(at) == 0 {
		return []byte(_s)
	}

	var buf bytes.Buffer
	index := 0
	// every table
	for _, a := range at {
		// table's sub tables
		temp := a.(*ANSITable)
		endIndex := temp.Bound[1] - startIndex
		for temp != nil {
			startIndex := temp.Bound[0] - startIndex
			// before table startIndex
			if startIndex > index {
				subRuneDatas := SliceFrom(s, index, startIndex)
				// subRuneDatas := lineRunes[index:startIndex]
				for _, runeData := range subRuneDatas {
					buf.WriteRune(runeData)
				}
				index += len(subRuneDatas)
			}
			// ansi insert
			buf.Write(temp.Data)
			// assign sub table
			temp = temp.Sub
		}
		// add rest
		subRuneDatas := SliceFrom(s, index, endIndex)
		for _, runeData := range subRuneDatas {
			buf.WriteRune(runeData)
		}
		index += len(subRuneDatas)
		// add end escape
		buf.WriteString(ESCAPE_SEQUENCE_END)
	}
	// add rest
	if index <= len(s)-1 {
		subRuneDatas := s[index:]
		for _, runeData := range subRuneDatas {
			buf.WriteRune(runeData)
		}
	}
	return buf.Bytes()
}
