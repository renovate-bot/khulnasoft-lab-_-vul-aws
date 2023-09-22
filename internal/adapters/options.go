package adapters

import (
	"github.com/khulnasoft-lab/defsec/pkg/debug"

	"github.com/khulnasoft-lab/vul-aws/pkg/concurrency"
	"github.com/khulnasoft-lab/vul-aws/pkg/progress"
)

type Options struct {
	ProgressTracker     progress.Tracker
	Region              string
	Endpoint            string
	Services            []string
	DebugWriter         debug.Logger
	ConcurrencyStrategy concurrency.Strategy
}
