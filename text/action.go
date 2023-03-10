package text

import (
	world "github.com/sapphire-ai-dev/sapphire-world"
)

func (w *textWorld) validCursorItem(currDir *directory, pos *actorPos, cmd int) bool {
	dirSize := len(currDir.content)
	if currDir.parent() != nil {
		dirSize++
	}

	if pos.cursorItem < 0 || pos.cursorItem >= dirSize {
		return false
	}

	if pos.cursorItem == 0 && cmd == changeItemCmdUp {
		return false
	}

	if pos.cursorItem == len(currDir.content)-1 && cmd == changeItemCmdDown {
		return false
	}

	return true
}

func (w *textWorld) locateItem(actorId int, cmd int) (*actorPos, item) {
	pos, posSeen := w.actors[actorId]
	if !posSeen {
		return nil, nil
	}

	currItem, currItemSeen := w.items[pos.currItemId]
	if !currItemSeen {
		return nil, nil
	}

	if _, ok := currItem.(*file); ok {
		if pos.cursorItem != 0 {
			return nil, nil
		}

		if cmd != changeItemCmdEnter {
			return nil, nil
		}
	}

	return pos, currItem
}

func (w *textWorld) changeItemReady(actorId, cmd int) bool {
	pos, currItem := w.locateItem(actorId, cmd)
	if pos == nil {
		return false
	}

	if _, ok := currItem.(*file); ok {
		return true
	}

	currDir := currItem.(*directory)
	return w.validCursorItem(currDir, pos, cmd)
}

func (w *textWorld) changeItemStep(actorId, cmd int) {
	pos, currItem := w.locateItem(actorId, cmd)
	if pos == nil {
		return
	}

	if _, ok := currItem.(*file); ok {
		w.actors[actorId].cursorItem = currItem.parent().id()
		return
	}

	currDir := currItem.(*directory)
	if !w.validCursorItem(currDir, pos, cmd) {
		return
	}

	if cmd == changeItemCmdUp {
		w.actors[actorId].cursorItem--
		return
	}

	if cmd == changeItemCmdDown {
		w.actors[actorId].cursorItem++
		return
	}

	cursorItem := pos.cursorItem
	if currDir.parent() != nil {
		cursorItem--
	}

	if cursorItem == -1 {
		w.actors[actorId].currItemId = currDir.parent().id()
	} else {
		w.actors[actorId].currItemId = currDir.content[cursorItem].id()
	}

	w.actors[actorId].cursorItem = 0
}

func (w *textWorld) changeItemWrap(actorId, cmd int) *world.ActionInterface {
	if cmd < 0 || cmd >= changeItemCmdEnd {
		return nil
	}

	return &world.ActionInterface{
		Name: pressKeyCmds[cmd],
		Ready: func() bool {
			return w.changeItemReady(actorId, cmd)
		},
		Step: func() {
			w.changeItemStep(actorId, cmd)
		},
	}
}

func (w *textWorld) identifyFile(actorId int) (*file, *actorPos) {
	pos, posSeen := w.actors[actorId]
	if !posSeen {
		return nil, nil
	}

	currItem, currItemSeen := w.items[pos.currItemId]
	if !currItemSeen {
		return nil, nil
	}

	currFile, isFile := currItem.(*file)
	if !isFile {
		return nil, nil
	}

	if pos.cursorLine < 0 || pos.cursorLine >= len(currFile.lines) {
		return nil, nil
	}

	currLine := currFile.lines[pos.cursorLine]
	if pos.cursorChar < 0 || pos.cursorChar > len(currLine.characters) {
		return nil, nil
	}

	return currFile, pos
}

func (w *textWorld) pressKeyReady(actorId int) bool {
	currFile, _ := w.identifyFile(actorId)
	return currFile != nil
}

func (w *textWorld) pressKeyStep(actorId, cmd int) {
	currFile, pos := w.identifyFile(actorId)
	if currFile == nil {
		return
	}

	if val, seen := pressKeyCmds[cmd]; seen {
		currLine := currFile.lines[pos.cursorLine]
		left, right := currLine.characters[:pos.cursorChar], currLine.characters[pos.cursorChar:]
		currLine.characters = append(append(left, currLine.newCharacter(val)), right...)
		w.actors[actorId].cursorChar++
	}
}

func (w *textWorld) pressKeyWrap(actorId, cmd int) *world.ActionInterface {
	if cmd < 0 || cmd >= pressKeyCmdEnd {
		return nil
	}

	return &world.ActionInterface{
		Name: "key" + pressKeyCmds[cmd],
		Ready: func() bool {
			return w.pressKeyReady(actorId)
		},
		Step: func() {
			w.pressKeyStep(actorId, cmd)
		},
	}
}

