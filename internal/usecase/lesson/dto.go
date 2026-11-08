package lesson

import "html/template"

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

type DetailDTO struct {
	Slug          string
	Title         string
	Summary       string
	BodyHTML      template.HTML
	SubjectCode   string
	SubjectName   string
	HasVideo      bool
	VideoURL      string
	ImageURL      string
	ReadingTime   int
	DateFormatted string
}

type CreateLessonInput struct {
	Slug        string
	Title       string
	Summary     string
	Body        string
	Subject     string
	ContentType string
	VideoURL    string
	ImageURL    string
	ReadingTime int
}

type UpdateLessonInput struct {
	Slug        string
	Title       string
	Summary     string
	Body        string
	Subject     string
	ContentType string
	VideoURL    string
	ImageURL    string
	ReadingTime int
}

type EditLessonDTO struct {
	Slug        string
	Title       string
	Summary     string
	Body        string
	Subject     string
	ContentType string
	VideoURL    string
	ImageURL    string
	ReadingTime int
}
