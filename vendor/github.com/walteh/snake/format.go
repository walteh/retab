package snake

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func FormatCaller(path string, number int) string {
	tot := strings.Split(path, "/")
	if len(tot) > 2 {
		last := tot[len(tot)-1]
		secondLast := tot[len(tot)-2]
		thirdLast := tot[len(tot)-3]
		return fmt.Sprintf("%s/%s %s:%s", thirdLast, secondLast, color.New(color.Bold).Sprint(last), color.New(color.FgHiRed, color.Bold).Sprintf("%d", number))
	} else {
		return fmt.Sprintf("%s:%d", path, number)
	}
}
