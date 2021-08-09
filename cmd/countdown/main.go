package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWidth() int {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return int(ws.Col)
}

func getHeight() int {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return int(ws.Row)
}

func sec2Human(insec int) string {
	var retv string
	hours := 0
	minutes := 0
	seconds := insec

	if seconds < 1 {
		return "None"
	}

	if seconds > 3600 {
		hours = int(seconds / 3600)
		seconds = seconds % 3600
	}

	if seconds > 60 {
		minutes = int(seconds / 60)
		seconds = seconds % 60
	}

	if hours > 0 {
		retv = fmt.Sprintf("%02d:", hours)
	}

	if minutes > 0  {
		retv = fmt.Sprintf("%s%02d:", retv, minutes)
	} else if (minutes == 0) && (hours > 0) {
		retv = fmt.Sprintf("%s00:", retv)
	}

	if len(retv) > 0 {
		retv = fmt.Sprintf("%s%02d", retv, seconds)
	} else {
		retv = strconv.Itoa(seconds) + " seconds"
	}

	return retv
}

// calculate the variouse changes
func getTimes(epoch int, duration int) (int, string) {
	cur_ts := time.Now()
	cur_sec := int(cur_ts.Unix())
	delta_sec := cur_sec - epoch
	remaining := duration - delta_sec
	percent := int(remaining * 100 / duration)
	return percent, sec2Human(remaining)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		os.Exit(0)
	}

	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		panic(nil)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	now := time.Now()
	start_secs := int(now.Unix())

	screenwidth := getWidth()
	screenheight := getHeight()
	padding := 10
	screenheight = (screenheight - (screenheight % 10)) / 2

	g0 := widgets.NewGauge()
	g0.Title = "Countdown"
	g0.Percent = 100
	g0.SetRect(padding, screenheight, screenwidth-padding, screenheight+7)

	g0.BarColor = ui.ColorGreen
	g0.BorderStyle.Fg = ui.ColorWhite
	g0.TitleStyle.Fg = ui.ColorCyan

	draw := func() {
		percent, tsstr := getTimes(start_secs, seconds)
		if percent > 65 {
			g0.BarColor = ui.ColorRed
		} else if percent > 33 {
			g0.BarColor = ui.ColorYellow
		} else {
			g0.BarColor = ui.ColorGreen
		}
		g0.Percent = percent
		g0.Title = fmt.Sprintf("Countdown %s", tsstr)
		ui.Render(g0)
	}

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	for {
		percent, _ := getTimes(start_secs, seconds)
		if percent < 1 {
			return
		}
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-ticker:
			draw()
		}
	}
}
