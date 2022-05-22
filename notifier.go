package ueloghandler

import (
	"context"
)

type Notifier interface {
	Logs() chan string
	Subscribe(ctx context.Context) error
	Flush() error
}
