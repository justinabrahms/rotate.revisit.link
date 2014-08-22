package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Image struct {
	Type string
	Data string // url?
}

func (i Image) ByteReader() io.Reader {
	dataUri := i.Data
	data := strings.Split(dataUri, ",")[1]

	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
}

type Audio struct {
	Type string
	Data string // url?
}

type Payload struct {
	Content Image
	Meta    struct {
		Audio Audio
	}
}

// Rotates an image 90deg to the left.
func rotate(m image.Image, dst image.RGBA) {
	orig := m.Bounds()

	for x := orig.Min.X; x < orig.Max.X; x++ {
		for y := orig.Min.Y; y < orig.Max.Y; y++ {
			dst.Set(y,
				orig.Max.Y-x+1,
				m.At(x, y))
		}
	}
}

func PayloadToPayload(p Payload) Payload {
	reader := p.Content.ByteReader()

	m, format, err := image.Decode(reader)
	if err != nil {
		log.Fatal("Couldn't decode: ", err)
	}

	img := image.NewRGBA(m.Bounds())
	rotate(m, *img)

	buf := bytes.NewBuffer(nil)
	err = jpeg.Encode(buf, img, nil)
	if err != nil {
		log.Fatal(err)
	}

	z := base64.StdEncoding.EncodeToString(buf.Bytes())

	return Payload{
		Content: Image{
			Type: fmt.Sprintf("image/%s", format),
			Data: fmt.Sprintf("data:image/%s;base64,%s", format, z),
		},
	}
}

func workIt(w http.ResponseWriter, r *http.Request) {
	var p Payload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		// TODO(justinabrahms): Make more resiliant.
		log.Fatal("Couldn't decode from json: ", err)
	}

	enc := json.NewEncoder(w)
	enc.Encode(PayloadToPayload(p))
}

func main() {
	http.HandleFunc("/", workIt)
	port := os.Getenv("PORT")
	if port == nil {
		port = "8080"
	}
	fmt.Println("Listening on " + port)
	http.ListenAndServe(":"+port, nil)
}
