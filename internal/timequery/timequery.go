package timequery

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Kind string

const (
	KindPreset    Kind = "preset"
	KindSingleDay Kind = "single_day"
	KindRange     Kind = "range"
)

const (
	LanguageEnglish = "en"
	LanguageChinese = "zh"
)

const (
	PresetToday      = "today"
	PresetYesterday  = "yesterday"
	PresetLast7Days  = "last-7d"
	PresetLast30Days = "last-30d"
	PresetThisWeek   = "this-week"
	PresetLastWeek   = "last-week"
	PresetThisMonth  = "this-month"
	PresetLastMonth  = "last-month"
	PresetThisYear   = "this-year"
	PresetLastYear   = "last-year"
)

type Spec struct {
	Kind   Kind   `json:"kind,omitempty"`
	Period string `json:"period,omitempty"`
	On     string `json:"on,omitempty"`
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
}

type Window struct {
	Start time.Time
	End   time.Time
	Label string
	Kind  Kind
}

func DefaultSpec() Spec {
	return Spec{Kind: KindPreset, Period: PresetLast7Days}
}

func HasValue(spec Spec) bool {
	return spec.Kind != "" || spec.Period != "" || spec.On != "" || spec.From != "" || spec.To != ""
}

func NormalizeLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case LanguageChinese:
		return LanguageChinese
	default:
		return LanguageEnglish
	}
}

func Resolve(spec Spec, loc *time.Location, now time.Time) (Window, error) {
	return ResolveWithLanguage(spec, loc, now, LanguageEnglish)
}

func ResolveWithLanguage(spec Spec, loc *time.Location, now time.Time, language string) (Window, error) {
	if loc == nil {
		loc = time.Local
	}
	if !HasValue(spec) {
		spec = DefaultSpec()
	}
	language = NormalizeLanguage(language)
	spec.Kind = Kind(strings.TrimSpace(string(spec.Kind)))
	spec.Period = strings.TrimSpace(spec.Period)
	spec.On = strings.TrimSpace(spec.On)
	spec.From = strings.TrimSpace(spec.From)
	spec.To = strings.TrimSpace(spec.To)
	if err := validateSpec(spec, language); err != nil {
		return Window{}, err
	}

	now = now.In(loc)
	switch spec.Kind {
	case KindPreset:
		return resolvePreset(spec.Period, loc, now, language)
	case KindSingleDay:
		return resolveSingleDay(spec.On, loc, language)
	case KindRange:
		return resolveRange(spec.From, spec.To, loc, language)
	default:
		return Window{}, localizedf(language, "unsupported time kind %s", "不支持的时间类型: %s", spec.Kind)
	}
}

func validateSpec(spec Spec, language string) error {
	switch spec.Kind {
	case KindPreset:
		if spec.Period == "" {
			return localized(language, "--period is required", "--period 不能为空")
		}
		if spec.On != "" || spec.From != "" || spec.To != "" {
			return localized(language, "--period cannot be used with --on or --from/--to", "--period 不能与 --on 或 --from/--to 同时使用")
		}
	case KindSingleDay:
		if spec.On == "" {
			return localized(language, "--on is required", "--on 不能为空")
		}
		if spec.Period != "" || spec.From != "" || spec.To != "" {
			return localized(language, "--on cannot be used with --period or --from/--to", "--on 不能与 --period 或 --from/--to 同时使用")
		}
	case KindRange:
		if spec.Period != "" || spec.On != "" {
			return localized(language, "--from/--to cannot be used with --period or --on", "--from/--to 不能与 --period 或 --on 同时使用")
		}
		if spec.From == "" || spec.To == "" {
			return localized(language, "--from and --to must be provided together", "--from 和 --to 必须同时提供")
		}
	default:
		return localizedf(language, "unsupported time kind %s", "不支持的时间类型: %s", spec.Kind)
	}
	return nil
}

