package memory

import (
	"context"
	"errors"
	"sync"

	"edu-platform/internal/domain/lesson"
)

type LessonRepository struct {
	mu      sync.RWMutex
	lessons []*lesson.Lesson
}

func NewLessonRepository() *LessonRepository {
	repo := &LessonRepository{}
	repo.seed()
	return repo
}

func (r *LessonRepository) FindPublished(ctx context.Context) ([]*lesson.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*lesson.Lesson
	for _, l := range r.lessons {
		if l.IsPublished() {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *LessonRepository) FindPublishedBySubject(ctx context.Context, subject lesson.Subject) ([]*lesson.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*lesson.Lesson
	for _, l := range r.lessons {
		if l.IsPublished() && l.Subject() == subject {
			result = append(result, l)
		}
	}
	return result, nil
}

func (r *LessonRepository) FindBySlug(ctx context.Context, slug string) (*lesson.Lesson, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, l := range r.lessons {
		if l.Slug() == slug && l.IsPublished() {
			return l, nil
		}
	}
	return nil, lesson.ErrLessonNotFound
}

func (r *LessonRepository) seed() {
	l1, _ := lesson.NewLesson(
		"1",
		"vostochnye-slavyane-v-drevnosti",
		"Восточные славяне в древности: расселение, быт и верования",
		"Разбираем карту расселения племенных союзов, хозяйственный уклад, подсечно-огневое земледелие и путь «из варяг в греки».",
		`<h3>1. Расселение восточнославянских племен</h3>...`,
		lesson.SubjectHistory,
		lesson.TypeMixed,
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"", // <-- Пустая строка для imageURL по умолчанию в seed
		12,
	)

	l2, _ := lesson.NewLesson(
		"2",
		"reformy-izbrannoy-rady",
		"Реформы Избранной рады: централизация при Иване IV",
		"Видеоразбор Земского собора 1549 года, принятия Судебника 1550 года, отмены кормлений и создания стрелецкого войска.",
		`<h3>Основные направления реформ 1550-х годов</h3>...`,
		lesson.SubjectHistory,
		lesson.TypeVideo,
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"", // <-- Пустая строка для imageURL
		25,
	)

	l3, _ := lesson.NewLesson(
		"3",
		"politicheskaya-sistema-obshchestva",
		"Политическая система общества и её институты",
		"Структура политической системы: институциональная, нормативная, коммуникативная и культурно-идеологическая подсистемы с примерами для второй части экзамена.",
		`<h3>Понятие и структура политической системы</h3>...`,
		lesson.SubjectSocial,
		lesson.TypeArticle,
		"", // Видео нет
		"", // <-- Пустая строка для imageURL
		8,
	)

	l4, _ := lesson.NewLesson(
		"4",
		"ekonomicheskie-sistemy-i-sobstvennost",
		"Типы экономических систем: традиционная, командная, рыночная",
		"Сравнительный анализ экономических систем, функции рынка, несовершенства рыночного механизма и роль государства в экономике.",
		`<h3>Главные вопросы экономики</h3>...`,
		lesson.SubjectSocial,
		lesson.TypeMixed,
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"", // <-- Пустая строка для imageURL
		15,
	)

	r.lessons = []*lesson.Lesson{l1, l2, l3, l4}
}

func (r *LessonRepository) Save(ctx context.Context, l *lesson.Lesson) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.lessons {
		if existing.Slug() == l.Slug() {
			return errors.New("lesson with this slug already exists")
		}
	}

	r.lessons = append([]*lesson.Lesson{l}, r.lessons...) // Добавляем новый урок в начало списка
	return nil
}

func (r *LessonRepository) Delete(ctx context.Context, slug string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, l := range r.lessons {
		if l.Slug() == slug {
			// Удаляем элемент из среза с сохранением порядка остальных
			r.lessons = append(r.lessons[:i], r.lessons[i+1:]...)
			return nil
		}
	}
	return lesson.ErrLessonNotFound
}

func (r *LessonRepository) Update(ctx context.Context, updated *lesson.Lesson) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.lessons {
		if existing.Slug() == updated.Slug() {
			r.lessons[i] = updated
			return nil
		}
	}

	return lesson.ErrLessonNotFound
}
