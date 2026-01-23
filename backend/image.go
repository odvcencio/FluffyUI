package backend

// ImageProtocol identifies the terminal image protocol.
type ImageProtocol int

const (
	ImageProtocolKitty ImageProtocol = iota
	ImageProtocolSixel
)

// ImageFormat identifies the pixel format.
type ImageFormat int

const (
	ImageFormatRGBA ImageFormat = iota
)

// Image describes a pixel image to be rendered.
type Image struct {
	Width      int
	Height     int
	CellWidth  int
	CellHeight int
	Format     ImageFormat
	Protocol   ImageProtocol
	Pixels     []byte
}

// ImageTarget captures image operations during rendering.
type ImageTarget interface {
	SetImage(x, y int, img Image)
}

// ImageWriter renders images to the terminal backend.
type ImageWriter interface {
	DrawImage(x, y int, img Image)
}
