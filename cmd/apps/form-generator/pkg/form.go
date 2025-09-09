package pkg

import (
    "context"
    "fmt"

    uform "github.com/go-go-golems/uhoh/pkg"
    "github.com/go-go-golems/uhoh/pkg/wizard"
    "github.com/go-go-golems/uhoh/pkg/wizard/steps"

    "google.golang.org/api/forms/v1"
)

// newQuestionItem wraps a forms.Question into a forms.Item with a title.
func newQuestionItem(title string, q *forms.Question) *forms.Item {
    return &forms.Item{
        Title:        title,
        QuestionItem: &forms.QuestionItem{Question: q},
    }
}

// newTextItem creates a static text item (non-question) with optional description.
func newTextItem(title string, description string) *forms.Item {
    return &forms.Item{
        Title:       title,
        Description: description,
        TextItem:    &forms.TextItem{},
    }
}

// newPageBreakItem creates a section/page break item using the step title/description.
func newPageBreakItem(title string, description string) *forms.Item {
    return &forms.Item{
        Title:         title,
        Description:   description,
        PageBreakItem: &forms.PageBreakItem{},
    }
}

// fieldToItem converts a Uhoh field to a Google Forms item.
func fieldToItem(f *uform.Field) (*forms.Item, error) {
    switch f.Type {
    case "input":
        item := newQuestionItem(f.Title, &forms.Question{
            Required:     false,
            TextQuestion: &forms.TextQuestion{Paragraph: false},
        })
        item.Description = f.Description
        return item, nil
    case "text":
        item := newQuestionItem(f.Title, &forms.Question{
            Required:     false,
            TextQuestion: &forms.TextQuestion{Paragraph: true},
        })
        item.Description = f.Description
        return item, nil
    case "select":
        opts := make([]*forms.Option, 0, len(f.Options))
        for _, o := range f.Options {
            label := o.Label
            if label == "" {
                label = fmt.Sprintf("%v", o.Value)
            }
            opts = append(opts, &forms.Option{Value: label})
        }
        item := newQuestionItem(f.Title, &forms.Question{
            Required: true,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "RADIO",
                Options: opts,
                Shuffle: false,
            },
        })
        item.Description = f.Description
        return item, nil
    case "multiselect":
        opts := make([]*forms.Option, 0, len(f.Options))
        for _, o := range f.Options {
            label := o.Label
            if label == "" {
                label = fmt.Sprintf("%v", o.Value)
            }
            opts = append(opts, &forms.Option{Value: label})
        }
        item := newQuestionItem(f.Title, &forms.Question{
            Required: false,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "CHECKBOX",
                Options: opts,
                Shuffle: false,
            },
        })
        item.Description = f.Description
        return item, nil
    case "confirm":
        opts := []*forms.Option{
            {Value: "Yes"},
            {Value: "No"},
        }
        item := newQuestionItem(f.Title, &forms.Question{
            Required: true,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "RADIO",
                Options: opts,
                Shuffle: false,
            },
        })
        item.Description = f.Description
        return item, nil
    default:
        return nil, fmt.Errorf("unsupported field type: %s", f.Type)
    }
}

