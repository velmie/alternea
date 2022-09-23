package manipulation

import (
	"context"
	"io"

	"github.com/velmie/alternea/dframe"
)

type Tablifier interface {
	Table(in []byte) (*dframe.Table, error)
}

type Remapper interface {
	Remap(in []byte) ([]byte, error)
}

type (
	// DataTransformer transforms input data and writes it to the given writer
	DataTransformer interface {
		Transform(ctx context.Context, pages <-chan []byte, w io.Writer) error
	}
)
