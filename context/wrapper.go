package context

import (
	"context"
)

type Wrapper interface {
	Wrap(context.Context) context.Context
}
