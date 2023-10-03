package files

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/storage"
)

const (
	fNamelanguage  = "GeoLang.json"
	fNameHistory   = "History.json"
	fNameLawBasics = "LawBasics.json"
)

type Storage struct {
	basePath string
}

type questionInFile struct {
	Number  int            `json:"Number"`
	Answers []answerInFile `json:"Answers"`
	Text    string         `json:"Text"`
}

type answerInFile struct {
	Code              string `json:"Code"`
	Text              string `json:"Text"`
	ThisCorrectAnswer bool   `json:"ThisCorrectAnswer"`
}

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) PickQuestion(topic storage.Topic, number int) (*storage.Question, error) {

	topicQuestions := s.topicQuestions(topic)

	if len(topicQuestions) < number {
		err := fmt.Errorf("No more questions")
		return nil, err
	}

	topicQuestion := topicQuestions[number-1]

	question := &storage.Question{
		Topic:  topic,
		Number: topicQuestion.Number,
		Text:   topicQuestion.Text,
	}

	for _, answerInFile := range topicQuestion.Answers {

		newAnswer := &storage.Answer{
			Code:              answerInFile.Code,
			Text:              answerInFile.Text,
			ThisCorrectAnswer: answerInFile.ThisCorrectAnswer,
		}
		question.Answers = append(question.Answers, *newAnswer)
	}

	return question, nil

}

func (s Storage) QuantityQuestions(topic storage.Topic) int {

	topicQuestions := s.topicQuestions(topic)

	return len(topicQuestions)

}

func (s Storage) topicQuestions(topic storage.Topic) []*questionInFile {

	fileName := ""

	switch topic {
	case storage.GeorgianLanguage:
		fileName = fNamelanguage
	case storage.History:
		fileName = fNameHistory
	case storage.LawBasics:
		fileName = fNameLawBasics
	}

	file, err := os.ReadFile(filepath.Join(s.basePath, fileName))
	if err != nil {
		log.Fatal(err)
	}

	questions := []*questionInFile{}
	err = json.Unmarshal(file, &questions)

	if err != nil {
		log.Fatal(err)
	}

	return questions
}
