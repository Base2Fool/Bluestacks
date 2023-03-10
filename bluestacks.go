package bluestacks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/g4s8/hexcolor"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
)

type PxColorPipe struct {
	Reader                   io.Reader
	Output                   io.Writer
	HttpClient               *http.Client
	URL                      string
	HexMax                   int
	McCoords                 []MouseCursorCoords
	Colors                   Colors
	Opaques                  OpaqueColors
	BufOne, BufTwo, BufThree bytes.Buffer
	Err                      error
}

type option func(px *PxColorPipe) error

type Colors struct {
	RGBAs [][]color.RGBA
}

type OpaqueColors struct {
	Colors []color.RGBA
}

type MouseCursorCoords struct {
	X, Y int
}

type TriHexColors struct {
	First  string `json:"first"`
	Second string `json:"second"`
	Third  string `json:"third"`
}

func NewPxColorPipe(opts ...option) (PxColorPipe, error) {
	px := PxColorPipe{
		Output:     os.Stdout,
		HttpClient: http.DefaultClient,
		HexMax:     36,
	}
	for _, opt := range opts {
		err := opt(&px)
		if err != nil {
			return PxColorPipe{}, err
		}
	}
	return px, nil
}

func WithInputFromArgs(args []string) option {
	return func(px *PxColorPipe) error {
		if len(args) < 1 {
			return errors.New("an url argument is required")
		}
		if len(args) > 1 {
			return errors.New("command only takes one argument")
		}
		px.URL = args[0]
		return nil
	}
}

func ByPickingPixels() []MouseCursorCoords {
	fmt.Println("Please wait a few seconds, then click three positions.")
	var coords []int16
	var clicks int
	evChan := hook.Start()
	for ev := range evChan {
		if ev.Clicks == 1 && ev.Kind == 8 {
			coords = append(coords, ev.X, ev.Y)
			clicks++
			if clicks == 3 {
				break
			}
		}
	}
	hook.End()
	mcCoords := []MouseCursorCoords{
		{X: int(coords[0]), Y: int(coords[1])},
		{X: int(coords[2]), Y: int(coords[3])},
		{X: int(coords[4]), Y: int(coords[5])},
	}
	return mcCoords
}

func (px *PxColorPipe) Stdout() *PxColorPipe {
	if px.Err != nil {
		return px
	}
	_, err := io.Copy(px.Output, px.Reader)
	if err != nil {
		px.Err = err
	}
	return px
}

func (px *PxColorPipe) Hex() *PxColorPipe {
	if px.HexMax <= 0 {
		px.Err = errors.New("hexmax field must be greater than 0")
		px.Reader = strings.NewReader("")
		return px
	}
	if len(px.McCoords) == 0 {
		px.Err = errors.New("no mouse cursor coordinate has been set")
		px.Reader = strings.NewReader("")
		return px
	}
	for i := 0; i < px.HexMax; i++ {
		for i, v := range px.McCoords {
			robotgo.Move(v.X, v.Y)
			switch i {
			case 0:
				px.BufOne.WriteString("#" + robotgo.GetMouseColor() + " ")
			case 1:
				px.BufTwo.WriteString("#" + robotgo.GetMouseColor() + " ")
			case 2:
				px.BufThree.WriteString("#" + robotgo.GetMouseColor() + " ")
			}
			robotgo.MilliSleep(200)
		}
	}
	return px

}

