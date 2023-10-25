package mp3lib

import (
	"context"
	"golang.org/x/sync/errgroup"
	"io"
)

// CopyWithCancel is similar to io.Copy but can be cancelled using the provided context.
func CopyWithCancel(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	g, ctx_ := errgroup.WithContext(ctx)
	ctx, childCancel := context.WithCancel(ctx_)

	r, w := io.Pipe()

	// Goroutine to copy from src to PipeWriter
	g.Go(func() error {
		defer func(w *io.PipeWriter) {
			_ = w.Close()
		}(w)
		_, err := io.Copy(w, src)
		return err
	})

	var written int64 // variable to capture the number of bytes written to dst
	// Goroutine to copy from PipeReader to dst
	g.Go(func() error {
		defer childCancel()
		var err error
		defer func(r *io.PipeReader) {
			_ = r.Close()
		}(r)
		written, err = io.Copy(dst, r)
		return err
	})

	// Goroutine to listen for ctx cancellation and close the PipeWriter
	g.Go(func() error {
		<-ctx.Done()
		_ = w.CloseWithError(ctx.Err())
		return nil
	})

	// Wait for all tasks to complete or return on the first error.
	return written, g.Wait()
}
