package app

import (
	"io"

	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

const unsetOptionValue = "__unset__"

func RunBaseConfigWizard(in io.Reader, out io.Writer, language string) (Config, error) {
	return runBaseConfigWizardWithPrompter(newWizardPrompter(in, out), language)
}

func runBaseConfigWizardWithPrompter(prompter wizardPrompter, language string) (Config, error) {
	cfg := Config{}
	promptLanguage := i18n.NormalizeLanguage(language)
	prompter.SetLanguage(promptLanguage)

	selectedLanguage, err := askOptionalLanguage(prompter, promptLanguage)
	if err != nil {
		return cfg, err
	}
	cfg.Language = selectedLanguage
	if cfg.Language != "" {
		prompter.SetLanguage(cfg.Language)
		promptLanguage = cfg.Language
	}

	mode, err := askOptionalMappedSelect(prompter, i18n.T(promptLanguage, "wizard.mode"), optionalModeOptions(promptLanguage))
	if err != nil {
		return cfg, err
	}
	switch mode {
	case "single":
		cfg.RepoPath, err = prompter.Input(i18n.T(promptLanguage, "wizard.repo"), "", false)
		if err != nil {
			return cfg, err
		}
	case "multi":
		cfg.ReposPath, err = prompter.Input(i18n.T(promptLanguage, "wizard.repos"), "", false)
		if err != nil {
			return cfg, err
		}
	}

	cfg.Time, err = askOptionalTimeSpec(prompter, promptLanguage)
	if err != nil {
		return cfg, err
	}

	cfg.Format, err = askOptionalMappedSelect(prompter, i18n.T(promptLanguage, "wizard.format"), optionalFormatOptions(promptLanguage))
	if err != nil {
		return cfg, err
	}
	cfg.Prompt, err = askOptionalPrompt(prompter, promptLanguage)
	if err != nil {
		return cfg, err
	}
	cfg.Author, err = prompter.Input(i18n.T(promptLanguage, "wizard.author"), "", false)
	if err != nil {
		return cfg, err
	}
	cfg.OutputFile, err = prompter.Input(i18n.T(promptLanguage, "wizard.output"), "", false)
	if err != nil {
		return cfg, err
	}
	cfg.Provider, err = askOptionalMappedSelect(prompter, i18n.T(promptLanguage, "wizard.provider"), optionalProviderOptions(promptLanguage))
	if err != nil {
		return cfg, err
	}
	cfg.BaseURL, err = prompter.Input(i18n.T(promptLanguage, "wizard.base_url"), "", false)
	if err != nil {
		return cfg, err
	}
	cfg.APIKey, err = prompter.Input(i18n.T(promptLanguage, "wizard.api_key"), "", true)
	if err != nil {
		return cfg, err
	}
	cfg.Model, err = prompter.Input(i18n.T(promptLanguage, "wizard.model"), "", false)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func askOptionalLanguage(prompter wizardPrompter, language string) (string, error) {
	return askOptionalMappedSelect(prompter, "Language（语言）", optionalLanguageOptions(language))
}

func askOptionalPrompt(prompter wizardPrompter, language string) (string, error) {
	choice, err := askOptionalMappedSelect(prompter, i18n.T(language, "wizard.prompt"), optionalPromptOptions(language))
	if err != nil {
		return "", err
	}
	if choice == "custom-file" {
		return prompter.Input(i18n.T(language, "wizard.prompt.custom_path"), "", false)
	}
	return choice, nil
}

func askOptionalTimeSpec(prompter wizardPrompter, language string) (timequery.Spec, error) {
	choice, err := askOptionalMappedSelect(prompter, i18n.T(language, "wizard.time_type"), optionalTimeTypeOptions(language))
	if err != nil {
		return timequery.Spec{}, err
	}

	switch choice {
	case "":
		return timequery.Spec{}, nil
	case "preset":
		period, err := askOptionalMappedSelect(prompter, i18n.T(language, "wizard.period"), optionalPeriodOptions(language))
		if err != nil {
			return timequery.Spec{}, err
		}
		if period == "" {
			return timequery.Spec{}, nil
		}
		return timequery.Spec{Kind: timequery.KindPreset, Period: period}, nil
	case "day":
		on, err := prompter.Input(i18n.T(language, "wizard.on"), "", false)
		if err != nil {
			return timequery.Spec{}, err
		}
		if on == "" {
			return timequery.Spec{}, nil
		}
		return timequery.Spec{Kind: timequery.KindSingleDay, On: on}, nil
	case "range":
		from, err := prompter.Input(i18n.T(language, "wizard.from"), "", false)
		if err != nil {
			return timequery.Spec{}, err
		}
		to, err := prompter.Input(i18n.T(language, "wizard.to"), "", false)
		if err != nil {
			return timequery.Spec{}, err
		}
		if from == "" || to == "" {
			return timequery.Spec{}, nil
		}
		return timequery.Spec{Kind: timequery.KindRange, From: from, To: to}, nil
	default:
		return timequery.Spec{}, nil
	}
}

func askOptionalMappedSelect(prompter wizardPrompter, label string, options []selectOption) (string, error) {
	choice, err := prompter.Select(label, optionLabels(options), optionLabelByValue(options, unsetOptionValue))
	if err != nil {
		return "", err
	}
	selected := optionValueByLabel(options, choice)
	if selected == unsetOptionValue {
		return "", nil
	}
	return selected, nil
}

func optionalLanguageOptions(language string) []selectOption {
	return prependUnsetOption(language, languageOptions)
}

func optionalModeOptions(language string) []selectOption {
	return prependUnsetOption(language, localizedModeOptions(language))
}

func optionalTimeTypeOptions(language string) []selectOption {
	return prependUnsetOption(language, localizedTimeTypeOptions(language))
}

func optionalPeriodOptions(language string) []selectOption {
	return prependUnsetOption(language, localizedPeriodOptions(language))
}

func optionalFormatOptions(language string) []selectOption {
	return prependUnsetOption(language, localizedFormatOptions(language))
}

func optionalPromptOptions(language string) []selectOption {
	return prependUnsetOption(language, localizedPromptOptions(language))
}

func optionalProviderOptions(language string) []selectOption {
	return prependUnsetOption(language, providerOptions)
}

func prependUnsetOption(language string, options []selectOption) []selectOption {
	withUnset := make([]selectOption, 0, len(options)+1)
	withUnset = append(withUnset, selectOption{
		Label: i18n.T(language, "wizard.unset"),
		Value: unsetOptionValue,
	})
	withUnset = append(withUnset, options...)
	return withUnset
}
