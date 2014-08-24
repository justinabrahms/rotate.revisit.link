package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

type Image struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (i Image) ByteReader() io.Reader {
	dataUri := i.Data
	data := strings.Split(dataUri, ",")[1]

	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
}

type Audio struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type Meta struct {
	Audio Audio `json:"audio"`
}

type Payload struct {
	Content Image `json:"content"`
	Meta    Meta  `json:"meta"`
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

func drawFutzery(m image.Image, dst image.RGBA) {
	// TODO(justinabrahms): randomize
	point := image.Point{0, 0}

	draw.Draw(m, m.Bounds(), dst, point, draw.Src)
}

func PayloadToPayload(p Payload) (Payload, error) {
	reader := p.Content.ByteReader()

	m, format, err := image.Decode(reader)
	if err != nil {
		return p, err
	}

	// Not sure how to handle non-jpegs yet.
	if format != "jpeg" {
		log.Printf("We only know how to decode jpegs, not %s", format)
		return p, nil
	}

	img := image.NewRGBA(m.Bounds())
	rotate(m, *img)

	buf := bytes.NewBuffer(nil)
	err = jpeg.Encode(buf, img, nil)
	if err != nil {
		return p, err
	}

	z := base64.StdEncoding.EncodeToString(buf.Bytes())

	return Payload{
		Content: Image{
			Type: fmt.Sprintf("image/%s", format),
			Data: fmt.Sprintf("data:image/%s;base64,%s", format, z),
		},
	}, nil
}

func workIt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Println("Health check")
		w.WriteHeader(200)
		return
	}

	var p Payload
	enc := json.NewEncoder(w)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		log.Printf("ERROR: %s", err)
		enc.Encode(p)
		return
	}

	new_payload, err := PayloadToPayload(p)
	if err != nil {
		log.Printf("ERROR: %s", err)
		enc.Encode(p)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc.Encode(new_payload)
}

func main() {
	rand.Seed(1)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<a href='https://github.com/revisitors'>You know me?</a>")
	})
	http.HandleFunc("/service", workIt)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Listening on " + port)
	http.ListenAndServe(":"+port, nil)
}
