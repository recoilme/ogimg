package og

import (
	"bufio"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "golang.org/x/image/webp"

	//"golang.org/x/image/webp"

	"github.com/stretchr/testify/assert"
)

func Test_Tguitarist(t *testing.T) {
	rc, err := fileReader("0.jpg")
	assert.NoError(t, err)
	defer rc.Close()
	bin, _, err := crop(rc, 1.9, "")
	assert.NoError(t, err)
	ioutil.WriteFile("./smartcrop.jpg", bin, 0666)
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

func fileReader(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bufio.NewReader(f)), nil
}
