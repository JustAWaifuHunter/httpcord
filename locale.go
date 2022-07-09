package httpcord

type (
	Locale string
	Dictionary map[Locale]string
)

const (
	EnglishUS    Locale = "en-US"
	EnglishGB    Locale = "en-GB"
	Bulgarian    Locale = "bg"
	ChineseCN    Locale = "zh-CN"
	ChineseTW    Locale = "zh-TW"
	Croatian     Locale = "hr"
	Czech        Locale = "cs"
	Danish       Locale = "da"
	Dutch        Locale = "nl"
	Finnish      Locale = "fi"
	French       Locale = "fr"
	German       Locale = "de"
	Greek        Locale = "el"
	Hindi        Locale = "hi"
	Hungarian    Locale = "hu"
	Italian      Locale = "it"
	Japanese     Locale = "ja"
	Korean       Locale = "ko"
	Lithuanian   Locale = "lt"
	Norwegian    Locale = "no"
	Polish       Locale = "pl"
	PortugueseBR Locale = "pt-BR"
	Romanian     Locale = "ro"
	Russian      Locale = "ru"
	SpanishES    Locale = "es-ES"
	Swedish      Locale = "sv-SE"
	Thai         Locale = "th"
	Turkish      Locale = "tr"
	Ukrainian    Locale = "uk"
	Vietnamese   Locale = "vi"
)
