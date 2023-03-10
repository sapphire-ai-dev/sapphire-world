package text

import (
	"testing"

	world "github.com/sapphire-ai-dev/sapphire-world"
	"github.com/stretchr/testify/assert"
)

func TestChangeItemErrorHandling(t *testing.T) {
	w := newTextWorld()
	assert.Nil(t, w.changeItemWrap(0, changeItemCmdEnd))
	c1 := w.changeItemWrap(0, changeItemCmdUp)

	// actor does not exist
	assert.False(t, c1.Ready())
	c1.Step()

	// cannot go up
	actorId, _ := w.NewActor()
	c2 := w.changeItemWrap(actorId, changeItemCmdUp)
	assert.False(t, c2.Ready())
	c2.Step()

	// actor not on an item
	root := w.rootDirectory
	w.actors[actorId].currItemId = root.id() + 1
	assert.False(t, c2.Ready())
	c2.Step()
}

func assertItemDirections(t *testing.T, w *textWorld, actorId int, itemDirections map[int]string) {
	imgs := w.Look(actorId)
	imgMap := map[int]*world.Image{}
	for _, img := range imgs {
		imgMap[img.Id] = img
	}

	for itemId, directionEnum := range itemDirections {
		assert.Contains(t, imgMap[itemId].Transient[0].Labels, directionEnum)
	}
}

func TestChangeItemUpDown(t *testing.T) {
	w := newTextWorld()
	root := w.rootDirectory
	actorId, _ := w.NewActor()
	f1, f2, f3 := root.newFile("fName1"), root.newFile("fName2"), root.newFile("fName3")
	ciU := w.changeItemWrap(actorId, changeItemCmdUp)
	ciD := w.changeItemWrap(actorId, changeItemCmdDown)

	assertPos0 := func() {
		assert.False(t, ciU.Ready())
		assert.True(t, ciD.Ready())
		assertItemDirections(t, w, actorId, map[int]string{
			f1.id(): world.TernaryZro,
			f2.id(): world.TernaryNeg,
			f3.id(): world.TernaryNeg,
		})
	}

	assertPos1 := func() {
		assert.True(t, ciU.Ready())
		assert.True(t, ciD.Ready())
		assertItemDirections(t, w, actorId, map[int]string{
			f1.id(): world.TernaryPos,
			f2.id(): world.TernaryZro,
			f3.id(): world.TernaryNeg,
		})
	}

	assertPos2 := func() {
		assert.True(t, ciU.Ready())
		assert.False(t, ciD.Ready())
		assertItemDirections(t, w, actorId, map[int]string{
			f1.id(): world.TernaryPos,
			f2.id(): world.TernaryPos,
			f3.id(): world.TernaryZro,
		})
	}

	assertPos0()
	ciU.Step()
	assertPos0()
	ciD.Step()
	assertPos1()
	ciD.Step()
	assertPos2()
	ciD.Step()
	assertPos2()
	ciU.Step()
	assertPos1()
}

func TestChangeItemEnter(t *testing.T) {
	w := newTextWorld()
	actorId, _ := w.NewActor()
	c1 := w.changeItemWrap(actorId, changeItemCmdEnter)
	assert.False(t, c1.Ready())
	c1.Step()

	// enter child directory
	root := w.rootDirectory
	d := root.newDirectory("dName")
	assert.True(t, c1.Ready())
	c1.Step()
	assert.Equal(t, d.id(), w.actors[actorId].currItemId)

	// enter parent directory from directory
	assert.True(t, c1.Ready())
	c1.Step()
	assert.Equal(t, root.id(), w.actors[actorId].currItemId)

	// enter file
	c1.Step()
	assert.Equal(t, d.id(), w.actors[actorId].currItemId)
	f := d.newFile("fName")
	w.actors[actorId].cursorItem++
	assert.True(t, c1.Ready())
	c1.Step()
	assert.Equal(t, f.id(), w.actors[actorId].currItemId)

	// enter parent directory from file
	assert.True(t, c1.Ready())
	c1.Step()
	assert.Equal(t, f.id(), w.actors[actorId].currItemId)

	// enter other position from file should fail
	w.actors[actorId].cursorItem++
	c1.Step()
	assert.Equal(t, f.id(), w.actors[actorId].currItemId)
	w.actors[actorId].cursorItem++
	assert.False(t, c1.Ready())
	c1.Step()
	assert.Equal(t, f.id(), w.actors[actorId].currItemId)

	c2 := w.changeItemWrap(actorId, changeItemCmdUp)
	assert.False(t, c2.Ready())
	c2.Step()

	w.actors[actorId].cursorItem = 0
	assert.False(t, c2.Ready())
	c2.Step()
}

