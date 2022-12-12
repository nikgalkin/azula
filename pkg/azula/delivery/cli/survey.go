package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

const (
	surveyHelp = "ctrl+c to exit"
)

func SurveyCheckboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: labelWithCount(label, opts),
		Options: opts,
		Help:    surveyHelp,
	}
	surveyCheckErr(survey.AskOne(prompt, &res))

	return res
}

func SurveyList(label string, opts []string) string {
	res := ""
	prompt := &survey.Select{
		Message: labelWithCount(label, opts),
		Options: opts,
		Help:    surveyHelp,
	}
	surveyCheckErr(survey.AskOne(prompt, &res))
	return res
}

func labelWithCount(label string, opts []string) string {
	return fmt.Sprintf("%s(%d)", label, len(opts))
}

func surveyCheckErr(err error) {
	if err != nil {
		if err == terminal.InterruptErr {
			fmt.Println("=> interrupted")
			os.Exit(0)
		}
	}
}
