package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/buckket/go-blurhash"
	"pkg.mattglei.ch/timber"
)

func BlurImage(data []byte, decoder func(r io.Reader) (image.Image, error)) []byte {
	reader := bytes.NewReader(data)
	parsedImage, err := decoder(reader)
	if err != nil {
		timber.Error(err, "decoding image failed")
		return nil
	}

	width := parsedImage.Bounds().Dx()
	height := parsedImage.Bounds().Dy()
	blurData, err := blurhash.Encode(4, 3, parsedImage)
	if err != nil {
		timber.Error(err, "encoding image into blurhash failed")
		return nil
	}

	scaleDownFactor := 25
	blurImage, err := blurhash.Decode(blurData, width/scaleDownFactor, height/scaleDownFactor, 1)
	if err != nil {
		timber.Error(err, "decoding blurhash data into img failed")
		return nil
	}
	blurImageBuffer := new(bytes.Buffer)
	err = png.Encode(blurImageBuffer, blurImage)
	if err != nil {
		timber.Error(err, "creating png based off blurred image failed")
		return nil
	}
	return blurImageBuffer.Bytes()
}

func BlurDataURI(data []byte) string {
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(data))
}
