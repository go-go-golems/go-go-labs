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

// fieldToItem converts a Uhoh field to a Google Forms item.
func fieldToItem(f *uform.Field) (*forms.Item, error) {
    switch f.Type {
    case "input":
        return newQuestionItem(f.Title, &forms.Question{
            Required:     false,
            TextQuestion: &forms.TextQuestion{Paragraph: false},
        }), nil
    case "text":
        return newQuestionItem(f.Title, &forms.Question{
            Required:     false,
            TextQuestion: &forms.TextQuestion{Paragraph: true},
        }), nil
    case "select":
        opts := make([]*forms.Option, 0, len(f.Options))
        for _, o := range f.Options {
            label := o.Label
            if label == "" {
                label = fmt.Sprintf("%v", o.Value)
            }
            opts = append(opts, &forms.Option{Value: label})
        }
        return newQuestionItem(f.Title, &forms.Question{
            Required: true,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "RADIO",
                Options: opts,
                Shuffle: false,
            },
        }), nil
    case "multiselect":
        opts := make([]*forms.Option, 0, len(f.Options))
        for _, o := range f.Options {
            label := o.Label
            if label == "" {
                label = fmt.Sprintf("%v", o.Value)
            }
            opts = append(opts, &forms.Option{Value: label})
        }
        return newQuestionItem(f.Title, &forms.Question{
            Required: false,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "CHECKBOX",
                Options: opts,
                Shuffle: false,
            },
        }), nil
    case "confirm":
        opts := []*forms.Option{
            {Value: "Yes"},
            {Value: "No"},
        }
        return newQuestionItem(f.Title, &forms.Question{
            Required: true,
            ChoiceQuestion: &forms.ChoiceQuestion{
                Type:    "RADIO",
                Options: opts,
                Shuffle: false,
            },
        }), nil
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
            for _, g := range st.FormData.Groups {
                for _, f := range g.Fields {
                    item, err := fieldToItem(f)
                    if err != nil {
                        // Skip unsupported field types rather than failing the whole form
                        continue
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
                }
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

