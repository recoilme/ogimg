package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

var (
	//errors
	errURLLoad  = "Error: load url"
	aspectRatio = 1.9
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := html.EscapeString(r.URL.Path)
	log("mainPage", path)
	switch r.Method {
	case "GET":
		vals := r.URL.Query()
		url := vals.Get("url")
		if path == "/" && url == "" {
			w.WriteHeader(200)
			return
		}
		//res, err := og.GetOpenGraphFromUrl("https://hackernoon.com/golang-clean-archithecture-efd6d7c43047")

		bin, cntType, err := imgLoad(url)
		if err != nil {
			log("imgLoad err", err)
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", cntType)
		w.Write(bin)
		return
	default:
		log("wrong params")
		w.WriteHeader(503)
		return
	}
}

func log(a ...interface{}) (n int, err error) {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s ", time.Now().Format("15:04:05")))
	isError := false
	for _, s := range a {
		buf.WriteString(fmt.Sprint(s, " "))
		if strings.HasPrefix(reflect.TypeOf(a).String(), "*error") {
			isError = true
		}
	}
	if isError {
		fmt.Println(os.Stderr, buf.String())
	}
	return fmt.Fprintln(os.Stdout, buf.String())
}

// imgLoad load image data
func imgLoad(imgURL string) ([]byte, string, error) {
	log("imgLoad", imgURL)
	limitRead := 512
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
	defer cncl()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		//log.Println(err, imgUrl)
		return nil, "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log(err, imgURL)
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		fmt.Println(err, imgURL)
		return nil, "", errors.New(errURLLoad)
	}

	head, err := ioutil.ReadAll(io.LimitReader(resp.Body, int64(limitRead)))
	if err != nil {
		fmt.Println(err, imgURL)
		return nil, "", err
	}

	cntType := strings.ToLower(http.DetectContentType(head))
	//log.Println(cntType)
	if !strings.HasPrefix(cntType, "image") {
		return nil, "", errors.New(errURLLoad)
	}

	return crop(resp.Body, aspectRatio, cntType)
}

func crop(rc io.ReadCloser, ar float64, cntType string) ([]byte, string, error) {
	log("crop")
	img, _, err := image.Decode(rc)
	if err != nil {
		return nil, "", err
	}

	// ищем самый интересный квадрат
	min := img.Bounds().Max.X
	vertical := true
	if img.Bounds().Max.Y < min {
		min = img.Bounds().Max.Y
		vertical = false
	}
	_ = vertical
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	rect, err := analyzer.FindBestCrop(img, min, min)

	// берем от топ до ширина/отношение_сторон
	rect.Max.Y = int(float64(min)/ar) + rect.Min.Y

	sub, ok := img.(SubImager)
	if ok {
		cropImage := sub.SubImage(rect)
		buf := new(bytes.Buffer)
		//TODO CONVERT BY CONTENT TYPE
		jpeg.Encode(buf, cropImage, &jpeg.Options{Quality: 95})
		return buf.Bytes(), "image/jpeg", nil
	}
	return nil, "", errors.New("No SubImage support")
}
