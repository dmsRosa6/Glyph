package framework

type SpinnerContext struct {
	cycle []string
	index int
	size  int
}

func NewSpinnerContext(cycle []string, size int) *SpinnerContext {
	return &SpinnerContext{cycle: cycle, size: size}
}

func NewSlashSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"\\", "|", "/", "-",
	}, 1)
}

func NewDotsSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		".  ",
		".. ",
		"...",
		" ..",
		"  .",
		"   ",
	}, 3)
}

func NewPulseSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"●○○",
		"○●○",
		"○○●",
		"○●○",
	}, 3)
}

func NewBounceSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"[=  ]",
		"[ = ]",
		"[  =]",
		"[ = ]",
	}, 5)
}

func NewBrailleSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"⠁", "⠂", "⠄", "⡀",
		"⢀", "⠠", "⠐", "⠈",
	}, 1)
}

func NewBlockSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"▏",
		"▎",
		"▍",
		"▌",
		"▋",
		"▊",
		"▉",
		"█",
		"▉",
		"▊",
		"▋",
		"▌",
		"▍",
		"▎",
	}, 1)
}

func NewGrowSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"▁",
		"▂",
		"▃",
		"▄",
		"▅",
		"▆",
		"▇",
		"█",
		"▇",
		"▆",
		"▅",
		"▄",
		"▃",
		"▂",
	}, 1)
}

func NewSquareSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"◰",
		"◳",
		"◲",
		"◱",
	}, 1)
}

func NewClockSpinnerContext() *SpinnerContext {
	return NewSpinnerContext([]string{
		"◴",
		"◷",
		"◶",
		"◵",
	}, 1)
}

func (s *SpinnerContext) Cycle() string {
	val := s.cycle[s.index]
	s.index = (s.index + 1) % len(s.cycle)
	return val
}

func (s *SpinnerContext) CycleSize() int {
	return len(s.cycle)
}

func (s *SpinnerContext) SpinnerLength() int {
	return s.size
}
