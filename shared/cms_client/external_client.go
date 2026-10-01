package cms_client

import (
	"fmt"

	"github.com/google/uuid"
)

const rpcPath = "/api/v1/rpc"

type ListAttemptsModel struct {
	AttemptNumber  int    `json:"id"`
	AttemptID      string `json:"attempt_id"`
	UserID         int64  `json:"user_id"`
	UserName       string `json:"user_name"`
	Status         string `json:"status"`
	ServerClientID string `json:"server_client_id"`
	RoutingName    string `json:"lti_routing_name"`
	Repository     string `json:"repository"`
	TestPath       string `json:"test_path"`
	Result         any    `json:"result,omitempty"`
}

type ListAttemptsParams struct {
	AttemptIDs      []string `json:"attempt_ids,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	Offset          int      `json:"offset,omitempty"`
	ServerClientIDs []string `json:"server_client_ids,omitempty"`
	Statuses        []string `json:"statuses,omitempty"`
	UserIDs         []int64  `json:"user_ids,omitempty"`
}

func (c *CMSClient) ListAttempts(params ListAttemptsParams) ([]ListAttemptsModel, error) {
	var result struct {
		Result struct {
			Model []ListAttemptsModel `json:"model"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	resp, err := c.client.R().SetBasicAuth(c.Config.ClientID, c.Config.Token).
		SetBody(map[string]any{
			"jsonrpc": "2.0",
			"id":      uuid.New().String(),
			"method":  "lti_attempt.list_external",
			"params":  params,
		}).
		SetResult(&result).
		Post(rpcPath)

	if err != nil {
		return nil, fmt.Errorf("list_attempts: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("list_attempts: http %d: %s", resp.StatusCode(), resp.String())
	}
	if result.Error != nil {
		return nil, fmt.Errorf("list_attempts: rpc %d: %s", result.Error.Code, result.Error.Message)
	}

	return result.Result.Model, nil
}

type UpdateAttemptParams struct {
	AttemptID string `json:"attempt_id"`
	Status    string `json:"status"`
	Result    any    `json:"result,omitempty"`
}

func (c *CMSClient) UpdateAttempts(models []UpdateAttemptParams) (int, error) {
	var result struct {
		Result struct {
			Count int `json:"count"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	resp, err := c.client.R().SetBasicAuth(c.Config.ClientID, c.Config.Token).
		SetBody(map[string]any{
			"jsonrpc": "2.0",
			"id":      uuid.New().String(),
			"method":  "lti_attempt.update_external",
			"params":  map[string]any{"models": models},
		}).
		SetResult(&result).
		Post(rpcPath)

	if err != nil {
		return 0, fmt.Errorf("update_attempts: %w", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("update_attempts: http %d: %s", resp.StatusCode(), resp.String())
	}
	if result.Error != nil {
		return 0, fmt.Errorf("update_attempts: rpc %d: %s", result.Error.Code, result.Error.Message)
	}

	return result.Result.Count, nil
}
