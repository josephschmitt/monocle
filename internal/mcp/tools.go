package mcp

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/josephschmitt/monocle/internal/client"
	"github.com/josephschmitt/monocle/internal/protocol"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed tools.json
var toolsJSON []byte

type toolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var (
	toolsOnce    sync.Once
	toolDescMap  map[string]string
)

// toolDescriptions returns a map of tool name → description loaded from tools.json.
func toolDescriptions() map[string]string {
	toolsOnce.Do(func() {
		var defs []toolDef
		if err := json.Unmarshal(toolsJSON, &defs); err != nil {
			panic(fmt.Sprintf("parse embedded tools.json: %v", err))
		}
		toolDescMap = make(map[string]string, len(defs))
		for _, d := range defs {
			toolDescMap[d.Name] = d.Description
		}
	})
	return toolDescMap
}

func registerTools(s *sdkmcp.Server) {
	desc := toolDescriptions()

	sdkmcp.AddTool(s, &sdkmcp.Tool{
		Name:        "review_status",
		Description: desc["review_status"],
	}, handleReviewStatus)

	sdkmcp.AddTool(s, &sdkmcp.Tool{
		Name:        "get_feedback",
		Description: desc["get_feedback"],
	}, handleGetFeedback)

	sdkmcp.AddTool(s, &sdkmcp.Tool{
		Name:        "send_artifact",
		Description: desc["send_artifact"],
	}, handleSendArtifact)

	sdkmcp.AddTool(s, &sdkmcp.Tool{
		Name:        "add_files",
		Description: desc["add_files"],
	}, handleAddFiles)
}

// -- Tool parameter types --

type reviewStatusParams struct{}

type getFeedbackParams struct {
	Wait bool `json:"wait,omitempty"`
}

type sendArtifactParams struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type addFilesParams struct {
	Paths []string `json:"paths"`
}

// -- Tool handlers --

func handleReviewStatus(ctx context.Context, req *sdkmcp.CallToolRequest, _ reviewStatusParams) (*sdkmcp.CallToolResult, any, error) {
	c, err := client.ConnectDefault()
	if err != nil {
		return errResult("connect: %v", err), nil, nil
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.GetReviewStatusMsg{Type: protocol.TypeGetReviewStatus},
		client.DefaultTimeout,
	)
	if err != nil {
		return errResult("request: %v", err), nil, nil
	}

	status := resp.(*protocol.GetReviewStatusResponse)
	text := status.Status
	if status.Summary != "" {
		text = status.Summary
	}
	return textResult(text), nil, nil
}

func handleGetFeedback(ctx context.Context, req *sdkmcp.CallToolRequest, params getFeedbackParams) (*sdkmcp.CallToolResult, any, error) {
	c, err := client.ConnectDefault()
	if err != nil {
		return errResult("connect: %v", err), nil, nil
	}
	defer c.Close()

	timeout := client.DefaultTimeout
	if params.Wait {
		timeout = 0 // no deadline — block until feedback
	}

	resp, err := c.Request(
		&protocol.PollFeedbackMsg{Type: protocol.TypePollFeedback, Wait: params.Wait},
		timeout,
	)
	if err != nil {
		return errResult("request: %v", err), nil, nil
	}

	feedback := resp.(*protocol.PollFeedbackResponse)
	if !feedback.HasFeedback {
		return textResult("No feedback pending."), nil, nil
	}
	return textResult(feedback.Feedback), nil, nil
}

func handleSendArtifact(ctx context.Context, req *sdkmcp.CallToolRequest, params sendArtifactParams) (*sdkmcp.CallToolResult, any, error) {
	content := params.Content
	if content == "" && params.FilePath != "" {
		data, err := os.ReadFile(params.FilePath)
		if err != nil {
			return errResult("read file: %v", err), nil, nil
		}
		content = string(data)
		if params.ID == "" {
			params.ID = filepath.Base(params.FilePath)
		}
	}
	if content == "" {
		return errResult("either content or file_path is required"), nil, nil
	}

	c, err := client.ConnectDefault()
	if err != nil {
		return errResult("connect: %v", err), nil, nil
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.SubmitContentMsg{
			Type:        protocol.TypeSubmitContent,
			ID:          params.ID,
			Title:       params.Title,
			Content:     content,
			ContentType: params.ContentType,
			IsPlan:      true,
		},
		client.DefaultTimeout,
	)
	if err != nil {
		return errResult("request: %v", err), nil, nil
	}

	submit := resp.(*protocol.SubmitContentResponse)
	// Include the server-minted id when the caller passed an empty ID —
	// without this the agent has no way to address the artifact later
	// (mark reviewed, dismiss, fetch versions).
	body := submit.Message
	if submit.ID != "" {
		body = fmt.Sprintf("%s\nid: %s", submit.Message, submit.ID)
	}
	return textResult(body), nil, nil
}

func handleAddFiles(ctx context.Context, req *sdkmcp.CallToolRequest, params addFilesParams) (*sdkmcp.CallToolResult, any, error) {
	c, err := client.ConnectDefault()
	if err != nil {
		return errResult("connect: %v", err), nil, nil
	}
	defer c.Close()

	resp, err := c.Request(
		&protocol.AddAdditionalFilesMsg{
			Type:  protocol.TypeAddAdditionalFiles,
			Paths: params.Paths,
		},
		client.DefaultTimeout,
	)
	if err != nil {
		return errResult("request: %v", err), nil, nil
	}

	add := resp.(*protocol.AddAdditionalFilesResponse)
	return textResult(add.Message), nil, nil
}

// -- Helpers --

func textResult(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: text},
		},
	}
}

func errResult(format string, args ...any) *sdkmcp.CallToolResult {
	r := &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: fmt.Sprintf(format, args...)},
		},
		IsError: true,
	}
	return r
}
