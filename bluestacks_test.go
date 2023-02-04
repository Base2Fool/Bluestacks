package bluestacks_test

import (
	"bluestacks"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"image/color"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
		First:  "b28d2f",
		Second: "b6482d",
		Third:  "c0660f",
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

func TestOpacityIsFilteringTheMostOpaqueColors(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{133, 190, 198, 255}, {100, 170, 180, 255}, {125, 180, 190, 255}},
				{{232, 141, 52, 255}, {212, 121, 32, 255}, {202, 111, 22, 255}},
				{{100, 168, 23, 255}, {130, 178, 33, 255}, {140, 188, 37, 255}},
			},
		},
	}
	want := bluestacks.OpaqueColors{
		Colors: []color.RGBA{
			{100, 170, 180, 255},
			{202, 111, 22, 255},
			{100, 168, 23, 255}},
	}
	pxc.Opacity()
	got := pxc.Opaques
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestOpacityNoOpsWhenPxColorPipeHasAnError(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Err: errors.New("some non-nil error"),
		Colors: bluestacks.Colors{
			RGBAs: [][]color.RGBA{
				{{133, 190, 198, 255}, {100, 170, 180, 255}, {125, 180, 190, 255}},
				{{232, 141, 52, 255}, {212, 121, 32, 255}, {202, 111, 22, 255}},
				{{100, 168, 23, 255}, {130, 178, 33, 255}, {140, 188, 37, 255}},
			},
		},
	}
	pxc.Opacity()
	want := bluestacks.OpaqueColors{Colors: []color.RGBA{}}
	got := pxc.Opaques
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}

}

func TestOpaqueToHexIsFilteringTheRGBAsOfOpaquesToHexadecimal(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Reader: strings.NewReader(" "),
		Opaques: bluestacks.OpaqueColors{
			Colors: []color.RGBA{
				{100, 170, 180, 255},
				{202, 111, 22, 255},
				{100, 168, 23, 255}},
		},
	}
	pxc.OpaqueToHex()
	want := "64aab4 ca6f16 64a817"
	got, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, string(got)) {
		t.Error(cmp.Diff(want, string(got)))
	}
}

// TODO: TDD OpaqueToHex NoOP & Errors if any
func TestOpaqueToHexNoOpsWhenPxColorPipeHasAnError(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Err:     errors.New("some non-nil error"),
		Opaques: bluestacks.OpaqueColors{Colors: []color.RGBA{}},
	}
	pxc.OpaqueToHex()
	data, err := io.ReadAll(pxc.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Errorf("want no output from OpaqueToHex when PxColourPipe has an err, but got %v", data)
	}

}

func TestPatchIsPostingColorsToEndPoint(t *testing.T) {
	t.Parallel()
	pxc := bluestacks.PxColorPipe{
		Opaques: bluestacks.OpaqueColors{
			Colors: []color.RGBA{
				{255, 251, 187, 255},
				{19, 171, 19, 255},
				{255, 255, 255, 255}},
		},
	}
	h1 := func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		cFromReq := bluestacks.TriHexColors{}
		err = json.Unmarshal(data, &cFromReq)
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open("testdata/color_record_update.json")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		fileData, err := io.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		cFromFile := bluestacks.TriHexColors{}
		err = json.Unmarshal(fileData, &cFromFile)
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(cFromFile, cFromReq) {
			t.Error(cmp.Diff(cFromFile, cFromReq))
		}
		if cmp.Equal(cFromFile, cFromReq) {
			fmt.Fprintf(w, "Colors Match")
		}
	}
	ts := httptest.NewTLSServer(http.HandlerFunc(h1))
	defer ts.Close()
	pxc.HttpClient = ts.Client()
	resp, err := pxc.OpaqueToHex().ToJson().Patch(ts.URL)
	if err != nil {
		t.Fatal()
	}
	want := "Colors Match"
	if resp.StatusCode != http.StatusOK {
		t.Fatal()
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal()
	}
	got := string(data)
	if want != got {
		t.Errorf("want %q, but got %q", want, got)
	}

}
