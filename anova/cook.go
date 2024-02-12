package anova

import "fmt"

type Cook struct {
	Stages []*CookStage

	id   string
	oven *Oven
}

func NewCook(stages ...*CookStage) (*Cook, error) {
	for _, stage := range stages {
		if err := (*stage).Validate(); err != nil {
			return nil, ErrInvalidCookStage{Stage: stage, StageError: err}
		}
	}

	return &Cook{
		Stages: stages,

		id: generateRandomCookUuid(),
	}, nil
}

func (cook *Cook) Start(oven *Oven) error {
	if cook.oven != nil {
		return ErrCookAlreadyStarted{}
	}
	cook.oven = oven
	for _, stage := range cook.Stages {
		stage.cook = cook
	}
	return oven.StartCook(cook.Stages)
}

//TODO fix; doesn't seem to do anything currently. maybe we need to use the set_* methods?
//func (cook *Cook) Update() error {
//	if cook.oven == nil {
//		return ErrCookNotStarted{}
//	}
//	return cook.oven.UpdateCookStages(cook.Stages)
//}

func (cook *Cook) Stop() error {
	if cook.oven == nil {
		return ErrCookNotStarted{}
	}
	err := cook.oven.StopCook()
	if err != nil {
		cook.oven = nil
	}
	return err
}

type ErrInvalidCookStage struct {
	Stage      *CookStage
	StageError error
}

func (err ErrInvalidCookStage) Error() string {
	return fmt.Sprintf("invalid cook stage: %s (stage: %+v)", err.StageError, err.Stage)
}

type ErrCookNotStarted struct{}

func (err ErrCookNotStarted) Error() string {
	return "cannot mutate cook that has not been started"
}

type ErrCookAlreadyStarted struct{}

func (err ErrCookAlreadyStarted) Error() string {
	return "cannot start cook that has already been started"
}
