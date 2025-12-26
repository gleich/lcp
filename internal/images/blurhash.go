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
		return "", fmt.Errorf("decoding image failed: %w", err)
	}

	width := parsedImage.Bounds().Dx()
	height := parsedImage.Bounds().Dy()
	blurData, err := blurhash.Encode(4, 3, parsedImage)
	if err != nil {
		return "", fmt.Errorf("encoding image into blurhash: %w", err)
	}

	scaleDownFactor := 25
	blurImage, err := blurhash.Decode(blurData, width/scaleDownFactor, height/scaleDownFactor, 1)
	if err != nil {
		return "", fmt.Errorf("decoding blurhash data into image: %w", err)
	}
	blurImageBuffer := new(bytes.Buffer)
	err = png.Encode(blurImageBuffer, blurImage)
	if err != nil {
		return "", fmt.Errorf("creating png based on blurred image: %w", err)
	}
	return fmt.Sprintf(
		"data:image/png;base64,%s",
		base64.StdEncoding.EncodeToString(blurImageBuffer.Bytes()),
	), nil
}
