package pkg

import (
	"fmt"
	"sort"
	"strings"

	uform "github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/wizard"
	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/forms/v1"
)

type FieldMetadata struct {
	QuestionID string
	FieldKey   string
	FieldTitle string
	StepID     string
	StepTitle  string
	StepType   string
}

type stepAccumulator struct {
	rawID       string
	title       string
	description string
	items       []*forms.Item
}

func ConvertFormToWizard(form *forms.Form) (*wizard.Wizard, map[string]FieldMetadata, error) {
	if form == nil {
		return nil, nil, fmt.Errorf("form is nil")
	}

	w := &wizard.Wizard{}
	if form.Info != nil {
		w.Name = strings.TrimSpace(form.Info.Title)
		w.Description = strings.TrimSpace(form.Info.Description)
	}

	var stepsSlice steps.WizardSteps
	fieldMeta := map[string]FieldMetadata{}
	usedStepIDs := map[string]int{}
	usedFieldKeys := map[string]int{}

	var acc *stepAccumulator
	stepIndex := 0

	flush := func() error {
		if acc == nil {
			return nil
		}
		if len(acc.items) == 0 {
			acc = nil
			return nil
		}
		step, metas, err := buildStepFromAccumulator(acc, stepIndex, usedStepIDs, usedFieldKeys)
		if err != nil {
			return err
		}
		if step != nil {
			stepsSlice = append(stepsSlice, step)
			stepIndex++
			for _, m := range metas {
				if m.QuestionID != "" {
					fieldMeta[m.QuestionID] = m
				}
			}
		}
		acc = nil
		return nil
	}

	for _, item := range form.Items {
		if item == nil {
			continue
		}
		if item.PageBreakItem != nil {
			if err := flush(); err != nil {
				return nil, nil, err
			}
			acc = &stepAccumulator{
				rawID:       item.ItemId,
				title:       strings.TrimSpace(item.Title),
				description: strings.TrimSpace(item.Description),
				items:       []*forms.Item{},
			}
			continue
		}

		if acc == nil {
			acc = &stepAccumulator{
				rawID: strings.TrimSpace(item.ItemId),
				title: strings.TrimSpace(item.Title),
				items: []*forms.Item{},
			}
			if desc := strings.TrimSpace(item.Description); desc != "" && item.TextItem == nil {
				acc.description = desc
			}
		} else if acc.title == "" {
			acc.title = strings.TrimSpace(item.Title)
		}
		acc.items = append(acc.items, item)
	}

	if err := flush(); err != nil {
		return nil, nil, err
	}

	w.Steps = stepsSlice
	return w, fieldMeta, nil
}

func buildStepFromAccumulator(acc *stepAccumulator, index int, usedStepIDs map[string]int, usedFieldKeys map[string]int) (steps.Step, []FieldMetadata, error) {
	if acc == nil {
		return nil, nil, nil
	}

	stepTitle := strings.TrimSpace(acc.title)
	if stepTitle == "" && len(acc.items) > 0 {
		stepTitle = strings.TrimSpace(acc.items[0].Title)
	}
	stepDescription := strings.TrimSpace(acc.description)

	stepID := uniqueStepID(stepTitle, acc.rawID, index, usedStepIDs)

	questionItems := make([]*forms.Item, 0, len(acc.items))
	textItems := make([]*forms.Item, 0, len(acc.items))
	for _, item := range acc.items {
		if item == nil {
			continue
		}
		if item.QuestionItem != nil && item.QuestionItem.Question != nil {
			questionItems = append(questionItems, item)
			continue
		}
		if item.TextItem != nil {
			textItems = append(textItems, item)
		}
	}

	if len(questionItems) == 0 && len(textItems) > 0 {
		return buildInfoStep(stepID, stepTitle, stepDescription, textItems[0])
	}

	if len(questionItems) == 1 && len(textItems) == 0 {
		if canInterpretAsDecision(questionItems[0]) {
			step, meta := buildDecisionStep(stepID, stepTitle, stepDescription, questionItems[0], usedFieldKeys)
			return step, meta, nil
		}
	}

	if len(questionItems) > 0 {
		return buildFormStep(stepID, stepTitle, stepDescription, questionItems, usedFieldKeys)
	}

	log.Debug().Str("stepID", stepID).Msg("Skipping step with no interpretable items")
	return nil, nil, nil
}

func buildInfoStep(stepID, title, description string, item *forms.Item) (steps.Step, []FieldMetadata, error) {
	if item == nil {
		return nil, nil, fmt.Errorf("info step item is nil")
	}
	content := strings.TrimSpace(item.Description)
	if description != "" {
		prefix := strings.TrimSpace(description) + "\n\n"
		if strings.HasPrefix(content, prefix) {
			content = strings.TrimSpace(strings.TrimPrefix(content, prefix))
		}
	}
	info := &steps.InfoStep{
		BaseStep: steps.BaseStep{
			StepID:          stepID,
			StepType:        "info",
			StepTitle:       title,
			StepDescription: description,
		},
		Content: content,
	}
	return info, nil, nil
}

