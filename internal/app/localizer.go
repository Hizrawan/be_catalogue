package app

import (
	"encoding/json"
	"fmt"

	"be20250107/internal/config"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Localizer struct {
	localizer map[string]*i18n.Localizer
}

func NewLocalizer(config config.LocalizerConfig) *Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, supportedLang := range config.SupportedLanguages {
		filepath := fmt.Sprintf("%v/%v.json", config.Directory, supportedLang)
		bundle.MustLoadMessageFile(filepath)
	}

	localizer := &Localizer{
		localizer: make(map[string]*i18n.Localizer),
	}
	for _, supportedLang := range config.SupportedLanguages {
		localizer.localizer[supportedLang] = i18n.NewLocalizer(bundle, supportedLang)
	}

	return localizer
}

func (l *Localizer) GetLocalizedLanguage(langTag string, messageID string, templateData map[string]any) string {
	usedLocalizer := l.localizer[langTag]
	if usedLocalizer == nil {
		usedLocalizer = l.localizer["en-US"]
	}

	translatedMessage, _ := usedLocalizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
		TemplateData: templateData,
	})

	if translatedMessage == "" {
		defaultEngMessage, _ := l.localizer["en-US"].Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID: messageID,
			},
			TemplateData: templateData,
		})
		return defaultEngMessage
	}

	return translatedMessage
}

func (l *Localizer) GetLocalizedDefaultAndPrefMessage(langTag string, messageID string, templateData map[string]any) map[string]string {
	messages := make(map[string]string)
	defaultTranslatedMessage, _ := l.localizer["en-US"].Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
		TemplateData: templateData,
	})
	messages["en-US"] = defaultTranslatedMessage

	prefLanguageLocalizer := l.localizer[langTag]
	if prefLanguageLocalizer == nil {
		return messages
	}

	prefTranslatedMessage, _ := prefLanguageLocalizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: messageID,
		},
		TemplateData: templateData,
	})
	messages[langTag] = prefTranslatedMessage

	return messages
}
