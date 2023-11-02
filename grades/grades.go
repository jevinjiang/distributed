package grades

import (
	"fmt"
	"sync"
)

type Student struct {
	ID        int
	FirstName string
	LastName  string
	Grades    []Grade
}

type GradeType string

const (
	GradeQuiz = GradeType("Quiz")
	GradeTest = GradeType("Test")
	GradeExam = GradeType("Exam")
)

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}

func (stu Student) Average() float32 {
	var result float32
	for _, g := range stu.Grades {
		result += g.Score
	}
	return result / float32(len(stu.Grades))
}

type Students []Student

var (
	students      Students
	studentsMutex sync.Mutex
)

func (students Students) FindByID(id int) (*Student, error) {
	for _, stu := range students {
		if stu.ID == id {
			return &stu, nil
		}
	}
	return nil, fmt.Errorf("student with i %d not found", id)
}

func (students Students) GetByID(id int) (*Student, error) {
	for _, stu := range students {
		if stu.ID == id {
			return &stu, nil
		}
	}
	return nil, fmt.Errorf("student with i %d not found", id)
}
