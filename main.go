package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
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

func PayloadToPayload(p Payload) (Payload, error) {
	reader := p.Content.ByteReader()

	m, format, err := image.Decode(reader)
	if err != nil {
		return p, err
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
		w.WriteHeader(200)
		return
	}

	var p Payload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	new_payload, err := PayloadToPayload(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc.Encode(new_payload)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<a href='https://github.com/revisitors'>You know me?</a>")
	})
	http.HandleFunc("/rotate/", workIt)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Listening on " + port)
	http.ListenAndServe(":"+port, nil)
}
