package repo

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"github.com/disintegration/imaging"
)

const avatarSize = 300

// ResizeToAvatarJPEG decodes common image formats, crops to a centered 300×300 square, encodes JPEG.
func ResizeToAvatarJPEG(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("empty image")
	}
	img, err := imaging.Decode(bytes.NewReader(src), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	thumb := imaging.Fill(img, avatarSize, avatarSize, imaging.Center, imaging.Lanczos)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
