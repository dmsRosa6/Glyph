package render

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
)

//TODO i removed the old state comparation, eventually come to this again
func cellToANSI(c core.Cell, prev core.Cell) string {
	seq := ""
	seq += fgToANSI(c.Fg)
	seq += bgToANSI(c.Bg)
	//if c.Fg != prev.Fg {
	//	seq += fgToANSI(c.Fg)
	//}
	//if c.Bg != prev.Bg {
	//	seq += bgToANSI(c.Bg)
	//}

	seq += string(c.Ch)
	return seq
}

func fgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.R, color.G, color.B) 
}

func bgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", color.R, color.G, color.B) 
}