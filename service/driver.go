package service

import "context"

type Driver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
