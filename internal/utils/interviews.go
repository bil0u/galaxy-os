package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Interview logic

type Interview []InterviewQuestion

func (i Interview) IsValid() (bool, error) {
	var errs []error
	for _, question := range i {
		err := question.isProperlyConfigured()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return false, fmt.Errorf("invalid interview: %v", errs)
	}
	return true, nil
}

type InterviewQuestion struct {
	Question        string                `toml:"question"`
	Type            InterviewQuestionType `toml:"type"`
	PossibleAnswers []string              `toml:"possible_answers"`
	Weights         map[string]int        `toml:"weights"`

	UserAnswers []string
}

type InterviewQuestionType string

const (
	QuestionTypeFreeText       InterviewQuestionType = "free-text"
	QuestionTypeFreeNumber     InterviewQuestionType = "free-number"
	QuestionTypeMultipleChoice InterviewQuestionType = "multiple-choice"
	QuestionTypeSingleChoice   InterviewQuestionType = "single-choice"
)

// IsProperlyConfigured checks if the question is properly configured
func (i InterviewQuestion) isProperlyConfigured() error {
	var errs []error
	if i.Question == "" {
		errs = append(errs, fmt.Errorf("question is required"))
	}
	if i.Type == "" {
		errs = append(errs, fmt.Errorf("type is required"))
	}

	switch i.Type {
	case QuestionTypeFreeText:
		if len(i.PossibleAnswers) != 0 {
			errs = append(errs, fmt.Errorf("possible answers are not allowed for free text questions"))
		}
		if len(i.Weights) != 0 {
			errs = append(errs, fmt.Errorf("weights are not allowed for free text questions"))
		}
	case QuestionTypeFreeNumber:
		if len(i.PossibleAnswers) == 0 {
			errs = append(errs, fmt.Errorf("possible answers are required for free number questions"))
		}
		if len(i.Weights) != 0 {
			errs = append(errs, fmt.Errorf("weights are not allowed for free number questions"))
		}
		for _, possibleAnswer := range i.PossibleAnswers {
			errHelpString := `
			The possible answers for a free number question should be:
			- a plain number, allowing negative numbers starting with '-'
			- a comparison operator ['>', '<', '>=', '<='] followed by a number
			- a range operator [<number>-<number>] where the first number is lower than the second`
			pattern := regexp.MustCompile(`^(-?[0-9]+|[><]=?[0-9]+|[0-9]+-[0-9]+)$`)
			if !pattern.MatchString(possibleAnswer) {
				errs = append(errs, fmt.Errorf("invalid possible answer '%s'.\n%s", possibleAnswer, errHelpString))
			}
		}
	case QuestionTypeMultipleChoice:
		if len(i.PossibleAnswers) < 2 {
			errs = append(errs, fmt.Errorf("multiple choice question should have at least two possible answers"))
		}
		if len(i.Weights) != 0 && len(i.PossibleAnswers) != len(i.Weights) {
			errs = append(errs, fmt.Errorf("weights should have the same length as possible answers"))
		}

	case QuestionTypeSingleChoice:
		if len(i.PossibleAnswers) < 2 {
			errs = append(errs, fmt.Errorf("single choice question should have at least two possible answers"))
		}
		if len(i.Weights) != 0 && len(i.PossibleAnswers) != len(i.Weights) {
			errs = append(errs, fmt.Errorf("weights should have the same length as possible answers"))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid interview question: %v", errs)
	}
	return nil
}

// answerRespectFormat checks if the answer respects the format of the question
func (i InterviewQuestion) answerRespectFormat() (bool, error) {
	switch i.Type {
	case QuestionTypeFreeText:
		if len(i.UserAnswers) != 1 {
			return false, fmt.Errorf("free text question should have only one answer")
		}
		return true, nil
	case QuestionTypeFreeNumber:
		if len(i.UserAnswers) != 1 {
			return false, fmt.Errorf("free number question should have only one answer")
		}
		// Check if the answer is a number using regex, allowing negative numbers, and extract the number and the sign
		pattern := regexp.MustCompile(`^-?[0-9]+$`)
		if !pattern.MatchString(i.UserAnswers[0]) {
			return false, fmt.Errorf("free number question should have a number as answer")
		}

		// If we have possible answers, check if the answer respects each of the constraints
		if len(i.PossibleAnswers) == 0 {
			return true, nil
		}

		userAnswerInt := i.getAnswerAsInt()

		valid, err := intRespectConstraint(userAnswerInt, i.PossibleAnswers)
		if !valid {
			return false, err
		}
		return true, nil
	case QuestionTypeMultipleChoice:
		if len(i.UserAnswers) < 1 {
			return false, fmt.Errorf("multiple choice question should have at least one answer")
		}
		if len(i.PossibleAnswers) > 0 {
			for _, userAnswer := range i.UserAnswers {
				if !Contains(i.PossibleAnswers, userAnswer) {
					return false, fmt.Errorf("invalid answer '%s'", userAnswer)
				}
			}
		}
		return true, nil
	case QuestionTypeSingleChoice:
		if len(i.UserAnswers) != 1 {
			return false, fmt.Errorf("single choice question should have only one answer")
		}
		if len(i.PossibleAnswers) > 0 && !Contains(i.PossibleAnswers, i.UserAnswers[0]) {
			return false, fmt.Errorf("invalid answer '%s'", i.UserAnswers[0])
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown question type")
	}
}

func (i InterviewQuestion) getAnswerAsString(joinerIfMany string) string {
	if len(i.UserAnswers) == 0 {
		return ""
	} else if len(i.UserAnswers) == 1 {
		return i.UserAnswers[0]
	}
	return strings.Join(i.UserAnswers, joinerIfMany)
}

func (i InterviewQuestion) getAnswerAsInt() int {
	if len(i.UserAnswers) != 1 {
		return 0
	}
	answer, err := strconv.Atoi(i.UserAnswers[0])
	if err != nil {
		return 0
	}
	return answer
}

func (i InterviewQuestion) getAnswerAsSlice() []string {
	if len(i.UserAnswers) == 0 {
		return make([]string, 0)
	}
	return i.UserAnswers
}

func intRespectConstraint(intValue int, constraints []string) (bool, error) {

	// Constains should always be provided
	if len(constraints) == 0 {
		return false, fmt.Errorf("constraints should be provided")
	}

	var errs []error
	for _, possibleAnswer := range constraints {
		// Check if the constraint is a plain number
		plainNumber, err := strconv.Atoi(possibleAnswer)
		if err == nil {
			if intValue == plainNumber {
				return true, nil
			}
			errs = append(errs, fmt.Errorf("answer '%d' does not equal '%d'", intValue, plainNumber))
			continue
		}
		// Check >=/<= before >/< to avoid prefix mismatch
		if strings.HasPrefix(possibleAnswer, ">=") {
			possibleAnswerInt, _ := strconv.Atoi(possibleAnswer[2:])
			if !(intValue >= possibleAnswerInt) {
				errs = append(errs, fmt.Errorf("answer '%d' is not greater or equal to '%d'", intValue, possibleAnswerInt))
			}
		} else if strings.HasPrefix(possibleAnswer, "<=") {
			possibleAnswerInt, _ := strconv.Atoi(possibleAnswer[2:])
			if !(intValue <= possibleAnswerInt) {
				errs = append(errs, fmt.Errorf("answer '%d' is not lower or equal to '%d'", intValue, possibleAnswerInt))
			}
		} else if strings.HasPrefix(possibleAnswer, ">") {
			possibleAnswerInt, _ := strconv.Atoi(possibleAnswer[1:])
			if !(intValue > possibleAnswerInt) {
				errs = append(errs, fmt.Errorf("answer '%d' is not greater than '%d'", intValue, possibleAnswerInt))
			}
		} else if strings.HasPrefix(possibleAnswer, "<") {
			possibleAnswerInt, _ := strconv.Atoi(possibleAnswer[1:])
			if !(intValue < possibleAnswerInt) {
				errs = append(errs, fmt.Errorf("answer '%d' is not lower than '%d'", intValue, possibleAnswerInt))
			}
		}

		// Check if the answer is a range operator matching the pattern
		pattern := regexp.MustCompile(`^(-?[0-9]+)-(-?[0-9]+)$`)
		if pattern.MatchString(possibleAnswer) {
			possibleAnswers := strings.Split(possibleAnswer, "-")
			possibleAnswerMin, _ := strconv.Atoi(possibleAnswers[0])
			possibleAnswerMax, _ := strconv.Atoi(possibleAnswers[1])
			if !(intValue >= possibleAnswerMin && intValue <= possibleAnswerMax) {
				errs = append(errs, fmt.Errorf("answer '%d' is not in range [%d-%d]", intValue, possibleAnswerMin, possibleAnswerMax))
			}
		}
	}

	if len(errs) > 0 {
		return false, fmt.Errorf("invalid answer: %v", errs)
	}
	return true, nil
}
