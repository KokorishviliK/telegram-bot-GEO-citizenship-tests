package storage

type Storage interface {
	PickQuestion(topic Topic, number int) (*Question, error)
	QuantityQuestions(topic Topic) int
}

type Topic int

const (
	GeorgianLanguage Topic = iota + 1
	History
	LawBasics
)

type Question struct {
	Topic   Topic
	Number  int
	Answers []Answer
	Text    string
}

type Answer struct {
	Text              string
	Code              string
	ThisCorrectAnswer bool
}
