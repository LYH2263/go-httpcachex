package httpcachex

import (
	"context"
	"fmt"
)

func wrapUpstream(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("upstream failed: %v", err)
}

func wrapCancel(err error) error {
	if err == nil {
		return nil
	}
	if err == context.Canceled || err == context.DeadlineExceeded {
		return fmt.Errorf("%w: %v", ErrCanceled, err)
	}
	return err
}

func wrapPersist(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrPersist, err)
}
