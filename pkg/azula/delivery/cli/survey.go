package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func SurveyCheckboxes(label string, opts []string) []string {
	res := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	err := survey.AskOne(prompt, &res)
	if err != nil {
		if err == terminal.InterruptErr {
			fmt.Println("=> interrupted")
			os.Exit(0)
		}
	}

	return res
}

func SurveyList(label string, opts []string) string {
	res := ""
	prompt := &survey.Select{
		Message: label,
		Options: opts,
	}
	err := survey.AskOne(prompt, &res)
	if err != nil {
		if err == terminal.InterruptErr {
			fmt.Println("=> interrupted")
			os.Exit(0)
		}
	}
	return res
}
