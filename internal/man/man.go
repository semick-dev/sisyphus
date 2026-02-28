package man

import (
	"fmt"
	"strings"
	"time"
)

var sisyphusFrames = []string{
`         _          ________
        / \_       /        \
        \ /       /          \
        /_______//            \
       /________/|            |                       ||||||||
      /          |            |                  |||||
     /           |            |          ||||||||
    / \           \          /   ||||||||
    |  \            \______/|||||
    |   |       ||||||||||||
   /    |_ |||||
  /_  |||||
||||||`,

`           _          ________
          / \_       /        \
          \ /       /          \
          /_______//            \
         /________/|            |                     ||||||||
        /          |            |                |||||
       /           |            |        ||||||||
      / \           \          /  ||||||||
      |  \            \______/|||||
      |   |     ||||||||||||
     /    |_|||||
    /_  |||||
  ||||||`,

`              _          ________
             / \_       /        \
             \ /       /          \
             /_______//            \
            /________/|            |                 ||||||||
           /          |            |            |||||
          /           |            |    ||||||||
         / \           \          /||||||
         |  \            \______/|||||
         |   |   ||||||||||||
        /    |_|||||
       /_  |||||
     ||||||`,

`                 _          ________
                / \_       /        \
                \ /       /          \
                /_______//            \
               /________/|            |           ||||||||
              /          |            |      |||||
             /           |            ||||||||
            / \           \          /||||||
            |  \            \______/|||||
            |   | ||||||||||||
           /    |_|||||
          /_  |||||
        ||||||`,

`                    _          ________
                   / \_       /        \
                   \ /       /          \
                   /_______//            \
                  /________/|            |     ||||||||
                 /          |            |||||
                /           |    ||||||||
               / \           \          /||||||
               |  \            \______/|||||
               |   ||||||||||||
              /    |_|||||
             /_  |||||
           ||||||`,

`                       _          ________
                      / \_       /        \
                      \ /       /          \
                      /_______//            \
                     /________/|           ||||||||
                    /          |      |||||
                   /      ||||||||
                  / \           \          /||||||
                  |  \            \______/|||||
                  |   ||||||||||||
                 /    |_|||||
                /_  |||||
              ||||||`,
}

var baseArt = []string{
	"                                                                    ...",
	"                                                                  .##-.",
	"                                           .. ..... .....         .##-.",
	"                                             ..-)}}(+:.=~.       .<##-.",
	"                                            ..>{{{####{#(.   .. .=>##-.",
	"                                            ..##{#{><]{[:..   ..:>*##-.",
	"                                            ..{#**<****.... ....:>*##-.",
	"                                           ...}(******=.....-*>***>##-.",
	"                                    .......::+**<****~.:*>*****>+..##-.",
	"                                     .:>^^<<>^^^^^<*********>^:....##-.",
	"                                 ....)^^^^^^^^^^^*<******>>~.......##-.",
	"                                 ...<^^^^^^^^^^^^*)<<^=-:..........##-.",
	"                             .....:>^^^^^^<<<<<>^~......  ..   ....##-.",
	"                             .....>^^^^^^^^^^^^^: ..             ..##-.",
	"                              ...<^^^^^^^^^^^^^+. .              ..##-.",
	"                             ...)^^^^^^^^^^^^<-.                 ..##-.",
	"                              .<^^^^^^^^^^<>*. ....              ..##-.",
	"                           ..~>^^^^^^^<>^^<*.......              ..##-.",
	"                          ..>^^^^^^^<))^^^^:.......              ..##-.",
	"                          .:#{#{]<^^^^^^^^~....                  ..##-.",
	"                          .]{#########{[(^..   .                 ..##-.",
	"                          .#############}{{+.  ....              ..##-.",
	"                          .#{##########]#####+.....              ..##-.",
	"                         .)#{########{]{#######=...              ..##-.",
	"                     ....-{#########{[##########{:.              ..##-.",
	"                     ....}#########{+{###########{]..            ..##-.",
	"                      ..>########{)....=}{#########}...          ..##-.",
	"                      .~########{=........:)#{#####{...          ..##-.",
	"                  ....-########}.           .[#####]...          ..##-.",
	"                   .(#########=..            {###{{^...          ..##-.",
	"               ...}{{{{{{###^...            .######:.            ..##-.",
	"               .]###{{{{{}:...              .%####[.             ..##-.",
	"           ...*####{{##^.....               -####{=              ..##-.",
	"          ..:#######]. .                   .^{####.              ..##-.",
	"       . ..>{{{{{#=...                     .[###{+..  .          ..##-.",
	"       .-}{**{#<. .....                    .{{###. .  .          ..##-.",
	"      .:#{###:.                            .#{{{+...  .          ..##-.",
	" ...  ...>###}.....          ....        ..:}(]{<.... ....   ......##-.",
	"        ..:{{{}<:...         ....        ..(#{#{#{#{)-...      ....##-.",
	"::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::",
}

var offsets = []int{0, 1, 2, 3, 2, 1, 0, -1, -2, -1, 0, 1}

func shiftRegion(text string, offset int, width int) string {
	if offset > 0 {
		shifted := strings.Repeat(" ", offset) + text
		if len(shifted) > width {
			return shifted[:width]
		}
		return shifted
	}
	if offset < 0 {
		start := -offset
		if start >= len(text) {
			return strings.Repeat(" ", width)
		}
		shifted := text[start:]
		if len(shifted) < width {
			shifted += strings.Repeat(" ", width-len(shifted))
		}
		return shifted
	}
	if len(text) >= width {
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}

func buildFrame(lines []string, offset int) string {
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	out := make([]string, 0, len(lines))
	floorIdx := len(lines) - 1
	for i, line := range lines {
		padded := line
		if len(padded) < maxWidth {
			padded += strings.Repeat(" ", maxWidth-len(padded))
		}
		if i == floorIdx {
			out = append(out, padded)
			continue
		}
		out = append(out, shiftRegion(padded, offset, maxWidth))
	}
	return strings.Join(out, "\n")
}

func BuildFrames() []string {
	frames := make([]string, 0, len(offsets))
	for _, offset := range offsets {
		frames = append(frames, buildFrame(baseArt, offset))
	}
	return frames
}

func RenderStatic() string {
	frames := BuildFrames()
	return frames[0]
}

func Animate(repeatDelay time.Duration) {
	frames := BuildFrames()
	frameHeight := len(baseArt)
	moveUp := frameHeight - 1
	first := true
	defer fmt.Println()

	for {
		for _, frame := range frames {
			if first {
				fmt.Print(frame)
				first = false
			} else {
				fmt.Printf("\033[%dF\033[J%s", moveUp, frame)
			}
			time.Sleep(repeatDelay)
		}
	}
}