func (w *textWorld) validCursor(actorId int, cmd int) bool {
	currFile, pos := w.identifyFile(actorId)
	if currFile == nil {
		return false
	}

	if pos.cursorChar == 0 && (cmd == pressKeyCmdLeft || cmd == pressKeyCmdBackspace) {
		return false
	}

	if pos.cursorLine == 0 && cmd == pressKeyCmdUp {
		return false
	}

	currLine := currFile.lines[pos.cursorLine]
	if pos.cursorLine == len(currFile.lines)-1 && cmd == pressKeyCmdDown {
		return false
	}

	if pos.cursorChar == len(currLine.characters) && cmd == pressKeyCmdRight {
		return false
	}

	return true
}

func (w *textWorld) specialKeyReady(actorId int, cmd int) bool {
	return w.validCursor(actorId, cmd)
}

func (w *textWorld) specialKeyStep(actorId int, cmd int) {
	currFile, pos := w.identifyFile(actorId)
	if currFile == nil {
		return
	}

	if !w.validCursor(actorId, cmd) {
		return
	}

	currLine := currFile.lines[pos.cursorLine]

	switch cmd {
	case pressKeyCmdBackspace:
		w.actors[actorId].cursorChar--

		left, right := currLine.characters[:pos.cursorChar-1], currLine.characters[pos.cursorChar:]
		currLine.characters = append(left, right...)
	case pressKeyCmdEnter:
		currLine.characters = currLine.characters[:pos.cursorChar]
		newLine := currFile.newLine()
		newLine.characters = currLine.characters[pos.cursorChar:]

		up, down := currFile.lines[:pos.cursorLine], currFile.lines[pos.cursorLine+1:]
		up = append(up, currLine)
		currFile.lines = append(append(up, newLine), down...)
		w.actors[actorId].cursorLine++
		w.actors[actorId].cursorChar = 0
	case pressKeyCmdUp:
		w.actors[actorId].cursorLine--
	case pressKeyCmdDown:
		w.actors[actorId].cursorLine++
	case pressKeyCmdLeft:
		w.actors[actorId].cursorChar--
	default:
		// pressKeyCmdRight
		w.actors[actorId].cursorChar++
	}
}

func (w *textWorld) specialKeyWrap(actorId, cmd int) *world.ActionInterface {
	if specialKeyCmds[cmd] != true {
		return nil
	}

	return &world.ActionInterface{
		Name: "key" + pressKeyCmds[cmd],
		Ready: func() bool {
			return w.specialKeyReady(actorId, cmd)
		},
		Step: func() {
			w.specialKeyStep(actorId, cmd)
		},
	}
}

func (w *textWorld) newActionInterfaces(actorId int) []*world.ActionInterface {
	var result []*world.ActionInterface
	for cmd := range changeItemCmds {
		result = append(result, w.changeItemWrap(actorId, cmd))
	}

	for cmd := range pressKeyCmds {
		result = append(result, w.pressKeyWrap(actorId, cmd))
	}

	for cmd := range specialKeyCmds {
		result = append(result, w.specialKeyWrap(actorId, cmd))
	}

	return result
}

const (
	changeItemCmdUp = iota
	changeItemCmdDown
	changeItemCmdEnter
	changeItemCmdExec
	changeItemCmdEnd
)

var changeItemCmds = map[int]string{
	changeItemCmdUp:    "itemUp",
	changeItemCmdDown:  "itemDown",
	changeItemCmdEnter: "itemEnter",
	changeItemCmdExec:  "itemExec",
}

