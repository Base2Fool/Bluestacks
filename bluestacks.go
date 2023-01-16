package bluestacks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/g4s8/hexcolor"
	"github.com/go-vgo/robotgo"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strings"
)

var (
	redSum1   float64
	greenSum1 float64
	blueSum1  float64
	redSum2   float64
	greenSum2 float64
	blueSum2  float64
	redSum3   float64
	greenSum3 float64
	blueSum3  float64
)

type PxColorPipe struct {
	Reader                   io.Reader
	Output                   io.Writer
	HexMax                   int
	McCoords                 []MouseCursorCoords
	Colors                   Colors
	ColorsAvg                ColorsAverage
	BufOne, BufTwo, BufThree bytes.Buffer
	Err                      error
}

type Colors struct {
	RGBAs [][]color.RGBA
}

type ColorsAverage struct {
	FirstColor  color.RGBA
	SecondColor color.RGBA
	ThirdColor  color.RGBA
}

type MouseCursorCoords struct {
	X, Y int
}

type TriHexColors struct {
	FirstHexColor  string
	SecondHexColor string
	ThirdHexColor  string
}

func (px *PxColorPipe) Average(useSquaredAverage bool) *PxColorPipe {
	if px.Err != nil {
		px.Reader = strings.NewReader("")
		return px
	}
	var count float64
	for i, sliceRgb := range px.Colors.RGBAs {
		switch i {
		case 0:
			for _, col := range sliceRgb {
				if useSquaredAverage {
					redSum1 += float64(col.R) * float64(col.R)
					greenSum1 += float64(col.G) * float64(col.G)
					blueSum1 += float64(col.B) * float64(col.B)
					count++
				} else {
					redSum1 += float64(col.R)
					greenSum1 += float64(col.G)
					blueSum1 += float64(col.B)
					count++
				}
			}
		case 1:
			for _, col := range sliceRgb {
				if useSquaredAverage {
					redSum2 += float64(col.R) * float64(col.R)
					greenSum2 += float64(col.G) * float64(col.G)
					blueSum2 += float64(col.B) * float64(col.B)
				} else {
					redSum2 += float64(col.R)
					greenSum2 += float64(col.G)
					blueSum2 += float64(col.B)
				}
			}
		case 2:
			for _, col := range sliceRgb {
				if useSquaredAverage {
					redSum3 += float64(col.R) * float64(col.R)
					greenSum3 += float64(col.G) * float64(col.G)
					blueSum3 += float64(col.B) * float64(col.B)
				} else {
					redSum3 += float64(col.R)
					greenSum3 += float64(col.G)
					blueSum3 += float64(col.B)
				}
			}

		}
	}
	var firstHexAvg string
	var secondHexAvg string
	var thirdHexAvg string
	if useSquaredAverage {
		redAvg1 := uint8(math.Round(math.Sqrt(redSum1 / count)))
		greenAvg1 := uint8(math.Round(math.Sqrt(greenSum1 / count)))
		blueAvg1 := uint8(math.Round(math.Sqrt(blueSum1 / count)))

		redAvg2 := uint8(math.Round(math.Sqrt(redSum2 / count)))
		greenAvg2 := uint8(math.Round(math.Sqrt(greenSum2 / count)))
		blueAvg2 := uint8(math.Round(math.Sqrt(blueSum2 / count)))

		redAvg3 := uint8(math.Round(math.Sqrt(redSum3 / count)))
		greenAvg3 := uint8(math.Round(math.Sqrt(greenSum3 / count)))
		blueAvg3 := uint8(math.Round(math.Sqrt(blueSum3 / count)))

		px.ColorsAvg = ColorsAverage{
			FirstColor: color.RGBA{
				R: redAvg1,
				G: greenAvg1,
				B: blueAvg1,
				A: 255,
			},
			SecondColor: color.RGBA{
				R: redAvg2,
				G: greenAvg2,
				B: blueAvg2,
				A: 255,
			},
			ThirdColor: color.RGBA{
				R: redAvg3,
				G: greenAvg3,
				B: blueAvg3,
				A: 255,
			},
		}
		firstHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg1, greenAvg1, blueAvg1)
		secondHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg2, greenAvg2, blueAvg2)
		thirdHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg3, greenAvg3, blueAvg3)
	} else {
		redAvg1 := uint8(math.Round(redSum1 / count))
		greenAvg1 := uint8(math.Round(greenSum1 / count))
		blueAvg1 := uint8(math.Round(blueSum1 / count))

		redAvg2 := uint8(math.Round(redSum2 / count))
		greenAvg2 := uint8(math.Round(greenSum2 / count))
		blueAvg2 := uint8(math.Round(blueSum2 / count))

		redAvg3 := uint8(math.Round(redSum3 / count))
		greenAvg3 := uint8(math.Round(greenSum3 / count))
		blueAvg3 := uint8(math.Round(blueSum3 / count))

		px.ColorsAvg = ColorsAverage{
			FirstColor: color.RGBA{
				R: redAvg1,
				G: greenAvg1,
				B: blueAvg1,
				A: 255,
			},
			SecondColor: color.RGBA{
				R: redAvg2,
				G: greenAvg2,
				B: blueAvg2,
				A: 255,
			},
			ThirdColor: color.RGBA{
				R: redAvg3,
				G: greenAvg3,
				B: blueAvg3,
				A: 255,
			},
		}

		firstHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg1, greenAvg1, blueAvg1)
		secondHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg2, greenAvg2, blueAvg2)
		thirdHexAvg = fmt.Sprintf("%02x%02x%02x", redAvg3, greenAvg3, blueAvg3)

	}
	px.Reader = strings.NewReader(firstHexAvg + " " + secondHexAvg + " " + thirdHexAvg)
	return px
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
			robotgo.MilliSleep(300)
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

func Kill() error {
	pid, err := robotgo.FindIds("hd-player.exe")
	if err != nil {
		return err
	}
	// TODO: make a better error msg. Why isn't the above func erring when antivirus blocks run?
	if len(pid) < 1 {
		return errors.New("no pid was found. Try disable antivirus")
	}
	pidInt := pid[0]
	p, err := os.FindProcess(int(pidInt))
	if err != nil {
		return err
	}
	if err := p.Kill(); err != nil {
		return err
	}
	return nil
}

func MaxWindow(x, y int) {
	robotgo.MoveClick(x, y)
}

// Run runs Bluestacks
func Run(path string) error {
	_, err := robotgo.Run(path)
	if err != nil {
		return err
	}
	return nil
}

// RunStacks Trivial wrapper of Run
func RunStacks() error {
	err := Run("C:/Program Files/BlueStacks_nxt/HD-Player.exe")
	if err != nil {
		return err
	}
	return nil
}

func StartApp(x, y int) {
	robotgo.MoveClick(x, y)
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
		FirstHexColor:  string(data[:6]),
		SecondHexColor: string(data[7:13]),
		ThirdHexColor:  string(data[14:]),
	}
	dataJson, err := json.Marshal(thc)
	if err != nil {
		px.Err = err
		return px
	}
	px.Reader = bytes.NewReader(dataJson)
	return px
}