// BuildRequestsFromWizard transforms a Uhoh wizard into Google Forms batch update requests.
func BuildRequestsFromWizard(wz *wizard.Wizard) ([]*forms.Request, error) {
    requests := []*forms.Request{}
    index := 0

    for _, s := range wz.Steps {
        switch st := s.(type) {
        case *steps.FormStep:
            // Collect items for this step first
            stepItems := []*forms.Item{}
            for _, g := range st.FormData.Groups {
                for _, f := range g.Fields {
                    item, err := fieldToItem(f)
                    if err != nil {
                        continue
                    }
                    stepItems = append(stepItems, item)
                }
            }
            if len(stepItems) == 0 {
                break
            }
            // Insert a page break before the first item of this step (except at index 0)
            if index > 0 {
                pb := newPageBreakItem(st.Title(), st.Description())
                requests = append(requests, &forms.Request{
                    CreateItem: &forms.CreateItemRequest{
                        Item: pb,
                        Location: &forms.Location{
                            Index:           int64(index),
                            ForceSendFields: []string{"Index"},
                        },
                    },
                })
                index++
            }
            for _, item := range stepItems {
                requests = append(requests, &forms.Request{
                    CreateItem: &forms.CreateItemRequest{
                        Item: item,
                        Location: &forms.Location{
                            Index:           int64(index),
                            ForceSendFields: []string{"Index"},
                        },
                    },
                })
                index++
            }
        case *steps.DecisionStep:
            // Map a decision step to a multiple-choice question
            opts := make([]*forms.Option, 0, len(st.Choices))
            for _, c := range st.Choices {
                opts = append(opts, &forms.Option{Value: c})
            }
            item := newQuestionItem(st.Title(), &forms.Question{
                Required: true,
                ChoiceQuestion: &forms.ChoiceQuestion{
                    Type:    "RADIO",
                    Options: opts,
                    Shuffle: false,
                },
            })
            if index > 0 {
                pb := newPageBreakItem(st.Title(), st.Description())
                requests = append(requests, &forms.Request{
                    CreateItem: &forms.CreateItemRequest{
                        Item: pb,
                        Location: &forms.Location{
                            Index:           int64(index),
                            ForceSendFields: []string{"Index"},
                        },
                    },
                })
                index++
            }
            requests = append(requests, &forms.Request{
                CreateItem: &forms.CreateItemRequest{
                    Item: item,
                    Location: &forms.Location{
                        Index:           int64(index),
                        ForceSendFields: []string{"Index"},
                    },
                },
            })
            index++
        case *steps.InfoStep:
            // Map info step content to a static text item
            displayContent := st.Content
            if st.Description() != "" {
                displayContent = fmt.Sprintf("%s\n\n%s", st.Description(), st.Content)
            }
            item := newTextItem(st.Title(), displayContent)
            if index > 0 {
                pb := newPageBreakItem(st.Title(), st.Description())
                requests = append(requests, &forms.Request{
                    CreateItem: &forms.CreateItemRequest{
                        Item: pb,
                        Location: &forms.Location{
                            Index:           int64(index),
                            ForceSendFields: []string{"Index"},
                        },
                    },
                })
                index++
            }
            requests = append(requests, &forms.Request{
                CreateItem: &forms.CreateItemRequest{
                    Item: item,
                    Location: &forms.Location{
                        Index:           int64(index),
                        ForceSendFields: []string{"Index"},
                    },
                },
            })
            index++
        default:
            // Ignore info, action, summary for Google Forms structure
            continue
        }
    }

    return requests, nil
}

// CreateOrUpdateForm creates a new form or reuses an existing one and applies requests.
// If formID is empty, a new form is created with the provided title and description.
// Returns the final form object after updates.
func CreateOrUpdateForm(ctx context.Context, svc *forms.Service, formID string, title string, description string, requests []*forms.Request) (*forms.Form, error) {
    var form *forms.Form
    var err error

    if formID == "" {
        // Create the form (title only per API)
        form, err = svc.Forms.Create(&forms.Form{Info: &forms.Info{Title: title}}).Do()
        if err != nil {
            return nil, fmt.Errorf("failed to create form: %w", err)
        }
    } else {
        // Reuse existing form
        form, err = svc.Forms.Get(formID).Do()
        if err != nil {
            return nil, fmt.Errorf("failed to get existing form %s: %w", formID, err)
        }
        // Optionally update title if provided
        if title != "" && form != nil && form.Info != nil && title != form.Info.Title {
            _, err = svc.Forms.BatchUpdate(form.FormId, &forms.BatchUpdateFormRequest{
                Requests: []*forms.Request{
                    {
                        UpdateFormInfo: &forms.UpdateFormInfoRequest{
                            Info:       &forms.Info{Title: title},
                            UpdateMask: "title",
                        },
                    },
                },
            }).Do()
            if err != nil {
                return nil, fmt.Errorf("failed to update form title: %w", err)
            }
            // refresh
            form, _ = svc.Forms.Get(formID).Do()
        }
    }

    // Handle description update separately to avoid index conflicts
    if description != "" {
        _, err = svc.Forms.BatchUpdate(form.FormId, &forms.BatchUpdateFormRequest{
            Requests: []*forms.Request{
                {
                    UpdateFormInfo: &forms.UpdateFormInfoRequest{
                        Info:       &forms.Info{Description: description},
                        UpdateMask: "description",
                    },
                },
            },
        }).Do()
        if err != nil {
            return nil, fmt.Errorf("failed to update form description: %w", err)
        }
    }

    // Apply item creation requests
    if len(requests) > 0 {
        _, err = svc.Forms.BatchUpdate(form.FormId, &forms.BatchUpdateFormRequest{Requests: requests}).Do()
        if err != nil {
            return nil, fmt.Errorf("batch update failed: %w", err)
        }
    }

    // Fetch updated form
    return svc.Forms.Get(form.FormId).Do()
}

