package cli

import (
	"os"

	"github.com/nikgalkin/azula/pkg/azula/usecase"

	"github.com/spf13/cobra"
)

type CliHandler interface {
	Execute()
}

type cli struct {
	UC usecase.ManUsecase
}

func New(uc usecase.ManUsecase) CliHandler {
	return &cli{
		UC: uc,
	}
}

var rootCmd = &cobra.Command{
	Use:   "azula",
	Short: "Manipulates with docker registry objects",
	Long: `Use environment variable AZULA_REGISTRY to pass registry address. By default http://127.0.0.1:5000
  example:
    export AZULA_REGISTRY=https://some-registry.domain.com`,
}

var (
	meta = &cli{}
)

const (
	mgmtBack = "<= back"
)

func (c *cli) Execute() {
	meta = c
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
