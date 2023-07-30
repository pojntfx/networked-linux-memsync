package utils

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

func IsClosedErr(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.ECONNRESET) || strings.HasSuffix(err.Error(), "read: connection timed out") || strings.HasSuffix(err.Error(), "write: broken pipe") || strings.HasSuffix(err.Error(), "unexpected EOF") {
		return true
	}

	return false
}
