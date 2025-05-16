package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/buckket/go-blurhash"
)

type ImageDecoder func(r io.Reader) (image.Image, error)

func blur(data []byte, decoder ImageDecoder) (string, error) {
	reader := bytes.NewReader(data)
	parsedImage, err := decoder(reader)
	if err != nil {
		return "", fmt.Errorf("%w decoding image failed", err)
	}

	width := parsedImage.Bounds().Dx()
	height := parsedImage.Bounds().Dy()
	blurData, err := blurhash.Encode(4, 3, parsedImage)
	if err != nil {
		return "", fmt.Errorf("%w encoding image into blurhash failed", err)
	}

	scaleDownFactor := 25
	blurImage, err := blurhash.Decode(blurData, width/scaleDownFactor, height/scaleDownFactor, 1)
	if err != nil {
		return "", fmt.Errorf("%w decoding blurhash data into img failed", err)
	}
	blurImageBuffer := new(bytes.Buffer)
	err = png.Encode(blurImageBuffer, blurImage)
	if err != nil {
		return "", fmt.Errorf("%w creating png based off blurred image failed", err)
	}
	return fmt.Sprintf(
		"data:image/png;base64,%s",
		base64.StdEncoding.EncodeToString(blurImageBuffer.Bytes()),
	), nil
}
