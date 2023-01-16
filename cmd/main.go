package main

import (
	"bluestacks"
	"encoding/json"
	"fmt"
	"github.com/go-vgo/robotgo"
	"image/color"
	"io"
	"log"
)

func main() {
	//mc1 := bluestacks.MouseCursorCoords{X: 699, Y: 328}
	//mc2 := bluestacks.MouseCursorCoords{X: 915, Y: 328}
	//mc3 := bluestacks.MouseCursorCoords{X: 1079, Y: 326}
	//go bluestacks.RunStacks()
	//robotgo.Sleep(15)
	//// Coordinates for a screen size of 1920 x 1080
	//bluestacks.MaxWindow(1836, 124)
	//robotgo.Sleep(1)
	//// Coordinates for the sixth app in max window
	//bluestacks.StartApp(1637, 193)
	//robotgo.Sleep(5)
	//ClickGotItButton()
	//robotgo.Sleep(3)
	//ClickActiveTicket()
	//robotgo.Sleep(1)
	//pxc := bluestacks.PxColorPipe{
	//	HexMax:   10,
	//	McCoords: []bluestacks.MouseCursorCoords{mc1, mc2, mc3},
	//	Output:   os.Stdout,
	//}
	//pxc.Hex().HexToRGBA().Average(true).ToJson().Stdout()

	pxc := bluestacks.PxColorPipe{
		HexMax: 2,
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
			},
		},
	}
	for i := 0; i < 10; i++ {
		pxc.Average(false).ToJson()
		cl := bluestacks.TriHexColors{}
		data, err := io.ReadAll(pxc.Reader)
		if err != nil {
			log.Fatalln(err)
		}
		err = json.Unmarshal(data, &cl)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(cl)
	}
}

func ClickActiveTicket() {
	robotgo.MoveClick(938, 461)
}

func ClickGotItButton() {
	robotgo.MoveClick(1185, 1006)
}

func CurrentMousePosition() {
	for {
		robotgo.MilliSleep(200)
		fmt.Println(robotgo.GetMousePos())
	}
}