// HexToRGBA TODO: Handle Errors.
func (px *PxColorPipe) HexToRGBA() *PxColorPipe {
	if px.Err != nil {
		px.Colors = Colors{RGBAs: [][]color.RGBA{}}
		return px
	}
	buffs := []bytes.Buffer{px.BufOne, px.BufTwo, px.BufThree}
	var RGBsOne []color.RGBA
	var RGBsTwo []color.RGBA
	var RGBsThree []color.RGBA
	for i, buf := range buffs {
		switch i {
		case 0:
			hexCols := strings.Fields(buf.String())
			for _, hc := range hexCols {
				c, err := hexcolor.Parse(hc)
				if err != nil {
					log.Fatal(err)
				}
				RGBsOne = append(RGBsOne, c)
			}
		case 1:
			hexCols := strings.Fields(buf.String())
			for _, hc := range hexCols {
				c, err := hexcolor.Parse(hc)
				if err != nil {
					log.Fatal(err)
				}
				RGBsTwo = append(RGBsTwo, c)
			}
		case 2:
			hexCols := strings.Fields(buf.String())
			for _, hc := range hexCols {
				c, err := hexcolor.Parse(hc)
				if err != nil {
					log.Fatal(err)
				}
				RGBsThree = append(RGBsThree, c)
			}
		}
	}
	px.BufOne.Reset()
	px.BufTwo.Reset()
	px.BufThree.Reset()
	px.Colors = Colors{RGBAs: [][]color.RGBA{RGBsOne, RGBsTwo, RGBsThree}}
	return px
}

// TODO: Handle Error! Perhaps writing to a log file?
func RunCLI() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Kill)
		<-c
		os.Exit(1)
	}()
	px, err := NewPxColorPipe(WithInputFromArgs(os.Args[1:]))
	if err != nil {
		log.Fatal(err)
	}
	px.McCoords = ByPickingPixels()
	for {
		_, err := px.Hex().HexToRGBA().Opacity().OpaqueToHex().ToJson().Patch()
		if err != nil {
			log.Println(err)
		}
	}
}

func (px *PxColorPipe) String() (string, error) {
	if px.Err != nil {
		return "", px.Err
	}
	data, err := io.ReadAll(px.Reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (px *PxColorPipe) ToJson() *PxColorPipe {
	if px.Err != nil {
		px.Reader = strings.NewReader("")
		return px
	}
	data, err := io.ReadAll(px.Reader)
	if err != nil {
		px.Err = err
		return px
	}
	if len(string(data)) != 20 {
		px.Err = errors.New("toJson: length of input reader is invalid. Must be of length 20")
		return px
	}
	thc := TriHexColors{
		First:  string(data[:6]),
		Second: string(data[7:13]),
		Third:  string(data[14:]),
	}
	dataJson, err := json.Marshal(thc)
	if err != nil {
		px.Err = err
		return px
	}
	px.Reader = bytes.NewReader(dataJson)
	return px
}

func (px *PxColorPipe) Opacity() *PxColorPipe {
	if px.Err != nil {
		px.Opaques = OpaqueColors{Colors: []color.RGBA{}}
		return px
	}
	for _, rgbaBunch := range px.Colors.RGBAs {
		sort.Slice(rgbaBunch, func(i, j int) bool {
			return rgbaBunch[i].R < rgbaBunch[j].R
		})
	}
	px.Opaques = OpaqueColors{
		Colors: []color.RGBA{
			px.Colors.RGBAs[0][0],
			px.Colors.RGBAs[1][0],
			px.Colors.RGBAs[2][0]},
	}
	return px
}

func (px *PxColorPipe) OpaqueToHex() *PxColorPipe {
	if px.Err != nil {
		px.Reader = strings.NewReader("")
		return px
	}
	firstHex := fmt.Sprintf(
		"%02x%02x%02x",
		px.Opaques.Colors[0].R,
		px.Opaques.Colors[0].G,
		px.Opaques.Colors[0].B)
	secondHex := fmt.Sprintf(
		"%02x%02x%02x",
		px.Opaques.Colors[1].R,
		px.Opaques.Colors[1].G,
		px.Opaques.Colors[1].B)
	thirdHex := fmt.Sprintf(
		"%02x%02x%02x",
		px.Opaques.Colors[2].R,
		px.Opaques.Colors[2].G,
		px.Opaques.Colors[2].B)
	px.Reader = strings.NewReader(firstHex + " " + secondHex + " " + thirdHex)
	return px
}

func (px *PxColorPipe) Patch() (*http.Response, error) {
	if px.Err != nil {
		return nil, px.Err
	}
	req, err := http.NewRequest(http.MethodPatch, px.URL, px.Reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return px.HttpClient.Do(req)
}
