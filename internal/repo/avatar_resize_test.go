package repo

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/disintegration/imaging"
)

func TestResizeToAvatarJPEG_squareOutput(t *testing.T) {
	src := imaging.New(800, 400, color.NRGBA{R: 40, G: 120, B: 200, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	jpegBytes, err := ResizeToAvatarJPEG(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		t.Fatal(err)
	}
	b := decoded.Bounds()
	if b.Dx() != avatarSize || b.Dy() != avatarSize {
		t.Fatalf("expected %dx%d, got %dx%d", avatarSize, avatarSize, b.Dx(), b.Dy())
	}
}

func TestResizeToAvatarJPEG_emptyInput(t *testing.T) {
	if _, err := ResizeToAvatarJPEG(nil); err == nil {
		t.Fatal("expected error")
	}
}
