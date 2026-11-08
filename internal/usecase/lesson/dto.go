package lesson

import "html/template"

// CardDTO — проекция урока специально для карточек на главной странице.
type CardDTO struct {
	Slug          string
	Title         string
	Summary       string
	Subject       string
	ContentType   string
	HasVideo      bool
	ReadingTime   int
	DateFormatted string
}

// DetailDTO — полная проекция урока для страницы чтения статьи / просмотра видео.
type DetailDTO struct {
	Slug          string
	Title         string
	Summary       string
	BodyHTML      template.HTML // Безопасный HTML-конспект для рендера
	SubjectCode   string        // Например "history" (для кнопки «Назад к направлению»)
	SubjectName   string        // Например "История"
	HasVideo      bool
	VideoURL      string
	ReadingTime   int
	DateFormatted string
}