func resolvePreset(period string, loc *time.Location, now time.Time, language string) (Window, error) {
	switch period {
	case PresetToday:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		return Window{Start: start, End: now, Label: reportLabel("day", language), Kind: KindPreset}, nil
	case PresetYesterday:
		y := now.AddDate(0, 0, -1)
		start := time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, loc)
		return Window{Start: start, End: endOfDay(start), Label: reportLabel("day", language), Kind: KindPreset}, nil
	case PresetLast7Days:
		return Window{Start: now.AddDate(0, 0, -7), End: now, Label: reportLabel("week", language), Kind: KindPreset}, nil
	case PresetLast30Days:
		return Window{Start: now.AddDate(0, 0, -30), End: now, Label: reportLabel("month", language), Kind: KindPreset}, nil
	case PresetThisWeek:
		start := startOfWeek(now, loc)
		return Window{Start: start, End: now, Label: reportLabel("week", language), Kind: KindPreset}, nil
	case PresetLastWeek:
		thisWeekStart := startOfWeek(now, loc)
		lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
		return Window{Start: lastWeekStart, End: thisWeekStart.Add(-time.Second), Label: reportLabel("week", language), Kind: KindPreset}, nil
	case PresetThisMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return Window{Start: start, End: now, Label: reportLabel("month", language), Kind: KindPreset}, nil
	case PresetLastMonth:
		thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
		return Window{Start: lastMonthStart, End: thisMonthStart.Add(-time.Second), Label: reportLabel("month", language), Kind: KindPreset}, nil
	case PresetThisYear:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		return Window{Start: start, End: now, Label: reportLabel("year", language), Kind: KindPreset}, nil
	case PresetLastYear:
		thisYearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		lastYearStart := thisYearStart.AddDate(-1, 0, 0)
		return Window{Start: lastYearStart, End: thisYearStart.Add(-time.Second), Label: reportLabel("year", language), Kind: KindPreset}, nil
	default:
		return Window{}, localizedf(language, "unsupported period: %s", "不支持的时间周期: %s", period)
	}
}

func resolveSingleDay(on string, loc *time.Location, language string) (Window, error) {
	d, err := time.ParseInLocation("2006-01-02", on, loc)
	if err != nil {
		return Window{}, localizedf(language, "invalid date format: %v", "指定日期格式错误: %v", err)
	}
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	return Window{Start: start, End: endOfDay(start), Label: reportLabel("day", language), Kind: KindSingleDay}, nil
}

func resolveRange(fromRaw, toRaw string, loc *time.Location, language string) (Window, error) {
	if fromRaw == "" || toRaw == "" {
		return Window{}, localized(language, "--from and --to must be provided together", "--from 和 --to 必须同时提供")
	}
	from, err := time.ParseInLocation("2006-01-02", fromRaw, loc)
	if err != nil {
		return Window{}, localizedf(language, "invalid start date: %v", "开始日期格式错误: %v", err)
	}
	to, err := time.ParseInLocation("2006-01-02", toRaw, loc)
	if err != nil {
		return Window{}, localizedf(language, "invalid end date: %v", "结束日期格式错误: %v", err)
	}
	end := endOfDay(to)
	return Window{Start: from, End: end, Label: labelForDuration(from, end, language), Kind: KindRange}, nil
}

func labelForDuration(from, to time.Time, language string) string {
	daysDiff := to.Sub(from).Hours() / 24
	switch {
	case daysDiff <= 1:
		return reportLabel("day", language)
	case daysDiff <= 7:
		return reportLabel("week", language)
	case daysDiff <= 31:
		return reportLabel("month", language)
	case daysDiff <= 366:
		return reportLabel("year", language)
	default:
		return reportLabel("report", language)
	}
}

func reportLabel(kind, language string) string {
	if NormalizeLanguage(language) == LanguageChinese {
		switch kind {
		case "day":
			return "工作日报"
		case "week":
			return "工作周报"
		case "month":
			return "工作月报"
		case "year":
			return "工作年报"
		default:
			return "工作报告"
		}
	}
	switch kind {
	case "day":
		return "Daily Report"
	case "week":
		return "Weekly Report"
	case "month":
		return "Monthly Report"
	case "year":
		return "Yearly Report"
	default:
		return "Report"
	}
}

func startOfWeek(now time.Time, loc *time.Location) time.Time {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(weekday - 1))
}

func endOfDay(day time.Time) time.Time {
	return day.Add(24*time.Hour - time.Second)
}

func localized(language, en, zh string) error {
	if NormalizeLanguage(language) == LanguageChinese {
		return errors.New(zh)
	}
	return errors.New(en)
}

func localizedf(language, en, zh string, args ...any) error {
	if NormalizeLanguage(language) == LanguageChinese {
		return fmt.Errorf(zh, args...)
	}
	return fmt.Errorf(en, args...)
}
