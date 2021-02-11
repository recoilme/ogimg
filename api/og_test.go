package og

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/webp"

	//"golang.org/x/image/webp"
	"github.com/recoilme/smartcrop"
	"github.com/recoilme/smartcrop/nfnt"
	"github.com/stretchr/testify/assert"
)

func Test_Tguitarist(t *testing.T) {

	bin, err := ioutil.ReadFile("4.webp")
	assert.NoError(t, err)
	img, _, err := image.Decode(bytes.NewReader(bin))
	assert.NoError(t, err)

	//fmt.Printf("%d %d\n", img.Bounds().Max.X, img.Bounds().Max.Y)
	// ищем самый интересный квадрат
	min := img.Bounds().Max.X
	if img.Bounds().Max.Y < min {
		min = img.Bounds().Max.Y
	}
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	rect, err := analyzer.FindBestCrop(img, min, min)

	// берем от топ до ширина/отношение_сторон
	rect.Max.Y = int(float64(rect.Max.X)/1.9) + rect.Min.Y
	//fmt.Printf("%d %s\n", rect.Min.Y, rect.String())
	sub, ok := img.(SubImager)
	if ok {
		cropImage := sub.SubImage(rect)
		err = writeImage("jpeg", cropImage, "./smartcrop.jpg")
		assert.NoError(t, err)
	} else {
		t.Error(errors.New("No SubImage support"))
	}
}

func writeImage(imgtype string, img image.Image, name string) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		panic(err)
	}

	switch imgtype {
	case "png":
		return writeImageToPng(img, name)
	case "jpeg":
		return writeImageToJpeg(img, name)
	}

	return errors.New("Unknown image type")
}

func writeImageToJpeg(img image.Image, name string) error {
	fso, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fso.Close()

	return jpeg.Encode(fso, img, &jpeg.Options{Quality: 100})
}

func writeImageToPng(img image.Image, name string) error {
	fso, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fso.Close()

	return png.Encode(fso, img)
}
