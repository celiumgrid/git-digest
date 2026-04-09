package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/celiumgrid/git-digest/internal/ai"
	"github.com/celiumgrid/git-digest/internal/i18n"
	"github.com/celiumgrid/git-digest/internal/timequery"
)

const selectQuestionTemplateBase = `
{{- define "option"}}
    {{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }} {{else}}{{color "default"}}  {{end}}
    {{- .CurrentOpt.Value}}{{ if ne ($.GetDescription .CurrentOpt) "" }} - {{color "cyan"}}{{ $.GetDescription .CurrentOpt }}{{end}}
    {{- color "reset"}}
{{end}}
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{ .FilterMessage }}{{color "reset"}}
{{- if .ShowAnswer}}{{color "cyan"}} {{.Answer}}{{color "reset"}}{{"\n"}}
{{- else}}
  {{- "\n"}}
  {{- "  "}}{{- color "cyan"}}[__HELP__]{{color "reset"}}
  {{- "\n"}}
  {{- range $ix, $option := .PageEntries}}
    {{- template "option" $.IterateOption $ix $option}}
  {{- end}}
{{- end}}`

type wizardPrompter interface {
	SetLanguage(language string)
	Select(label string, options []string, defaultValue string) (string, error)
	Input(label, defaultValue string, secret bool) (string, error)
	Confirm(label string, defaultValue bool) (bool, error)
}

type surveyPrompter struct {
	stdio survey.AskOpt
	lang  string
}

type linePrompter struct {
	reader *bufio.Reader
	out    io.Writer
	lang   string
}

type selectOption struct {
	Label string
	Value string
}

var providerOptions = []selectOption{
	{Label: "OpenAI", Value: ai.ProviderOpenAI},
	{Label: "Gemini", Value: ai.ProviderGemini},
	{Label: "DeepSeek", Value: ai.ProviderDeepSeek},
}

var languageOptions = []selectOption{
	{Label: "English", Value: timequery.LanguageEnglish},
	{Label: "中文", Value: timequery.LanguageChinese},
}

var modelOptionsByProvider = map[string][]selectOption{
	ai.ProviderOpenAI: {
		{Label: "gpt-4.1-mini", Value: "gpt-4.1-mini"},
		{Label: "gpt-4.1", Value: "gpt-4.1"},
		{Label: "gpt-4o-mini", Value: "gpt-4o-mini"},
		{Label: "custom-model", Value: "custom-model"},
	},
	ai.ProviderGemini: {
		{Label: "gemini-2.5-pro", Value: "gemini-2.5-pro"},
		{Label: "gemini-2.5-flash", Value: "gemini-2.5-flash"},
		{Label: "custom-model", Value: "custom-model"},
	},
	ai.ProviderDeepSeek: {
		{Label: "deepseek-chat", Value: "deepseek-chat"},
		{Label: "deepseek-reasoner", Value: "deepseek-reasoner"},
		{Label: "custom-model", Value: "custom-model"},
	},
}

func newWizardPrompter(in io.Reader, out io.Writer) wizardPrompter {
	inFile, inOK := in.(*os.File)
	outFile, outOK := out.(*os.File)
	if inOK && outOK {
		return &surveyPrompter{stdio: survey.WithStdio(inFile, outFile, outFile), lang: i18n.LanguageEnglish}
	}
	return &linePrompter{reader: bufio.NewReader(in), out: out, lang: i18n.LanguageEnglish}
}

func (p *surveyPrompter) SetLanguage(language string) {
	p.lang = i18n.NormalizeLanguage(language)
}

func (p *linePrompter) SetLanguage(language string) {
	p.lang = i18n.NormalizeLanguage(language)
}

func (p *surveyPrompter) Select(label string, options []string, defaultValue string) (string, error) {
	survey.SelectQuestionTemplate = localizedSelectQuestionTemplate(p.lang)
	prompt := &survey.Select{Message: label, Options: options, Default: defaultValue}
	var answer string
	err := survey.AskOne(prompt, &answer, p.stdio)
	return answer, err
}

func (p *surveyPrompter) Input(label, defaultValue string, secret bool) (string, error) {
	var prompt survey.Prompt
	if secret {
		prompt = &survey.Password{Message: label}
	} else {
		prompt = &survey.Input{Message: label, Default: defaultValue}
	}
	var answer string
	err := survey.AskOne(prompt, &answer, p.stdio)
	if err != nil {
		return "", err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue, nil
	}
	return answer, nil
}

func (p *surveyPrompter) Confirm(label string, defaultValue bool) (bool, error) {
	prompt := &survey.Confirm{Message: label, Default: defaultValue}
	var answer bool
	err := survey.AskOne(prompt, &answer, p.stdio)
	return answer, err
}

