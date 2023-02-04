package main

import (
	"bluestacks"
	"log"
	"net/http"
)

const BaseURL = "http://127.0.0.1:8090"
const Endpoint = "/api/collections/colors/records/dmevgew9jk9u1ow"

func main() {
	////bluestacks.RunCLI()
	px := bluestacks.NewPxColorPipe()
	for i := 0; i < 4; i++ {
		resp, err := px.Hex().HexToRGBA().Opacity().OpaqueToHex().ToJson().Patch(BaseURL + Endpoint)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Println(resp.StatusCode)
		}
	}
}
