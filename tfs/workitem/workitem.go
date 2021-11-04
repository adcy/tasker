package workitem

import (
	"context"
	"errors"
	"strings"
	"tasker/ptr"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/workitemtracking"
	"github.com/spf13/viper"
)

type Client struct {
	client  workitemtracking.Client
	project string
	team    string
}

type Relation struct {
	URL  string
	Type string
}

func NewClient(ctx context.Context, conn *azuredevops.Connection, team, project string) (*Client, error) {
	client, err := workitemtracking.NewClient(ctx, conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		client:  client,
		project: project,
		team:    team,
	}, nil
}

func (api *Client) Get(ctx context.Context, taskID int) (*workitemtracking.WorkItem, error) {
	return api.client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id: ptr.FromInt(taskID),
	})
}

func (api *Client) FindCommonUserStory(ctx context.Context, iterationPath string) (*workitemtracking.WorkItemReference, error) {
	queryResult, err := api.client.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
		Wiql: &workitemtracking.Wiql{
			Query: ptr.FromStr(`
				SELECT [Id], [Title]
				FROM WorkItems
				WHERE [Work Item Type] = 'User Story'
					AND [System.IterationPath] = '` + iterationPath + `'
					AND [Title] CONTAINS 'Общие задачи'
					AND [State] = 'Active'
			`),
		},
		Project: &api.project,
		Team:    &api.team,
	})
	if err != nil {
		return nil, err
	}

	if len(*queryResult.WorkItems) > 0 {
		return &(*queryResult.WorkItems)[0], nil
	}

	return nil, errors.New("active user story with name '*Общие задачи*' not found in current sprint")
}

func (api *Client) Create(ctx context.Context, title, description, iterationPath string, estimate int, relations []*Relation, tags []string) (*workitemtracking.WorkItem, error) {
	discipline := viper.GetString("tfsDiscipline")

	areaPath := viper.GetString("tfsAreaPath")
	if areaPath == "" {
		areaPath = api.project + "\\" + api.team
	}

	tags = append(tags, "tasker")

	fields := []webapi.JsonPatchOperation{
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/System.IterationPath"),
			Value: iterationPath,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/System.AreaPath"),
			Value: areaPath,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/System.Title"),
			Value: title,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/System.Description"),
			Value: description,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/Microsoft.VSTS.Common.Discipline"),
			Value: discipline,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/Microsoft.VSTS.Scheduling.OriginalEstimate"),
			Value: estimate,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/Microsoft.VSTS.Scheduling.RemainingWork"),
			Value: estimate,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  ptr.FromStr("/fields/System.Tags"),
			Value: strings.Join(tags, "; "),
		},
	}

	for _, relation := range relations {
		fields = append(fields, webapi.JsonPatchOperation{
			Op:   &webapi.OperationValues.Add,
			Path: ptr.FromStr("/relations/-"),
			Value: workitemtracking.WorkItemRelation{
				Rel: ptr.FromStr(relation.Type),
				Url: &relation.URL,
			},
		})
	}

	task, err := api.client.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Type:     ptr.FromStr("Task"),
		Project:  &api.project,
		Document: &fields,
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (api *Client) Assign(ctx context.Context, task *workitemtracking.WorkItem, user string) error {
	_, err := api.client.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Id:      task.Id,
		Project: &api.project,
		Document: &[]webapi.JsonPatchOperation{
			{
				Op:    &webapi.OperationValues.Test,
				Path:  ptr.FromStr("/rev"),
				Value: task.Rev,
			},
			{
				Op:    &webapi.OperationValues.Add,
				Path:  ptr.FromStr("/fields/System.AssignedTo"),
				Value: user,
			},
			{
				Op:    &webapi.OperationValues.Add,
				Path:  ptr.FromStr("/fields/System.State"),
				Value: "InProgress",
			},
		},
	})

	return err
}

func GetURL(w *workitemtracking.WorkItem) string {
	lm, ok := w.Links.(map[string]interface{})
	if ok {
		pm, ok := lm["html"].(map[string]interface{})
		if ok {
			href, ok := pm["href"]
			if ok {
				str, ok := href.(string)
				if ok {
					return str
				}
			}
		}
	}
	return *w.Url
}

func GetTitle(w *workitemtracking.WorkItem) string {
	title, ok := (*w.Fields)["System.Title"]
	if ok {
		titleStr, ok := title.(string)
		if ok {
			return titleStr
		}
	}
	return ""
}

func GetReference(w *workitemtracking.WorkItem) *workitemtracking.WorkItemReference {
	return &workitemtracking.WorkItemReference{
		Id:  w.Id,
		Url: w.Url,
	}
}

func GetIterationPath(w *workitemtracking.WorkItem) string {
	iterationPath, ok := (*w.Fields)["System.IterationPath"]
	if ok {
		iterationPathStr, ok := iterationPath.(string)
		if ok {
			return iterationPathStr
		}
	}
	return ""
}