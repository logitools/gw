package contxt

import (
	"context"
)

type InjectorFunc[T any] func(context.Context, T) (context.Context, error)
