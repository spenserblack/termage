package draw

import (
	"image"
	_ "image/png" // Register PNGs for tests
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/spenserblack/termage/internal/conversion"
)

// TestDrawTitle checks that the title would be properly drawn at the top-
// center of the screen.
func TestDrawTitle(t *testing.T) {
	s := NewMockScreen(10, 3)
	Title(s, "test")
	fail := false
	if actual := s.pixels[0][3].mainc; actual != 't' {
		t.Errorf(`pixel = %v, want t`, actual)
		fail = true
	}
	if actual := s.pixels[0][4].mainc; actual != 'e' {
		t.Errorf(`pixel = %v, want e`, actual)
		fail = true
	}
	if actual := s.pixels[0][5].mainc; actual != 's' {
		t.Errorf(`pixel = %v, want s`, actual)
		fail = true
	}
	if actual := s.pixels[0][6].mainc; actual != 't' {
		t.Errorf(`pixel = %v, want t`, actual)
		fail = true
	}
	if fail {
		width, _ := s.Size()
		titleRow := make([]rune, width)
		for i := range titleRow {
			titleRow[i] = s.pixels[0][i].mainc
		}
		t.Fatalf(`Title row = %q`, string(titleRow))
	}
}

// TestDrawImage checks that a standard image fitting the screen can be drawn.
func TestDrawImage(t *testing.T) {
	f, err := os.Open(getResource("black-and-transparent-2x2.png"))
	defer f.Close()
	if err != nil {
		panic(err)
	}
	i, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	s := NewMockScreen(4, 5)
	Image(s, conversion.RGBRunesFromImage(i), image.Point{0, 0})

	if actual, expected := s.pixels[2][1].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 1, 2 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[2][2].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 2, 2 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[3][1].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 1, 3 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[3][2].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 2, 3 = %q, want %q`, actual, expected)
	}

	var actualForeground tcell.Color
	actualForeground, _, _ = s.pixels[2][1].style.Decompose()
	if actual := actualForeground.Hex(); actual != 0x000000 {
		t.Errorf(`foreground @ 1, 2 = %v, want %v`, actual, 0x000000)
	}
	actualForeground, _, _ = s.pixels[3][2].style.Decompose()
	if actual := actualForeground.Hex(); actual != 0x000000 {
		t.Errorf(`foreground @ 2, 3 = %v, want %v`, actual, 0x000000)
	}
}

// TestDrawImageOverlapTitle checks that a standard image fitting the screen
// can be drawn.
func TestDrawImageOverlapTitle(t *testing.T) {
	f, err := os.Open(getResource("black-and-transparent-2x2.png"))
	defer f.Close()
	if err != nil {
		panic(err)
	}
	i, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	s := NewMockScreen(4, 5)
	Title(s, "test")
	Image(s, conversion.RGBRunesFromImage(i), image.Point{0, -2})

	for x, r := range "test" {
		if actual := s.pixels[0][x].mainc; actual != r {
			t.Errorf(`Title bar rune %d = %v, want %v`, x, actual, r)
		}
	}
	if actual, expected := s.pixels[1][1].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 1, 1 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[1][2].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 2, 1 = %q, want %q`, actual, expected)
	}
}

type MockScreen struct {
	width, height int
	pixels        [][]MockPixel
}

type MockPixel struct {
	mainc rune
	combc []rune
	style tcell.Style
}

func NewMockScreen(width, height int) *MockScreen {
	pixels := make([][]MockPixel, height)
	for i := range pixels {
		row := make([]MockPixel, width)
		for i := range row {
			row[i] = MockPixel{}
			row[i].mainc = ' '
		}
		pixels[i] = row
	}
	return &MockScreen{width, height, pixels}
}

func (s *MockScreen) Init() error {
	return nil
}

func (s *MockScreen) Fini()                  {}
func (s *MockScreen) Fill(rune, tcell.Style) {}

func (s *MockScreen) Clear() {
	panic("Not implemented")
}

func (s *MockScreen) SetCell(x int, y int, style tcell.Style, ch ...rune) {
	panic("Not implemented")
}

func (s *MockScreen) GetContent(x, y int) (
	mainc rune,
	combc []rune,
	style tcell.Style,
	width int,
) {
	panic("Not implemented")
}

func (s *MockScreen) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	s.pixels[y][x].mainc = mainc
	s.pixels[y][x].combc = combc
	s.pixels[y][x].style = style
}

func (s *MockScreen) SetStyle(style tcell.Style) {}

func (s *MockScreen) ShowCursor(int, int) {
}

func (s *MockScreen) HideCursor() {
}

func (s *MockScreen) Size() (int, int) {
	return s.width, s.height
}

func (s *MockScreen) ChannelEvents(ch chan<- tcell.Event, quit <-chan struct{}) {
}

func (s *MockScreen) PollEvent() tcell.Event {
	return nil
}

func (s *MockScreen) HasPendingEvent() bool {
	return false
}

func (s *MockScreen) PostEvent(ev tcell.Event) error {
	return nil
}

func (s *MockScreen) PostEventWait(tcell.Event) {}

func (s *MockScreen) EnableMouse(...tcell.MouseFlags) {}
func (s *MockScreen) DisableMouse()                   {}
func (s *MockScreen) EnablePaste()                    {}
func (s *MockScreen) DisablePaste()                   {}

func (s *MockScreen) HasMouse() bool {
	return false
}

func (s *MockScreen) Colors() int {
	return 0
}

func (s *MockScreen) Show() {}
func (s *MockScreen) Sync() {}

func (s *MockScreen) CharacterSet() string {
	return ""
}

func (s *MockScreen) RegisterRuneFallback(rune, string) {}
func (s *MockScreen) UnregisterRuneFallback(rune)       {}

func (s *MockScreen) CanDisplay(rune, bool) bool {
	return true
}

func (s *MockScreen) Resize(int, int, int, int) {}

func (s *MockScreen) HasKey(tcell.Key) bool {
	return true
}

func (s *MockScreen) Suspend() error {
	return nil
}

func (s *MockScreen) Resume() error {
	return nil
}

func (s *MockScreen) Beep() error {
	return nil
}

func thisDirOrPanic() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("Couldn't get directory of test")
	}
	return filepath.Dir(file)
}

func getResource(resourceName string) string {
	dir := thisDirOrPanic()
	return filepath.Join(dir, "..", "..", "_resources", "tests", "internal", "draw", resourceName)
}
