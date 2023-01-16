package bluestacks_test

import (
	"bluestacks"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-vgo/robotgo"
	"github.com/google/go-cmp/cmp"
	"image/color"
	"io"
	"strings"
	"testing"
)

func TestRunErrsWhenDefaultPathIsMissing(t *testing.T) {
	t.Parallel()
	err := bluestacks.Run("C:/Program Files/Bogus.exe")
	if err == nil {
		t.Error("want error, but got nil")
	}

}

func TestStdout(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	want := "#34a853"
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader(want),
		Output: buf,
	}
	pxc.Stdout()
	got := buf.String()
	if want != got {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func TestStdoutError(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader("#34a853"),
		Output: buf,
		Err:    errors.New("some non-nil error"),
	}
	pxc.Stdout()
	if buf.String() != "" {
		t.Errorf("want an empty string but got %s", buf.String())
	}

}

// Gets the three out of four colors of the playstore app icon when it's the first app on bluestacks
func TestHexWithHexMaxSetAtThree(t *testing.T) {
	go bluestacks.RunStacks()
	robotgo.Sleep(15)
	// Coordinates for a screen size of 1920 x 1080
	bluestacks.MaxWindow(1836, 124)
	robotgo.Sleep(1)
	pxc := bluestacks.PxColorPipe{
		HexMax: 3,
		McCoords: []bluestacks.MouseCursorCoords{
			{X: 244, Y: 182}, {X: 258, Y: 193}, {X: 240, Y: 206}},
	}
	pxc.Hex()
	want := "#34a853 #34a853 #34a853 #fabc05 #fabc05 #fabc05 #e94334 #e94334 #e94334 "
	r := io.MultiReader(&pxc.BufOne, &pxc.BufTwo, &pxc.BufThree)
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, string(got)) {
		t.Error(cmp.Diff(want, string(got)))
	}
}

func TestHexNoOpsWithInvalidHexMax(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		HexMax: 0,
		McCoords: []bluestacks.MouseCursorCoords{
			{X: 244, Y: 182}, {X: 258, Y: 193}, {X: 240, Y: 206}},
	}
	pxc.Hex()
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Errorf("want no output from Hex after invalid HexMax, but got %s", data)
	}
}

func TestHexSetsErrorWhenNoMouseCursorCoordinateIsSet(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		HexMax: 4,
	}
	pxc.Hex()
	if pxc.Err == nil {
		t.Error("want error when mcCoords field is empty, but got nil")
	}
}

func TestHexSetsErrorAfterInvalidHexMax(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		HexMax: 0,
		McCoords: []bluestacks.MouseCursorCoords{
			{X: 244, Y: 182}, {X: 258, Y: 193}, {X: 240, Y: 206}},
	}
	pxc.Hex()
	if pxc.Err == nil {
		t.Error("want error after invalid HexMax, but got nil")
	}
}

func TestHexStringConvertsToColorRGBA(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		BufOne:   *bytes.NewBufferString("#34a853 "),
		BufTwo:   *bytes.NewBufferString("#fabc05 "),
		BufThree: *bytes.NewBufferString("#e94334 "),
	}
	pxc.HexToRGBA()
	want := bluestacks.Colors{
		RGBAs: [][]color.RGBA{
			{{52, 168, 83, 255}},
			{{250, 188, 5, 255}},
			{{233, 67, 52, 255}},
		},
	}
	got := pxc.Colors
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestThreeBuffersResetAfterHexConversion(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		BufOne:   *bytes.NewBufferString("#34a853 "),
		BufTwo:   *bytes.NewBufferString("#fabc05 "),
		BufThree: *bytes.NewBufferString("#e94334 "),
	}
	pxc.HexToRGBA()
	want := 0
	got := pxc.BufOne.Len() + pxc.BufTwo.Len() + pxc.BufThree.Len()
	if want != got {
		t.Errorf("want zero data in buffers but got %d", got)
	}
}

func TestHexToRGBANoOpsWhenPxErrIsNonNil(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		BufOne:   *bytes.NewBufferString("#34a853 "),
		BufTwo:   *bytes.NewBufferString("#fabc05 "),
		BufThree: *bytes.NewBufferString("#e94334 "),
		Err:      errors.New("some non-nil error"),
	}
	pxc.HexToRGBA()
	want := bluestacks.Colors{RGBAs: [][]color.RGBA{}}
	got := pxc.Colors
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAverage(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{52, 168, 83, 255}, {250, 188, 5, 255}, {233, 67, 52, 255}},
				{{133, 23, 56, 255}, {212, 121, 32, 255}, {200, 72, 47, 255}},
				{{100, 168, 23, 255}, {255, 112, 13, 255}, {222, 27, 9, 255}},
			},
		},
	}
	pxc.Average(false)
	want := "b28d2f b6482d c0660f"
	got, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if want != string(got) {
		t.Errorf("want %q, but got %q", want, got)
	}
}

func TestAverageNoOpsWhenPxErrIsNonNil(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Err: errors.New("some non-nil error"),
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{52, 168, 83, 255}, {250, 188, 5, 255}, {233, 67, 52, 255}},
				{{133, 23, 56, 255}, {212, 121, 32, 255}, {200, 72, 47, 255}},
				{{100, 168, 23, 255}, {255, 112, 13, 255}, {222, 27, 9, 255}},
			},
		},
	}
	pxc.Average(false)
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Errorf("want data to have no output when pipe has"+
			" an non-nil error when it reaches average filter, but got %s", data)
	}

}

func TestAverageInColorRGBA(t *testing.T) {
	// Why does this fail with t.Parallel() ?
	//t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
				{{255, 255, 255, 255}, {255, 255, 255, 255}},
			},
		},
	}

	pxc.Average(false)
	want := bluestacks.ColorsAverage{
		FirstColor:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
		SecondColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		ThirdColor:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	got := pxc.ColorsAvg
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestAverageIsConsistentWhenColorsRemainTheSame(t *testing.T) {
	//t.Parallel()
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
	// Why does each iteration decrement the color fields by 1?
	for i := 0; i < 4; i++ {
		pxc.Average(false)
	}
	want := bluestacks.ColorsAverage{
		FirstColor:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
		SecondColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		ThirdColor:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	got := pxc.ColorsAvg
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}

func TestString(t *testing.T) {
	t.Parallel()
	want := "b28d2f b6482d c0660f"
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader(want),
	}
	got, err := pxc.String()
	if err != nil {
		t.Fatal(err)
	}
	if want != got {
		t.Errorf("want %q, but got %q", want, got)
	}

}

func TestStringError(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Err:    errors.New("some non-nil error"),
		Reader: strings.NewReader(""),
	}
	_, err := pxc.String()
	if err == nil {
		t.Error("want a error from String when PxColorPipe has an error, but got nil")
	}

}

func TestToJson(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader("b28d2f b6482d c0660f"),
	}
	pxc.ToJson()
	want := bluestacks.TriHexColors{
		FirstHexColor:  "b28d2f",
		SecondHexColor: "b6482d",
		ThirdHexColor:  "c0660f",
	}
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	got := bluestacks.TriHexColors{}
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestToJsonNoOpsWhenPxErrIsNonNil(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader("b28d2f b6482d c0660f"),
		Err:    errors.New("some non-nil error"),
	}
	pxc.ToJson()
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Errorf("want no output from ToJson when PxColorPipe has an error, but got %s", data)
	}
}
func TestToJsonInvalidInputReader(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader("b28d2f b6482d "),
	}
	pxc.ToJson()
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Errorf("want no output from ToJson when length of input reader is not 20, but got %s", data)
	}
}

// TODO: TDD a Post sink
