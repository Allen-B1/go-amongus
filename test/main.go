package main

import (
	"image/color"
	"os"
	"strconv"

	"github.com/allen-b1/go-amongus"
	"github.com/allen-b1/go-amongus/test/roles"
	"github.com/mitchellh/go-ps"
)

var (
	JesterColor   color.Color = color.RGBA{200, 50, 200, 255}
	CrewmateColor color.Color
	ImpostorColor color.Color
)

func main() {
	processNum, _ := strconv.Atoi(os.Args[1])
	pslist, err := ps.Processes()
	if err != nil {
		panic(err)
	}
	auPid := int32(-1)
	passed := 0
	for _, proc := range pslist {
		if proc.Executable() == "Among Us.exe" {
			auPid = int32(proc.Pid())
			if passed >= processNum {
				break
			}
			passed++
		}
	}
	if auPid < 0 {
		panic("Among Us not open")
	}

	au, err := amongus.NewFromPID(uint32(auPid))
	if err != nil {
		panic(err)
	}

	roles.Init(au)
	jesterID := roles.Register(&roles.Role{Name: "Jester", Helper: "Make Crewmates vote you off", Count: 1, Color: color.RGBA{230, 50, 200, 255}})
	for {
		roles.Loop()
		_ = jesterID
	}
}