func TestPressKeyErrorHandling(t *testing.T) {
	w := newTextWorld()
	c1 := w.pressKeyWrap(0, pressKeyCmd0)

	// actor does not exist
	assert.False(t, c1.Ready())
	c1.Step()

	// not on a file
	actorId, _ := w.NewActor()
	c2 := w.pressKeyWrap(actorId, pressKeyCmd0)
	assert.False(t, c2.Ready())
	c2.Step()

	// still not on a file
	root := w.rootDirectory
	w.actors[actorId].currItemId = root.id() + 1
	assert.False(t, c2.Ready())
	c2.Step()

	// not on a line
	f := root.newFile("fName")
	w.actors[actorId].currItemId = f.id()
	w.actors[actorId].cursorLine = -1
	assert.False(t, c2.Ready())
	c2.Step()
	w.actors[actorId].cursorLine = 1
	assert.False(t, c2.Ready())
	c2.Step()

	// not on a char
	w.actors[actorId].cursorLine = 0
	w.actors[actorId].cursorChar = -1
	assert.False(t, c2.Ready())
	c2.Step()
	w.actors[actorId].cursorChar = 1
	assert.False(t, c2.Ready())
	c2.Step()

	// invalid key code
	assert.Nil(t, w.pressKeyWrap(actorId, pressKeyCmdEnd))

	// the failed actions had no effects
	assert.Len(t, f.lines, 1)
	assert.Len(t, f.lines[0].characters, 0)
}

func TestPressKey(t *testing.T) {
	w := newTextWorld()
	actorId, _ := w.NewActor()
	root := w.rootDirectory
	f := root.newFile("fName")
	c0 := w.pressKeyWrap(actorId, pressKeyCmd0)

	w.actors[actorId].currItemId = f.id()
	assert.True(t, c0.Ready())
	assert.Equal(t, len(f.lines), 1)
	assert.Equal(t, len(f.lines[0].characters), 0)
	c0.Step()
	assert.Equal(t, len(f.lines), 1)
	assert.Equal(t, len(f.lines[0].characters), 1)
	assert.Equal(t, f.lines[0].characters[0].shape, pressKeyCmds[pressKeyCmd0])
}

func TestSpecialKey(t *testing.T) {
	w := newTextWorld()
	actorId, _ := w.NewActor()

	root := w.rootDirectory
	f := root.newFile("fName")

	cErr := w.specialKeyWrap(actorId, pressKeyCmdLeft)
	w.actors[actorId].currItemId = -1
	assert.False(t, cErr.Ready())
	cErr.Step()

	w.actors[actorId].currItemId = f.id()

	assert.Nil(t, w.specialKeyWrap(actorId, pressKeyCmd0))
	ciL := w.specialKeyWrap(actorId, pressKeyCmdLeft)
	ciR := w.specialKeyWrap(actorId, pressKeyCmdRight)
	ciB := w.specialKeyWrap(actorId, pressKeyCmdBackspace)
	ciE := w.specialKeyWrap(actorId, pressKeyCmdEnter)
	ciU := w.specialKeyWrap(actorId, pressKeyCmdUp)
	ciD := w.specialKeyWrap(actorId, pressKeyCmdDown)

	assert.False(t, ciB.Ready())
	assert.False(t, ciL.Ready())
	assert.False(t, ciU.Ready())

	// create three chars in file
	c0 := w.pressKeyWrap(actorId, pressKeyCmd0)
	c1 := w.pressKeyWrap(actorId, pressKeyCmd1)
	c2 := w.pressKeyWrap(actorId, pressKeyCmd2)
    c0.Step()
    c1.Step()
    c2.Step()

    assert.Equal(t, w.actors[actorId].cursorChar, 3)
    assert.False(t, ciR.Ready())

    ciL.Step()
    assert.Equal(t, w.actors[actorId].cursorChar, 2)
    ciR.Step()
    assert.Equal(t, w.actors[actorId].cursorChar, 3)

    ciB.Step()
    assert.Equal(t, w.actors[actorId].cursorChar, 2)
    assert.Equal(t, len(f.lines[0].characters), 2)

    ciL.Step()
    ciE.Step()
    assert.Equal(t, w.actors[actorId].cursorChar, 0)
    assert.Equal(t, w.actors[actorId].cursorLine, 1)

    ciU.Step()
    assert.Equal(t, w.actors[actorId].cursorLine, 0)

    ciD.Step()
    assert.Equal(t, w.actors[actorId].cursorLine, 1)
    assert.False(t, ciD.Ready())
    ciD.Step()
}
