package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kway-teow/git-digest/internal/app"
	"github.com/kway-teow/git-digest/internal/i18n"
	"github.com/kway-teow/git-digest/internal/timequery"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// 版本信息，由 GoReleaser 在构建时注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var cliCfg app.Config

var rootCmd = &cobra.Command{
	Use: "git-digest",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd, args, false)
	},
}

var wizardCmd = &cobra.Command{
	Use: "wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(cmd, args, true)
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(_ *cobra.Command, _ []string) {
		language := preferredLanguage(os.Args[1:])
		fmt.Printf(i18n.T(language, "main.version_label")+"\n", version)
		fmt.Printf(i18n.T(language, "main.commit_label")+"\n", commit)
		fmt.Printf(i18n.T(language, "main.build_date_label")+"\n", date)
	},
}

func init() {
	language := preferredLanguage(os.Args[1:])
	localizeCLI(language)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(wizardCmd)

	rootCmd.PersistentFlags().StringVar(&cliCfg.Language, "language", "", i18n.T(language, "flag.language"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Time.Period, "period", "", i18n.T(language, "flag.period"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Time.On, "on", "", i18n.T(language, "flag.on"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Time.From, "from", "", i18n.T(language, "flag.from"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Time.To, "to", "", i18n.T(language, "flag.to"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Format, "format", "text", i18n.T(language, "flag.format"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.OutputFile, "output", "", i18n.T(language, "flag.output"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.RepoPath, "repo", "", i18n.T(language, "flag.repo"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.ReposPath, "repos", "", i18n.T(language, "flag.repos"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Provider, "provider", "", i18n.T(language, "flag.provider"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.BaseURL, "base-url", "", i18n.T(language, "flag.base_url"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.APIKey, "api-key", "", i18n.T(language, "flag.api_key"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Model, "model", "", i18n.T(language, "flag.model"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Author, "author", "", i18n.T(language, "flag.author"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.Prompt, "prompt", "basic", i18n.T(language, "flag.prompt"))

	rootCmd.PersistentFlags().BoolVar(&cliCfg.Interactive, "interactive", false, i18n.T(language, "flag.interactive"))
	rootCmd.PersistentFlags().StringVar(&cliCfg.ConfigPath, "config", "", i18n.T(language, "flag.config"))
	rootCmd.PersistentFlags().BoolVar(&cliCfg.NoConfig, "no-config", false, i18n.T(language, "flag.no_config"))
	rootCmd.PersistentFlags().BoolVar(&cliCfg.SaveAsDefault, "save-config", false, i18n.T(language, "flag.save_config"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string, forceInteractive bool) error {
	base := app.DefaultConfig()
	changed := changedFlags(cmd)
	normalizeTimeFlags(&cliCfg, changed)

	cfgPath := cliCfg.ConfigPath
	if cfgPath == "" {
		defaultPath, err := app.DefaultConfigPath()
		if err != nil {
			return err
		}
		cfgPath = defaultPath
	}

	fileCfg := app.Config{}
	if !cliCfg.NoConfig {
		loaded, err := app.LoadConfig(cfgPath, preferredLanguage(os.Args[1:]))
		if err != nil {
			return err
		}
		fileCfg = loaded
	}

	cfg := app.MergeConfig(base, fileCfg, cliCfg, changed)
	if forceInteractive {
		cfg.Interactive = true
	}

	if app.ShouldUseInteractive(args, cfg.Interactive) {
		interactiveCfg, err := app.RunWizard(os.Stdin, os.Stdout, cfg)
		if err != nil {
			return err
		}
		cfg = interactiveCfg
	}

	if cfg.SaveAsDefault {
		if err := app.SaveConfig(cfgPath, cfg, cfg.Language); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, i18n.T(cfg.Language, "main.saved_config")+"\n", cfgPath)
	}

	service := app.NewService(os.Stdout, os.Stderr)
	return service.Run(cfg)
}

func changedFlags(cmd *cobra.Command) map[string]bool {
	changed := make(map[string]bool)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		changed[flag.Name] = true
	})
	return changed
}

func normalizeTimeFlags(cfg *app.Config, changed map[string]bool) {
	switch {
	case changed["period"]:
		cfg.Time.Kind = timequery.KindPreset
	case changed["on"]:
		cfg.Time.Kind = timequery.KindSingleDay
	case changed["from"] || changed["to"]:
		cfg.Time.Kind = timequery.KindRange
	}
}

func preferredLanguage(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--language" && i+1 < len(args):
			return i18n.NormalizeLanguage(args[i+1])
		case strings.HasPrefix(arg, "--language="):
			return i18n.NormalizeLanguage(strings.TrimPrefix(arg, "--language="))
		}
	}
	return i18n.LanguageEnglish
}

func localizeCLI(language string) {
	rootCmd.Short = i18n.T(language, "main.short")
	rootCmd.Long = i18n.T(language, "main.long")
	wizardCmd.Short = i18n.T(language, "main.wizard_short")
	versionCmd.Short = i18n.T(language, "main.version_short")
}
