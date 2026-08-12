package limiter

import (
	"context"
	"fmt"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/buf"
	"golang.org/x/time/rate"
)

type Writer struct {
	writer  buf.Writer
	limiter *rate.Limiter
	ctx     context.Context
}

func (l *Limiter) RateWriter(ctx context.Context, writer buf.Writer, limiter *rate.Limiter) buf.Writer {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Writer{
		writer:  writer,
		limiter: limiter,
		ctx:     ctx,
	}
}

func (w *Writer) Close() error {
	return common.Close(w.writer)
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	for !mb.IsEmpty() {
		burst := w.limiter.Burst()
		if burst <= 0 {
			buf.ReleaseMulti(mb)
			return fmt.Errorf("rate limiter burst must be positive")
		}

		chunkSize := int32(burst)
		if burst > int(^uint32(0)>>1) {
			chunkSize = int32(^uint32(0) >> 1)
		}
		remaining, chunk := buf.SplitSize(mb, chunkSize)
		mb = remaining
		if err := w.limiter.WaitN(w.ctx, int(chunk.Len())); err != nil {
			buf.ReleaseMulti(chunk)
			buf.ReleaseMulti(mb)
			return err
		}
		if err := w.writer.WriteMultiBuffer(chunk); err != nil {
			buf.ReleaseMulti(mb)
			return err
		}
	}
	return nil
}
