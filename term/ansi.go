package term

import (
	"fmt"

	"github.com/dmsRosa6/glyph/core"
)

//TODO i removed the old state comparation, eventually come to this again
func CellToANSI(c core.Cell) string {
	seq := ""

	if(!c.Bg.IsTransparent){
		seq += bgToANSI(c.Bg)
	}
	
	if(!c.Fg.IsTransparent){
		seq += fgToANSI(c.Fg)
		seq += string(c.Ch)
	}

	if(c.Ch == 0){
		seq += " "
	}
	
	return seq
}

func fgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", color.R, color.G, color.B) 
}

func bgToANSI(color core.Color) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", color.R, color.G, color.B) 
}