const (
	pressKeyCmd0 = iota
	pressKeyCmd1
	pressKeyCmd2
	pressKeyCmd3
	pressKeyCmd4
	pressKeyCmd5
	pressKeyCmd6
	pressKeyCmd7
	pressKeyCmd8
	pressKeyCmd9
	pressKeyCmdA
	pressKeyCmdB
	pressKeyCmdC
	pressKeyCmdD
	pressKeyCmdE
	pressKeyCmdF
	pressKeyCmdG
	pressKeyCmdH
	pressKeyCmdI
	pressKeyCmdJ
	pressKeyCmdK
	pressKeyCmdL
	pressKeyCmdM
	pressKeyCmdN
	pressKeyCmdO
	pressKeyCmdP
	pressKeyCmdQ
	pressKeyCmdR
	pressKeyCmdS
	pressKeyCmdT
	pressKeyCmdU
	pressKeyCmdV
	pressKeyCmdW
	pressKeyCmdX
	pressKeyCmdY
	pressKeyCmdZ
	pressKeyCmdShift0
	pressKeyCmdShift1
	pressKeyCmdShift2
	pressKeyCmdShift3
	pressKeyCmdShift4
	pressKeyCmdShift5
	pressKeyCmdShift6
	pressKeyCmdShift7
	pressKeyCmdShift8
	pressKeyCmdShift9
	pressKeyCmdMinus
	pressKeyCmdPlus
	pressKeyCmdUnderscore
	pressKeyCmdEqual
	pressKeyCmdLeftSquareBracket
	pressKeyCmdLeftCurlyBracket
	pressKeyCmdRightSquareBracket
	pressKeyCmdRightCurlyBracket
	pressKeyCmdSpace
	pressKeyCmdComma
	pressKeyCmdPeriod
	pressKeyCmdSlash
	pressKeyCmdShiftComma
	pressKeyCmdShiftPeriod
	pressKeyCmdShiftSlash
	pressKeyCmdBackSlash
	pressKeyCmdVertical
	pressKeyCmdBackspace
	pressKeyCmdEnter
	pressKeyCmdUp
	pressKeyCmdDown
	pressKeyCmdLeft
	pressKeyCmdRight
	pressKeyCmdEnd
)

var pressKeyCmds = map[int]string{
	pressKeyCmd1:                  "1",
	pressKeyCmd2:                  "2",
	pressKeyCmd3:                  "3",
	pressKeyCmd4:                  "4",
	pressKeyCmd5:                  "5",
	pressKeyCmd6:                  "6",
	pressKeyCmd7:                  "7",
	pressKeyCmd8:                  "8",
	pressKeyCmd9:                  "9",
	pressKeyCmdA:                  "a",
	pressKeyCmdB:                  "b",
	pressKeyCmdC:                  "c",
	pressKeyCmdD:                  "d",
	pressKeyCmd0:                  "0",
	pressKeyCmdE:                  "e",
	pressKeyCmdF:                  "f",
	pressKeyCmdG:                  "g",
	pressKeyCmdH:                  "h",
	pressKeyCmdI:                  "i",
	pressKeyCmdJ:                  "j",
	pressKeyCmdK:                  "k",
	pressKeyCmdL:                  "l",
	pressKeyCmdM:                  "m",
	pressKeyCmdN:                  "n",
	pressKeyCmdO:                  "o",
	pressKeyCmdP:                  "p",
	pressKeyCmdQ:                  "q",
	pressKeyCmdR:                  "r",
	pressKeyCmdS:                  "s",
	pressKeyCmdT:                  "t",
	pressKeyCmdU:                  "u",
	pressKeyCmdV:                  "v",
	pressKeyCmdW:                  "w",
	pressKeyCmdX:                  "x",
	pressKeyCmdY:                  "y",
	pressKeyCmdZ:                  "z",
	pressKeyCmdShift0:             "!",
	pressKeyCmdShift1:             "@",
	pressKeyCmdShift2:             "#",
	pressKeyCmdShift3:             "$",
	pressKeyCmdShift4:             "%",
	pressKeyCmdShift5:             "^",
	pressKeyCmdShift6:             "&",
	pressKeyCmdShift7:             "*",
	pressKeyCmdShift8:             "(",
	pressKeyCmdShift9:             ")",
	pressKeyCmdMinus:              "-",
	pressKeyCmdPlus:               "+",
	pressKeyCmdUnderscore:         "_",
	pressKeyCmdEqual:              "=",
	pressKeyCmdLeftSquareBracket:  "[",
	pressKeyCmdLeftCurlyBracket:   "{",
	pressKeyCmdRightSquareBracket: "]",
	pressKeyCmdRightCurlyBracket:  "}",
	pressKeyCmdSpace:              " ",
	pressKeyCmdComma:              ",",
	pressKeyCmdPeriod:             ".",
	pressKeyCmdSlash:              "/",
	pressKeyCmdShiftComma:         "<",
	pressKeyCmdShiftPeriod:        ">",
	pressKeyCmdShiftSlash:         "?",
	pressKeyCmdBackSlash:          "\\",
	pressKeyCmdVertical:           "|",
}

var specialKeyCmds = map[int]bool{
	pressKeyCmdBackspace: true,
	pressKeyCmdEnter:     true,
	pressKeyCmdUp:        true,
	pressKeyCmdDown:      true,
	pressKeyCmdLeft:      true,
	pressKeyCmdRight:     true,
}
