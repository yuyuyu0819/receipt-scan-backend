package vision

import (
	cloudvision  "cloud.google.com/go/vision/v2/apiv1"
    visionpb "cloud.google.com/go/vision/v2/apiv1/visionpb"
	"context"
	"encoding/base64"
	"os"
	"strings"

	"receiptScan-backend/internal/domain/receipt"
	"receiptScan-backend/internal/usecase/ocr"
)

type client struct {
	visionClient *cloudvision.ImageAnnotatorClient
}

func NewClient(ctx context.Context) (ocr.OCRService, error) {
	// GOOGLE_APPLICATION_CREDENTIALS は .env で設定済みの前提
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil, ErrMissingCredentials
	}

	c, err := cloudvision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	return &client{visionClient: c}, nil
}

var ErrMissingCredentials = &ConfigError{Msg: "GOOGLE_APPLICATION_CREDENTIALS is not set"}

type ConfigError struct {
	Msg string
}

func (e *ConfigError) Error() string {
	return e.Msg
}

// OCRService 実装
func (c *client) DetectText(ctx context.Context, imageBase64 string) (receipt.OcrResult, error) {
	// data URI プレフィックスを剥がす
	if idx := strings.Index(imageBase64, ","); idx != -1 {
		imageBase64 = imageBase64[idx+1:]
	}

	imgBytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return receipt.OcrResult{}, err
	}

	image := &visionpb.Image{
		Content: imgBytes,
	}

	feature := &visionpb.Feature{
		Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION,
	}

	reqs := []*visionpb.AnnotateImageRequest{
		{
			Image: image,
			Features: []*visionpb.Feature{
				feature,
			},
		},
	}

	resp, err := c.visionClient.BatchAnnotateImages(ctx, &visionpb.BatchAnnotateImagesRequest{
		Requests: reqs,
	})
	if err != nil {
		return receipt.OcrResult{}, err
	}

	if len(resp.Responses) == 0 {
		return receipt.OcrResult{}, nil
	}

	res := resp.Responses[0]

	var text string
	if res.FullTextAnnotation != nil {
		text = res.FullTextAnnotation.Text
	} else if len(res.TextAnnotations) > 0 {
		text = res.TextAnnotations[0].Description
	}

	return receipt.OcrResult{
		RawText: text,
	}, nil
}