func buildDecisionStep(stepID, title, description string, item *forms.Item, usedFieldKeys map[string]int) (steps.Step, []FieldMetadata) {
	q := item.QuestionItem.Question
	options := []string{}
	if q.ChoiceQuestion != nil {
		for _, opt := range q.ChoiceQuestion.Options {
			if opt == nil {
				continue
			}
			val := strings.TrimSpace(opt.Value)
			if val == "" {
				continue
			}
			options = append(options, val)
		}
	}
	targetKey := uniqueFieldKey(title, q.QuestionId, usedFieldKeys)
	decision := &steps.DecisionStep{
		BaseStep: steps.BaseStep{
			StepID:          stepID,
			StepType:        "decision",
			StepTitle:       title,
			StepDescription: description,
		},
		TargetKey: targetKey,
		Choices:   options,
	}
	meta := FieldMetadata{
		QuestionID: q.QuestionId,
		FieldKey:   targetKey,
		FieldTitle: title,
		StepID:     stepID,
		StepTitle:  title,
		StepType:   "decision",
	}
	return decision, []FieldMetadata{meta}
}

func buildFormStep(stepID, title, description string, items []*forms.Item, usedFieldKeys map[string]int) (steps.Step, []FieldMetadata, error) {
	fields := []*uform.Field{}
	metas := []FieldMetadata{}

	for _, item := range items {
		q := item.QuestionItem.Question
		fieldType, options := interpretQuestion(q)
		if fieldType == "" {
			log.Debug().Str("stepID", stepID).Msg("Skipping unsupported question type")
			continue
		}
		fieldTitle := strings.TrimSpace(item.Title)
		if fieldTitle == "" {
			fieldTitle = "Untitled Question"
		}
		key := uniqueFieldKey(fieldTitle, q.QuestionId, usedFieldKeys)
		field := &uform.Field{
			Type:        fieldType,
			Key:         key,
			Title:       fieldTitle,
			Description: strings.TrimSpace(item.Description),
			Options:     options,
		}
		fields = append(fields, field)
		metas = append(metas, FieldMetadata{
			QuestionID: q.QuestionId,
			FieldKey:   key,
			FieldTitle: fieldTitle,
			StepID:     stepID,
			StepTitle:  title,
			StepType:   "form",
		})
	}

	if len(fields) == 0 {
		return nil, metas, nil
	}

	group := &uform.Group{
		Fields: fields,
	}
	formStep := &steps.FormStep{
		BaseStep: steps.BaseStep{
			StepID:          stepID,
			StepType:        "form",
			StepTitle:       title,
			StepDescription: description,
		},
		FormData: uform.Form{
			Groups: []*uform.Group{group},
		},
	}
	return formStep, metas, nil
}

func interpretQuestion(q *forms.Question) (string, []*uform.Option) {
	if q == nil {
		return "", nil
	}
	if q.TextQuestion != nil {
		if q.TextQuestion.Paragraph {
			return "text", nil
		}
		return "input", nil
	}
	if q.ChoiceQuestion != nil {
		switch strings.ToUpper(q.ChoiceQuestion.Type) {
		case "CHECKBOX":
			return "multiselect", mapOptions(q.ChoiceQuestion.Options)
		case "RADIO":
			if looksLikeConfirm(q.ChoiceQuestion.Options) {
				return "confirm", nil
			}
			return "select", mapOptions(q.ChoiceQuestion.Options)
		}
	}
	return "", nil
}

func mapOptions(opts []*forms.Option) []*uform.Option {
	if len(opts) == 0 {
		return nil
	}
	res := make([]*uform.Option, 0, len(opts))
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		label := strings.TrimSpace(opt.Value)
		if label == "" {
			continue
		}
		res = append(res, &uform.Option{Label: label, Value: label})
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func looksLikeConfirm(opts []*forms.Option) bool {
	if len(opts) != 2 {
		return false
	}
	values := []string{}
	for _, opt := range opts {
		if opt == nil {
			return false
		}
		values = append(values, strings.ToLower(strings.TrimSpace(opt.Value)))
	}
	sort.Strings(values)
	return values[0] == "no" && values[1] == "yes"
}

func canInterpretAsDecision(item *forms.Item) bool {
	if item == nil || item.QuestionItem == nil {
		return false
	}
	q := item.QuestionItem.Question
	if q == nil || q.ChoiceQuestion == nil {
		return false
	}
	return strings.ToUpper(q.ChoiceQuestion.Type) == "RADIO"
}

func uniqueStepID(title, fallback string, index int, used map[string]int) string {
	if stepID, _, ok := decodeMetadata(fallback); ok && stepID != "" {
		return ensureUnique(stepID, used)
	}
	candidate := slug.Make(strings.TrimSpace(title))
	if candidate == "" {
		candidate = slug.Make(strings.TrimSpace(fallback))
	}
	if candidate == "" {
		candidate = fmt.Sprintf("step-%d", index+1)
	}
	return ensureUnique(candidate, used)
}

func uniqueFieldKey(title, fallback string, used map[string]int) string {
	if _, fieldKey, ok := decodeMetadata(fallback); ok && fieldKey != "" {
		return ensureUnique(fieldKey, used)
	}
	candidate := slug.Make(strings.TrimSpace(title))
	if candidate == "" {
		candidate = slug.Make(strings.TrimSpace(fallback))
	}
	if candidate == "" {
		candidate = fmt.Sprintf("field-%d", len(used)+1)
	}
	return ensureUnique(candidate, used)
}

func ensureUnique(candidate string, used map[string]int) string {
	if candidate == "" {
		candidate = "item"
	}
	if count, ok := used[candidate]; ok {
		count++
		used[candidate] = count
		return fmt.Sprintf("%s-%d", candidate, count)
	}
	used[candidate] = 1
	return candidate
}