func (p *linePrompter) Select(label string, options []string, defaultValue string) (string, error) {
	return p.Input(fmt.Sprintf("%s [%s]", label, strings.Join(options, "/")), defaultValue, false)
}

func (p *linePrompter) Input(label, defaultValue string, _ bool) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.out, i18n.T(p.lang, "wizard.line.default"), label, defaultValue)
	} else {
		fmt.Fprintf(p.out, i18n.T(p.lang, "wizard.line.prompt"), label)
	}
	line, err := p.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

func (p *linePrompter) Confirm(label string, defaultValue bool) (bool, error) {
	defaultText := "N"
	if defaultValue {
		defaultText = "Y"
	}
	value, err := p.Input(label+" [y/N]", defaultText, false)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(value), "y"), nil
}

func RunWizard(in io.Reader, out io.Writer, base Config) (Config, error) {
	cfg, err := runWizardWithPrompter(newWizardPrompter(in, out), base)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func runWizardWithPrompter(prompter wizardPrompter, base Config) (Config, error) {
	cfg := base
	var err error

	cfg.Language, err = askLanguage(prompter, cfg.Language)
	if err != nil {
		return cfg, err
	}
	prompter.SetLanguage(cfg.Language)

	modeOptions := localizedModeOptions(cfg.Language)
	mode, err := prompter.Select(i18n.T(cfg.Language, "wizard.mode"), optionLabels(modeOptions), optionLabelByValue(modeOptions, chooseDefault(cfg.ReposPath != "", "multi", "single")))
	if err != nil {
		return cfg, err
	}
	modeValue := optionValueByLabel(modeOptions, mode)
	if modeValue == "multi" {
		cfg.RepoPath = ""
		cfg.ReposPath, err = prompter.Input(i18n.T(cfg.Language, "wizard.repos"), chooseDefault(cfg.ReposPath != "", cfg.ReposPath, "."), false)
		if err != nil {
			return cfg, err
		}
	} else {
		cfg.ReposPath = ""
		cfg.RepoPath, err = prompter.Input(i18n.T(cfg.Language, "wizard.repo"), chooseDefault(cfg.RepoPath != "", cfg.RepoPath, "."), false)
		if err != nil {
			return cfg, err
		}
	}

	cfg.Time, err = askTimeSpec(prompter, cfg.Time, cfg.Language)
	if err != nil {
		return cfg, err
	}

	formatOptions := localizedFormatOptions(cfg.Language)
	cfg.Format, err = askMappedSelect(prompter, i18n.T(cfg.Language, "wizard.format"), formatOptions, chooseDefault(cfg.Format != "", cfg.Format, "text"))
	if err != nil {
		return cfg, err
	}
	cfg.Prompt, err = askPrompt(prompter, cfg.Prompt, cfg.Language)
	if err != nil {
		return cfg, err
	}
	cfg.Author, err = prompter.Input(i18n.T(cfg.Language, "wizard.author"), cfg.Author, false)
	if err != nil {
		return cfg, err
	}
	cfg.OutputFile, err = prompter.Input(i18n.T(cfg.Language, "wizard.output"), cfg.OutputFile, false)
	if err != nil {
		return cfg, err
	}

	previousProvider := cfg.Provider
	cfg.Provider, err = askProvider(prompter, chooseDefault(cfg.Provider != "", cfg.Provider, ai.DefaultProvider), cfg.Language)
	if err != nil {
		return cfg, err
	}

	baseURLDefault := providerAwareValue(previousProvider, cfg.Provider, cfg.BaseURL, ai.DefaultBaseURL)
	cfg.BaseURL, err = prompter.Input(i18n.T(cfg.Language, "wizard.base_url"), baseURLDefault, false)
	if err != nil {
		return cfg, err
	}
	cfg.APIKey, err = prompter.Input(i18n.T(cfg.Language, "wizard.api_key"), cfg.APIKey, true)
	if err != nil {
		return cfg, err
	}
	modelDefault := providerAwareValue(previousProvider, cfg.Provider, cfg.Model, ai.DefaultModelName)
	cfg.Model, err = askModel(prompter, cfg.Provider, modelDefault, cfg.Language)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func askTimeSpec(prompter wizardPrompter, current timequery.Spec, language string) (timequery.Spec, error) {
	typeOptions := localizedTimeTypeOptions(language)
	choice, err := prompter.Select(i18n.T(language, "wizard.time_type"), optionLabels(typeOptions), optionLabelByValue(typeOptions, defaultTimeType(current)))
	if err != nil {
		return current, err
	}
	choice = optionValueByLabel(typeOptions, choice)

	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "preset":
		periodOptions := localizedPeriodOptions(language)
		period, err := askMappedSelect(prompter, i18n.T(language, "wizard.period"), periodOptions, defaultPeriod(current))
		if err != nil {
			return current, err
		}
		return timequery.Spec{Kind: timequery.KindPreset, Period: period}, nil
	case "day":
		on, err := prompter.Input(i18n.T(language, "wizard.on"), current.On, false)
		if err != nil {
			return current, err
		}
		return timequery.Spec{Kind: timequery.KindSingleDay, On: on}, nil
	case "range":
		from, err := prompter.Input(i18n.T(language, "wizard.from"), current.From, false)
		if err != nil {
			return current, err
		}
		to, err := prompter.Input(i18n.T(language, "wizard.to"), current.To, false)
		if err != nil {
			return current, err
		}
		return timequery.Spec{Kind: timequery.KindRange, From: from, To: to}, nil
	default:
		return current, fmt.Errorf(i18n.T(language, "wizard.unsupported_time_type"), choice)
	}
}

func defaultTimeType(spec timequery.Spec) string {
	switch spec.Kind {
	case timequery.KindSingleDay:
		return "day"
	case timequery.KindRange:
		return "range"
	default:
		return "preset"
	}
}

func defaultPeriod(spec timequery.Spec) string {
	if spec.Kind == timequery.KindPreset && spec.Period != "" {
		return spec.Period
	}
	return timequery.PresetLast7Days
}

func providerAwareValue(previousProvider, selectedProvider, currentValue string, defaultFn func(string) string) string {
	currentValue = strings.TrimSpace(currentValue)
	if currentValue == "" {
		return defaultFn(selectedProvider)
	}
	if strings.TrimSpace(previousProvider) == "" {
		previousProvider = ai.DefaultProvider
	}
	if currentValue == defaultFn(previousProvider) {
		return defaultFn(selectedProvider)
	}
	return currentValue
}

func askLanguage(prompter wizardPrompter, current string) (string, error) {
	current = i18n.NormalizeLanguage(current)
	choice, err := prompter.Select("Language（语言）", optionLabels(languageOptions), optionLabelByValue(languageOptions, current))
	if err != nil {
		return current, err
	}
	return optionValueByLabel(languageOptions, choice), nil
}

func chooseDefault(condition bool, whenTrue, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func askModel(prompter wizardPrompter, provider, current, language string) (string, error) {
	options, ok := modelOptionsByProvider[provider]
	if !ok || len(options) == 0 {
		return prompter.Input(i18n.T(language, "wizard.model"), current, false)
	}
	defaultLabel := optionLabelByValue(options, chooseDefault(current != "", current, ai.DefaultModelName(provider)))
	choice, err := prompter.Select(i18n.T(language, "wizard.model"), optionLabels(options), defaultLabel)
	if err != nil {
		return current, err
	}
	selected := optionValueByLabel(options, choice)
	if selected == "custom-model" {
		return prompter.Input(i18n.T(language, "wizard.custom_model"), current, false)
	}
	return selected, nil
}

func askPrompt(prompter wizardPrompter, current, language string) (string, error) {
	promptOptions := localizedPromptOptions(language)
	defaultLabel := optionLabelByValue(promptOptions, chooseDefault(current != "" && current != "basic" && current != "detailed" && current != "targeted", "custom-file", chooseDefault(current != "", current, "basic")))
	choice, err := prompter.Select(i18n.T(language, "wizard.prompt"), optionLabels(promptOptions), defaultLabel)
	if err != nil {
		return current, err
	}
	selected := optionValueByLabel(promptOptions, choice)
	if selected == "custom-file" {
		return prompter.Input(i18n.T(language, "wizard.prompt.custom_path"), current, false)
	}
	return selected, nil
}

func askProvider(prompter wizardPrompter, current, language string) (string, error) {
	choice, err := prompter.Select(i18n.T(language, "wizard.provider"), optionLabels(providerOptions), optionLabelByValue(providerOptions, current))
	if err != nil {
		return current, err
	}
	return optionValueByLabel(providerOptions, choice), nil
}

func askMappedSelect(prompter wizardPrompter, label string, options []selectOption, current string) (string, error) {
	choice, err := prompter.Select(label, optionLabels(options), optionLabelByValue(options, current))
	if err != nil {
		return current, err
	}
	return optionValueByLabel(options, choice), nil
}

func localizedModeOptions(language string) []selectOption {
	return []selectOption{
		{Label: i18n.T(language, "wizard.mode.single"), Value: "single"},
		{Label: i18n.T(language, "wizard.mode.multi"), Value: "multi"},
	}
}

func localizedTimeTypeOptions(language string) []selectOption {
	return []selectOption{
		{Label: i18n.T(language, "wizard.time_type.preset"), Value: "preset"},
		{Label: i18n.T(language, "wizard.time_type.day"), Value: "day"},
		{Label: i18n.T(language, "wizard.time_type.range"), Value: "range"},
	}
}

func localizedPeriodOptions(language string) []selectOption {
	return []selectOption{
		{Label: localizedPeriodLabel(language, timequery.PresetToday), Value: timequery.PresetToday},
		{Label: localizedPeriodLabel(language, timequery.PresetYesterday), Value: timequery.PresetYesterday},
		{Label: localizedPeriodLabel(language, timequery.PresetLast7Days), Value: timequery.PresetLast7Days},
		{Label: localizedPeriodLabel(language, timequery.PresetLast30Days), Value: timequery.PresetLast30Days},
		{Label: localizedPeriodLabel(language, timequery.PresetThisWeek), Value: timequery.PresetThisWeek},
		{Label: localizedPeriodLabel(language, timequery.PresetLastWeek), Value: timequery.PresetLastWeek},
		{Label: localizedPeriodLabel(language, timequery.PresetThisMonth), Value: timequery.PresetThisMonth},
		{Label: localizedPeriodLabel(language, timequery.PresetLastMonth), Value: timequery.PresetLastMonth},
		{Label: localizedPeriodLabel(language, timequery.PresetThisYear), Value: timequery.PresetThisYear},
		{Label: localizedPeriodLabel(language, timequery.PresetLastYear), Value: timequery.PresetLastYear},
	}
}

func localizedFormatOptions(language string) []selectOption {
	return []selectOption{
		{Label: i18n.T(language, "wizard.format.text"), Value: "text"},
		{Label: i18n.T(language, "wizard.format.markdown"), Value: "markdown"},
	}
}

func localizedPromptOptions(language string) []selectOption {
	return []selectOption{
		{Label: i18n.T(language, "wizard.prompt.basic"), Value: "basic"},
		{Label: i18n.T(language, "wizard.prompt.detailed"), Value: "detailed"},
		{Label: i18n.T(language, "wizard.prompt.targeted"), Value: "targeted"},
		{Label: i18n.T(language, "wizard.prompt.custom"), Value: "custom-file"},
	}
}

func localizedPeriodLabel(language, value string) string {
	if i18n.NormalizeLanguage(language) == i18n.LanguageChinese {
		switch value {
		case timequery.PresetToday:
			return "今天"
		case timequery.PresetYesterday:
			return "昨天"
		case timequery.PresetLast7Days:
			return "最近 7 天"
		case timequery.PresetLast30Days:
			return "最近 30 天"
		case timequery.PresetThisWeek:
			return "本周"
		case timequery.PresetLastWeek:
			return "上周"
		case timequery.PresetThisMonth:
			return "本月"
		case timequery.PresetLastMonth:
			return "上个月"
		case timequery.PresetThisYear:
			return "今年"
		case timequery.PresetLastYear:
			return "去年"
		}
	}
	switch value {
	case timequery.PresetToday:
		return "Today"
	case timequery.PresetYesterday:
		return "Yesterday"
	case timequery.PresetLast7Days:
		return "Last 7 days"
	case timequery.PresetLast30Days:
		return "Last 30 days"
	case timequery.PresetThisWeek:
		return "This week"
	case timequery.PresetLastWeek:
		return "Last week"
	case timequery.PresetThisMonth:
		return "This month"
	case timequery.PresetLastMonth:
		return "Last month"
	case timequery.PresetThisYear:
		return "This year"
	case timequery.PresetLastYear:
		return "Last year"
	default:
		return value
	}
}

func localizedSelectQuestionTemplate(language string) string {
	help := i18n.T(language, "wizard.select_help")
	return strings.ReplaceAll(selectQuestionTemplateBase, "__HELP__", help)
}

func optionLabels(options []selectOption) []string {
	labels := make([]string, 0, len(options))
	for _, option := range options {
		labels = append(labels, option.Label)
	}
	return labels
}

func optionValueByLabel(options []selectOption, label string) string {
	for _, option := range options {
		if option.Label == label {
			return option.Value
		}
	}
	return label
}

func optionLabelByValue(options []selectOption, value string) string {
	for _, option := range options {
		if option.Value == value || option.Label == value {
			return option.Label
		}
	}
	if len(options) == 0 {
		return value
	}
	return options[0].Label
}
