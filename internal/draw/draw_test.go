package draw

import (
	"errors"
	"image"
	_ "image/png" // Register PNGs for tests
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/spenserblack/termage/internal/conversion"
)

// TestRedraw checks that the screen would be cleared and a new image and
// title drawn.
func TestRedraw(t *testing.T) {
	f, err := os.Open(getResource("black-and-transparent-2x2.png"))
	defer f.Close()
	if err != nil {
		panic(err)
	}
	i, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	s := NewMockScreen(4, 7)
	runes := conversion.RGBRunesFromImage(i)
	Title(s, "test")
	Image(s, runes, image.Point{1, 1})

	redrawnTitle := "TEST"
	Redraw(s, redrawnTitle, runes, image.Point{0, 0})

	expectedTitle := []rune(redrawnTitle)
	for i, pixel := range s.pixels[0] {
		if actual, expected := pixel.mainc, expectedTitle[i]; actual != expected {
			t.Errorf(`Title row %d = %q, want %q`, i, actual, expected)
		}
	}

	expectedImage := [][]rune{
		{' ', ' ', ' ', ' '},
		{' ', '█', ' ', ' '},
		{' ', ' ', '█', ' '},
		{' ', ' ', ' ', ' '},
	}

	for y, row := range s.pixels[3:] {
		for x, pixel := range row {
			if actual, expected := pixel.mainc, expectedImage[x][y]; actual != expected {
				t.Errorf(`pixel (%d, %d) = %q, want %q`, x, y, actual, expected)
			}
		}
	}
}

// TestDrawTitle checks that the title would be properly drawn at the top-
// center of the screen.
func TestDrawTitle(t *testing.T) {
	s := NewMockScreen(10, 5)
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

// TestDrawMultilineTitle checks that a title cannot fit in one line will get
// wrapped to the next line.
func TestDrawMultilineTitle(t *testing.T) {
	const width int = 7
	s := NewMockScreen(width, 6)
	Title(s, "11111 222")
	var (
		row1 = make([]rune, width, width)
		row2 = make([]rune, width, width)
	)
	for i, p := range s.pixels[0] {
		row1[i] = p.mainc
	}
	for i, p := range s.pixels[1] {
		row2[i] = p.mainc
	}

	if actual, expected := string(row1), " 11111 "; actual != expected {
		t.Errorf(`Row 1 = %q, want %q`, actual, expected)
	}
	if actual, expected := string(row2), "  222  "; actual != expected {
		t.Errorf(`Row 2 = %q, want %q`, actual, expected)
	}
}

// TestDrawMultipleTitles checks that a title can be drawn over another title
// without retaining any artifacts from the old title.
func TestDrawMultipleTitles(t *testing.T) {
	const width int = 5
	s := NewMockScreen(width, 6)
	Title(s, "aaaaa b")
	Title(s, "foo bar")
	var (
		row1 = make([]rune, width, width)
		row2 = make([]rune, width, width)
	)
	for i, p := range s.pixels[0] {
		row1[i] = p.mainc
	}
	for i, p := range s.pixels[1] {
		row2[i] = p.mainc
	}

	if actual, expected := string(row1), " foo "; actual != expected {
		t.Errorf(`Row 1 = %q, want %q`, actual, expected)
	}
	if actual, expected := string(row2), " bar "; actual != expected {
		t.Errorf(`Row 2 = %q, want %q`, actual, expected)
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
	s := NewMockScreen(4, 7)
	Image(s, conversion.RGBRunesFromImage(i), image.Point{0, 0})

	if actual, expected := s.pixels[4][1].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 1, 4 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[4][2].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 2, 4 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[5][1].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 1, 5 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[5][2].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 2, 5 = %q, want %q`, actual, expected)
	}

	var actualForeground tcell.Color
	actualForeground, _, _ = s.pixels[4][1].style.Decompose()
	if actual := actualForeground.Hex(); actual != 0x000000 {
		t.Errorf(`foreground @ 1, 2 = %v, want %v`, actual, 0x000000)
	}
	actualForeground, _, _ = s.pixels[5][2].style.Decompose()
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
	s := NewMockScreen(4, 7)
	Title(s, "test")
	Image(s, conversion.RGBRunesFromImage(i), image.Point{0, -2})

	for x, r := range "test" {
		if actual := s.pixels[0][x].mainc; actual != r {
			t.Errorf(`Title bar rune %d = %v, want %v`, x, actual, r)
		}
	}
	if actual, expected := s.pixels[3][1].mainc, ' '; actual != expected {
		t.Errorf(`rune @ 1, 3 = %q, want %q`, actual, expected)
	}
	if actual, expected := s.pixels[3][2].mainc, '█'; actual != expected {
		t.Errorf(`rune @ 2, 3 = %q, want %q`, actual, expected)
	}
}

// TestDrawError checks that an error message can be drawn to the screen.
func TestDrawError(t *testing.T) {
	s := NewMockScreen(12, 13)
	// TODO line 5 "cannot draw:"
	// TODO line 6 error message
	Error(s, errors.New("test"))

	statusLine := make([]rune, 0, 12)
	for _, p := range s.pixels[7] {
		statusLine = append(statusLine, p.mainc)
	}
	if actual, expected := string(statusLine), "cannot draw:"; actual != expected {
		t.Errorf(`Status line = %q, want %q`, actual, expected)
	}

	errorLine := make([]rune, 0, 12)
	for _, p := range s.pixels[8] {
		errorLine = append(errorLine, p.mainc)
	}
	if actual, expected := string(errorLine), "    test    "; actual != expected {
		t.Errorf(`Error line = %q, want %q`, actual, expected)
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
	var emptyStyle tcell.Style
	for y := range s.pixels {
		for x := range s.pixels[y] {
			s.pixels[y][x].mainc = ' '
			s.pixels[y][x].combc = nil
			s.pixels[y][x].style = emptyStyle
		}
	}
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
	if x < 0 || x >= s.width || y < 0 || y >= s.height {
		return
	}
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
