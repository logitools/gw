package uds

import (
	"io"
)

type CommandHandler interface {
	Command() string   // Unique Name
	GroupName() string // for display grouping
	Desc() string
	Usage() string
	HandleCommand(args []string, w io.Writer) error
}
