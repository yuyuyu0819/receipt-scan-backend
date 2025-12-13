package vision

import (
	"bytes"
	"errors"
	"image/jpeg"

	fitz "github.com/gen2brain/go-fitz/v2"
)

func looksLikePDF(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	if bytes.HasPrefix(data, []byte("%PDF")) {
		return true
	}

	return false
}

func convertPDFToJPEG(pdfBytes []byte) ([]byte, error) {
	doc, err := fitz.NewFromMemory(pdfBytes)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	img, err := doc.Image(0)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, errors.New("failed to convert pdf to jpeg: empty result")
	}

	return buf.Bytes(), nil
}
