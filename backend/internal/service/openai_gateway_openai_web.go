package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
	_ "golang.org/x/image/webp"
)

const openAIWebDefaultImageSize = "2K"

const (
	openAIWebMaxAttachmentCount = 4
	openAIWebMaxAttachmentBytes = 8 << 20
)

type openAIWebConversationTurn struct {
	AssistantText string
	PointerInfos  []openAIImagePointerInfo
	LatestNodeID  string
	ResponseID    *string
	Pending       bool
}

type openAIWebResolvedConversationTurn struct {
	AssistantText   string
	AssistantImages []OpenAIWebThreadMessageImage
	LatestNodeID    string
	ResponseID      *string
}

func (t *openAIWebConversationTurn) IsReady() bool {
	if t == nil {
		return false
	}
	if len(t.PointerInfos) > 0 {
		return true
	}
	if strings.TrimSpace(t.AssistantText) != "" {
		return true
	}
	return strings.TrimSpace(t.LatestNodeID) != "" && !t.Pending
}

func (s *OpenAIGatewayService) ForwardOpenAIWebMessage(ctx context.Context, input *OpenAIWebForwardMessageInput) (*OpenAIWebForwardMessageResult, error) {
	if input == nil || input.Thread == nil || input.APIKey == nil || input.Account == nil {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	model := strings.TrimSpace(input.Thread.RequestedModel)
	if requestedModel := strings.TrimSpace(input.RequestedModel); requestedModel != "" {
		model = requestedModel
	}
	if model == "" {
		model = defaultOpenAIWebRequestedModelForAccount(input.Account)
	}
	if isOpenAIImageGenerationModel(model) {
		return s.forwardOpenAIWebImageMessage(ctx, input, model)
	}
	reasoningEffort := resolveOpenAIWebReasoningEffort(input.ReasoningEffort)

	token, tokenType, err := s.GetAccessToken(ctx, input.Account)
	if err != nil {
		return nil, err
	}
	if tokenType != "oauth" {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	client, err := newOpenAIBackendAPIClient(resolveOpenAIProxyURL(input.Account))
	if err != nil {
		return nil, err
	}
	headers, err := s.buildOpenAIBackendAPIHeaders(input.Account, token)
	if err != nil {
		return nil, err
	}
	if bootstrapErr := bootstrapOpenAIBackendAPI(ctx, client, headers); bootstrapErr != nil {
		logger.LegacyPrintf("service.openai_gateway", "OpenAI web bootstrap failed: %v", bootstrapErr)
	}

	chatReqs, err := fetchOpenAIChatRequirements(ctx, client, headers)
	if err != nil {
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, err)
	}
	if chatReqs.Turnstile.Required || chatReqs.Arkose.Required {
		return nil, s.wrapOpenAIImageBackendError(
			ctx,
			nil,
			input.Account,
			newOpenAIImageSyntheticStatusError(
				http.StatusForbidden,
				"chat-requirements requires unsupported challenge",
				openAIChatGPTChatRequirementsURL,
			),
		)
	}

	if initErr := initializeOpenAIImageConversation(ctx, client, headers); initErr != nil {
		logger.LegacyPrintf("service.openai_gateway", "OpenAI web conversation init failed: %v", initErr)
	}

	startTime := time.Now()
	conversationID, parentMessageID := resolveOpenAIWebConversationState(input.Thread, input.APIKey.ID)
	userMessageID := uuid.NewString()
	proofToken := generateOpenAIProofToken(chatReqs.ProofOfWork.Required, chatReqs.ProofOfWork.Seed, chatReqs.ProofOfWork.Difficulty, headers.Get("User-Agent"))
	uploads, err := prepareOpenAIWebUploads(input.Attachments)
	if err != nil {
		return nil, err
	}
	submissionText := openAIWebSubmissionText(strings.TrimSpace(input.Content), len(uploads) > 0)
	conduitToken, parentMessageID, err := prepareOpenAIWebConversation(
		ctx,
		client,
		headers,
		submissionText,
		conversationID,
		parentMessageID,
		model,
		reasoningEffort,
		userMessageID,
		chatReqs.Token,
		proofToken,
	)
	if err != nil {
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, err)
	}

	uploadedFiles, err := uploadOpenAIImageFiles(ctx, client, headers, uploads)
	if err != nil {
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, err)
	}

	convReq := buildOpenAIWebConversationRequest(submissionText, model, reasoningEffort, conversationID, parentMessageID, userMessageID, uploadedFiles)
	requestBody, err := json.Marshal(convReq)
	if err != nil {
		return nil, err
	}

	convHeaders := cloneHTTPHeader(headers)
	convHeaders.Set("Accept", "text/event-stream")
	convHeaders.Set("Content-Type", "application/json")
	convHeaders.Set("openai-sentinel-chat-requirements-token", chatReqs.Token)
	if conduitToken != "" {
		convHeaders.Set("x-conduit-token", conduitToken)
	}
	if proofToken != "" {
		convHeaders.Set("openai-sentinel-proof-token", proofToken)
	}

	resp, err := client.R().
		SetContext(ctx).
		DisableAutoReadResponse().
		SetHeaders(headerToMap(convHeaders)).
		SetBodyJsonMarshal(convReq).
		Post(openAIChatGPTConversationURL)
	if err != nil {
		return nil, fmt.Errorf("openai web conversation request failed: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if resp.StatusCode >= 400 {
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, handleOpenAIImageBackendError(resp))
	}

	streamConversationID, usage, firstTokenMs, asyncMode, err := readOpenAIWebConversationStream(resp, startTime)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(streamConversationID) != "" {
		conversationID = strings.TrimSpace(streamConversationID)
	}

	resolvedTurn, err := fetchOpenAIWebConversationTurn(ctx, client, headers, conversationID, userMessageID, asyncMode)
	if err != nil {
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, err)
	}

	result := &OpenAIForwardResult{
		RequestID:     strings.TrimSpace(resp.Header.Get("x-request-id")),
		Usage:         usage,
		Model:         model,
		UpstreamModel: model,
		Stream:        false,
		Duration:      time.Since(startTime),
		FirstTokenMs:  firstTokenMs,
	}
	if reasoningEffort != "" {
		result.ReasoningEffort = openAIWebStringPtr(reasoningEffort)
	}
	if imageCount := len(resolvedTurn.AssistantImages); imageCount > 0 {
		result.BillingModel = "gpt-image-2"
		result.ImageCount = imageCount
		result.ImageSize = openAIWebDefaultImageSize
	}

	var upstreamConversationID *string
	if strings.TrimSpace(conversationID) != "" {
		upstreamConversationID = openAIWebStringPtr(strings.TrimSpace(conversationID))
	}
	var upstreamSessionID *string
	if strings.TrimSpace(resolvedTurn.LatestNodeID) != "" {
		upstreamSessionID = openAIWebStringPtr(strings.TrimSpace(resolvedTurn.LatestNodeID))
	}

	return &OpenAIWebForwardMessageResult{
		Result:                 result,
		AssistantText:          strings.TrimSpace(resolvedTurn.AssistantText),
		AssistantImages:        append([]OpenAIWebThreadMessageImage(nil), resolvedTurn.AssistantImages...),
		ResponseID:             resolvedTurn.ResponseID,
		RequestPayloadHash:     HashUsageRequestPayload(requestBody),
		UpstreamEndpoint:       "/backend-api/f/conversation",
		UpstreamConversationID: upstreamConversationID,
		UpstreamSessionID:      upstreamSessionID,
	}, nil
}

func (s *OpenAIGatewayService) forwardOpenAIWebImageMessage(
	ctx context.Context,
	input *OpenAIWebForwardMessageInput,
	model string,
) (*OpenAIWebForwardMessageResult, error) {
	if s.httpUpstream == nil {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	requestModel := strings.TrimSpace(model)
	if requestModel == "" {
		requestModel = "gpt-image-2"
	}
	if err := validateOpenAIImagesModel(requestModel); err != nil {
		return nil, err
	}

	token, tokenType, err := s.GetAccessToken(ctx, input.Account)
	if err != nil {
		return nil, err
	}
	if tokenType != "oauth" {
		return nil, ErrOpenAIWebGatewayUnavailable
	}

	uploads, err := prepareOpenAIWebUploads(input.Attachments)
	if err != nil {
		return nil, err
	}

	prompt := strings.TrimSpace(input.Content)
	if prompt == "" && len(uploads) > 0 {
		prompt = "Edit this image."
	}
	endpoint := openAIImagesGenerationsEndpoint
	if len(uploads) > 0 {
		endpoint = openAIImagesEditsEndpoint
	}
	parsed := &OpenAIImagesRequest{
		Endpoint:       endpoint,
		ContentType:    "application/json",
		Model:          requestModel,
		ExplicitModel:  true,
		Prompt:         prompt,
		N:              1,
		ResponseFormat: "b64_json",
		OutputFormat:   "png",
		Uploads:        uploads,
	}
	applyOpenAIImagesDefaults(parsed)
	parsed.SizeTier = normalizeOpenAIImageSizeTier(parsed.Size)
	parsed.RequiredCapability = classifyOpenAIImagesCapability(parsed)

	requestBody, err := buildOpenAIImagesResponsesRequest(parsed, requestModel)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatgptCodexURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Host = "chatgpt.com"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("OpenAI-Beta", "responses=experimental")
	req.Header.Set("originator", resolveOpenAIUpstreamOriginator(nil, false))
	if customUA := strings.TrimSpace(input.Account.GetOpenAIUserAgent()); customUA != "" {
		req.Header.Set("User-Agent", customUA)
	} else {
		req.Header.Set("User-Agent", codexCLIUserAgent)
	}
	if chatgptAccountID := strings.TrimSpace(input.Account.GetChatGPTAccountID()); chatgptAccountID != "" {
		req.Header.Set("chatgpt-account-id", chatgptAccountID)
	}

	proxyURL := ""
	if input.Account.ProxyID != nil && input.Account.Proxy != nil {
		proxyURL = input.Account.Proxy.URL()
	}

	startTime := time.Now()
	resp, err := s.httpUpstream.Do(req, proxyURL, input.Account.ID, input.Account.Concurrency)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, ErrOpenAIWebGatewayUnavailable
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		upstreamMsg := sanitizeUpstreamErrorMessage(extractUpstreamErrorMessage(body))
		if upstreamMsg == "" {
			upstreamMsg = fmt.Sprintf("openai web image request failed: status %d", resp.StatusCode)
		}
		statusErr := &openAIImageStatusError{
			StatusCode:      resp.StatusCode,
			Message:         upstreamMsg,
			ResponseBody:    body,
			ResponseHeaders: resp.Header.Clone(),
			RequestID:       strings.TrimSpace(resp.Header.Get("x-request-id")),
		}
		if resp.Request != nil && resp.Request.URL != nil {
			statusErr.URL = resp.Request.URL.String()
		} else if req.URL != nil {
			statusErr.URL = req.URL.String()
		}
		return nil, s.wrapOpenAIImageBackendError(ctx, nil, input.Account, statusErr)
	}

	body, err := ReadUpstreamResponseBody(resp.Body, s.cfg, nil, nil)
	if err != nil {
		return nil, err
	}

	var usage OpenAIUsage
	for _, line := range bytes.Split(body, []byte("\n")) {
		line = bytes.TrimRight(line, "\r")
		data, ok := extractOpenAISSEDataLine(string(line))
		if !ok || data == "" || data == "[DONE]" {
			continue
		}
		s.parseSSEUsageBytes([]byte(data), &usage)
	}

	results, _, _, firstMeta, _, err := collectOpenAIImagesFromResponsesBody(body)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("upstream did not return image output")
	}
	if strings.TrimSpace(firstMeta.Model) == "" {
		firstMeta.Model = requestModel
	}
	if strings.TrimSpace(firstMeta.Size) != "" {
		parsed.SizeTier = normalizeOpenAIImageSizeTier(firstMeta.Size)
	}

	images := make([]OpenAIWebThreadMessageImage, 0, len(results))
	for _, item := range results {
		mergeOpenAIResponsesImageMeta(&item, firstMeta)
		data := strings.TrimSpace(item.Result)
		if data == "" {
			continue
		}
		mimeType := openAIImageOutputMIMEType(item.OutputFormat)
		width, height := openAIWebImageDimensionsFromSize(item.Size)
		images = append(images, OpenAIWebThreadMessageImage{
			DataURL:       "data:" + mimeType + ";base64," + data,
			MimeType:      mimeType,
			RevisedPrompt: strings.TrimSpace(item.RevisedPrompt),
			Width:         width,
			Height:        height,
		})
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("upstream did not return image output")
	}

	result := &OpenAIForwardResult{
		RequestID:     strings.TrimSpace(resp.Header.Get("x-request-id")),
		Usage:         usage,
		Model:         requestModel,
		UpstreamModel: requestModel,
		BillingModel:  requestModel,
		Stream:        false,
		Duration:      time.Since(startTime),
		ImageCount:    len(images),
		ImageSize:     parsed.SizeTier,
	}

	return &OpenAIWebForwardMessageResult{
		Result:             result,
		AssistantImages:    images,
		RequestPayloadHash: HashUsageRequestPayload(requestBody),
		UpstreamEndpoint:   "/backend-api/codex/responses",
	}, nil
}

func openAIWebImageDimensionsFromSize(size string) (int, int) {
	size = strings.TrimSpace(strings.ToLower(size))
	if size == "" || !strings.Contains(size, "x") {
		return 0, 0
	}
	var width, height int
	if _, err := fmt.Sscanf(size, "%dx%d", &width, &height); err != nil {
		return 0, 0
	}
	if width <= 0 || height <= 0 {
		return 0, 0
	}
	return width, height
}

func prepareOpenAIWebConversation(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	content string,
	conversationID string,
	parentMessageID string,
	model string,
	reasoningEffort string,
	userMessageID string,
	chatToken string,
	proofToken string,
) (string, string, error) {
	parentMessageID = strings.TrimSpace(parentMessageID)
	if parentMessageID == "" {
		parentMessageID = uuid.NewString()
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "auto"
	}

	payload := map[string]any{
		"action":                "next",
		"client_prepare_state":  "success",
		"fork_from_shared_post": false,
		"parent_message_id":     parentMessageID,
		"model":                 model,
		"timezone_offset_min":   openAITimezoneOffsetMinutes(),
		"timezone":              openAITimezoneName(),
		"conversation_mode":     map[string]any{"kind": "primary_assistant"},
		"system_hints":          []string{"picture_v2"},
		"supports_buffering":    true,
		"supported_encodings":   []string{"v1"},
		"partial_query": map[string]any{
			"id":     strings.TrimSpace(userMessageID),
			"author": map[string]any{"role": "user"},
			"content": map[string]any{
				"content_type": "text",
				"parts":        []string{content},
			},
		},
		"client_contextual_info": map[string]any{
			"app_name": "chatgpt.com",
		},
	}
	if strings.TrimSpace(reasoningEffort) != "" {
		payload["reasoning_effort"] = strings.TrimSpace(reasoningEffort)
	}
	if strings.TrimSpace(conversationID) != "" {
		payload["conversation_id"] = strings.TrimSpace(conversationID)
	}

	prepareHeaders := cloneHTTPHeader(headers)
	prepareHeaders.Set("Accept", "*/*")
	prepareHeaders.Set("Content-Type", "application/json")
	if strings.TrimSpace(chatToken) != "" {
		prepareHeaders.Set("openai-sentinel-chat-requirements-token", strings.TrimSpace(chatToken))
	}
	if strings.TrimSpace(proofToken) != "" {
		prepareHeaders.Set("openai-sentinel-proof-token", strings.TrimSpace(proofToken))
	}

	var result struct {
		ConduitToken string `json:"conduit_token"`
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headerToMap(prepareHeaders)).
		SetBodyJsonMarshal(payload).
		SetSuccessResult(&result).
		Post(openAIChatGPTConversationPrepareURL)
	if err != nil {
		return "", parentMessageID, err
	}
	if !resp.IsSuccessState() {
		return "", parentMessageID, newOpenAIImageStatusError(resp, "conversation prepare failed")
	}
	return strings.TrimSpace(result.ConduitToken), parentMessageID, nil
}

func buildOpenAIWebConversationRequest(
	content,
	model,
	reasoningEffort,
	conversationID,
	parentMessageID,
	userMessageID string,
	uploads []openAIUploadedImage,
) map[string]any {
	model = strings.TrimSpace(model)
	if model == "" {
		model = "auto"
	}

	parts := []any{content}
	attachments := make([]map[string]any, 0, len(uploads))
	contentType := "text"
	if len(uploads) > 0 {
		contentType = "multimodal_text"
		parts = make([]any, 0, len(uploads)+1)
		for _, upload := range uploads {
			parts = append(parts, map[string]any{
				"content_type":  "image_asset_pointer",
				"asset_pointer": "file-service://" + upload.FileID,
				"size_bytes":    upload.FileSize,
				"width":         upload.Width,
				"height":        upload.Height,
			})
			attachment := map[string]any{
				"id":       upload.FileID,
				"mimeType": upload.MimeType,
				"name":     upload.FileName,
				"size":     upload.FileSize,
			}
			if upload.Width > 0 {
				attachment["width"] = upload.Width
			}
			if upload.Height > 0 {
				attachment["height"] = upload.Height
			}
			attachments = append(attachments, attachment)
		}
		parts = append(parts, content)
	}

	metadata := map[string]any{
		"developer_mode_connector_ids": []any{},
		"selected_github_repos":        []any{},
		"selected_all_github_repos":    false,
		"system_hints":                 []string{"picture_v2"},
		"serialization_metadata": map[string]any{
			"custom_symbol_offsets": []any{},
		},
	}
	if len(attachments) > 0 {
		metadata["attachments"] = attachments
	}

	message := map[string]any{
		"id":     strings.TrimSpace(userMessageID),
		"author": map[string]any{"role": "user"},
		"content": map[string]any{
			"content_type": contentType,
			"parts":        parts,
		},
		"metadata":    metadata,
		"create_time": float64(time.Now().UnixMilli()) / 1000,
	}

	body := map[string]any{
		"action":                               "next",
		"client_prepare_state":                 "sent",
		"messages":                             []any{message},
		"parent_message_id":                    strings.TrimSpace(parentMessageID),
		"model":                                model,
		"timezone_offset_min":                  openAITimezoneOffsetMinutes(),
		"timezone":                             openAITimezoneName(),
		"conversation_mode":                    map[string]any{"kind": "primary_assistant"},
		"enable_message_followups":             true,
		"system_hints":                         []string{"picture_v2"},
		"supports_buffering":                   true,
		"supported_encodings":                  []string{"v1"},
		"paragen_cot_summary_display_override": "allow",
		"force_parallel_switch":                "auto",
		"client_contextual_info": map[string]any{
			"is_dark_mode":      false,
			"time_since_loaded": 200,
			"page_height":       900,
			"page_width":        1440,
			"pixel_ratio":       1,
			"screen_height":     1080,
			"screen_width":      1920,
			"app_name":          "chatgpt.com",
		},
	}
	if strings.TrimSpace(reasoningEffort) != "" {
		body["reasoning_effort"] = strings.TrimSpace(reasoningEffort)
	}
	if strings.TrimSpace(conversationID) != "" {
		body["conversation_id"] = strings.TrimSpace(conversationID)
	}
	return body
}

func prepareOpenAIWebUploads(attachments []OpenAIWebThreadMessageAttachment) ([]OpenAIImagesUpload, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	if len(attachments) > openAIWebMaxAttachmentCount {
		return nil, ErrOpenAIWebAttachmentOverflow
	}

	uploads := make([]OpenAIImagesUpload, 0, len(attachments))
	for i := range attachments {
		upload, err := decodeOpenAIWebAttachment(attachments[i], i)
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

func decodeOpenAIWebAttachment(item OpenAIWebThreadMessageAttachment, index int) (OpenAIImagesUpload, error) {
	mediaType, data, err := decodeOpenAIWebAttachmentDataURL(item.DataURL)
	if err != nil {
		return OpenAIImagesUpload{}, ErrOpenAIWebAttachmentInvalid.WithCause(err)
	}
	if len(data) == 0 {
		return OpenAIImagesUpload{}, ErrOpenAIWebAttachmentInvalid
	}
	if len(data) > openAIWebMaxAttachmentBytes {
		return OpenAIImagesUpload{}, ErrOpenAIWebAttachmentTooLarge
	}

	contentType := strings.TrimSpace(item.ContentType)
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		contentType = strings.TrimSpace(mediaType)
	}
	if detected := strings.TrimSpace(http.DetectContentType(data)); strings.HasPrefix(strings.ToLower(detected), "image/") {
		if contentType == "" || strings.EqualFold(contentType, "application/octet-stream") {
			contentType = detected
		}
	}
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return OpenAIImagesUpload{}, ErrOpenAIWebAttachmentInvalid
	}

	width := item.Width
	height := item.Height
	if width <= 0 || height <= 0 {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return OpenAIImagesUpload{}, ErrOpenAIWebAttachmentInvalid.WithCause(err)
		}
		width = cfg.Width
		height = cfg.Height
	}

	fileName := strings.TrimSpace(item.FileName)
	if fileName == "" {
		suffix := "png"
		switch {
		case strings.Contains(strings.ToLower(contentType), "jpeg"):
			suffix = "jpg"
		case strings.Contains(strings.ToLower(contentType), "webp"):
			suffix = "webp"
		case strings.Contains(strings.ToLower(contentType), "gif"):
			suffix = "gif"
		}
		fileName = fmt.Sprintf("image_%d.%s", index+1, suffix)
	}

	return OpenAIImagesUpload{
		FieldName:   "image",
		FileName:    fileName,
		ContentType: contentType,
		Data:        data,
		Width:       width,
		Height:      height,
	}, nil
}

func decodeOpenAIWebAttachmentDataURL(dataURL string) (string, []byte, error) {
	dataURL = strings.TrimSpace(dataURL)
	if !strings.HasPrefix(dataURL, "data:") {
		return "", nil, fmt.Errorf("attachment must be a data url")
	}

	header, payload, ok := strings.Cut(strings.TrimPrefix(dataURL, "data:"), ",")
	if !ok {
		return "", nil, fmt.Errorf("attachment data url is malformed")
	}
	if !strings.Contains(strings.ToLower(header), ";base64") {
		return "", nil, fmt.Errorf("attachment data url must be base64 encoded")
	}

	mediaType := strings.TrimSpace(strings.SplitN(header, ";", 2)[0])
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	if err != nil {
		return "", nil, err
	}
	return mediaType, data, nil
}

func resolveOpenAIWebConversationState(thread *OpenAIWebThread, apiKeyID int64) (string, string) {
	if thread == nil {
		return "", ""
	}
	conversationID := strings.TrimSpace(openAIWebDerefStringPtr(thread.UpstreamConversationID))
	parentMessageID := strings.TrimSpace(openAIWebDerefStringPtr(thread.UpstreamSessionID))
	pageSessionID := strings.TrimSpace(thread.PageSessionID)

	if pageSessionID != "" {
		legacyIsolated := isolateOpenAISessionID(apiKeyID, pageSessionID)
		if conversationID == legacyIsolated || parentMessageID == legacyIsolated {
			return "", ""
		}
	}
	if isOpenAIWebLegacySyntheticIdentifier(conversationID) && (parentMessageID == "" || isOpenAIWebLegacySyntheticIdentifier(parentMessageID)) {
		return "", ""
	}
	if isOpenAIWebLegacySyntheticIdentifier(parentMessageID) && conversationID == "" {
		return "", ""
	}
	if conversationID != "" && parentMessageID == "" {
		return "", ""
	}
	return conversationID, parentMessageID
}

func isOpenAIWebLegacySyntheticIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 16 {
		return false
	}
	for i := range value {
		ch := value[i]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
			return false
		}
	}
	return true
}

func readOpenAIWebConversationStream(resp *req.Response, startTime time.Time) (string, OpenAIUsage, *int, bool, error) {
	if resp == nil || resp.Body == nil {
		return "", OpenAIUsage{}, nil, false, fmt.Errorf("empty conversation response")
	}

	reader := bufio.NewReader(resp.Body)
	var (
		conversationID string
		firstTokenMs   *int
		usage          OpenAIUsage
		asyncMode      bool
	)

	for {
		line, err := reader.ReadString('\n')
		if strings.TrimSpace(line) != "" && firstTokenMs == nil {
			ms := int(time.Since(startTime).Milliseconds())
			firstTokenMs = &ms
		}
		if data, ok := extractOpenAISSEDataLine(strings.TrimRight(line, "\r\n")); ok && data != "" && data != "[DONE]" {
			dataBytes := []byte(data)
			if conversationID == "" {
				conversationID = strings.TrimSpace(gjson.GetBytes(dataBytes, "conversation_id").String())
				if conversationID == "" {
					conversationID = strings.TrimSpace(gjson.GetBytes(dataBytes, "v.conversation_id").String())
				}
			}
			mergeOpenAIUsage(&usage, dataBytes)
			asyncStatus := gjson.GetBytes(dataBytes, "async_status")
			if asyncStatus.Exists() {
				if asyncStatus.Int() > 0 {
					asyncMode = true
				}
				if text := strings.TrimSpace(asyncStatus.String()); text != "" && text != "0" && text != "false" {
					asyncMode = true
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", OpenAIUsage{}, firstTokenMs, asyncMode, err
		}
	}

	return conversationID, usage, firstTokenMs, asyncMode, nil
}

func fetchOpenAIWebConversationTurn(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	conversationID string,
	rootMessageID string,
	asyncMode bool,
) (*openAIWebResolvedConversationTurn, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return nil, fmt.Errorf("openai web conversation id is empty")
	}

	deadline := time.Now().Add(20 * time.Second)
	if asyncMode {
		deadline = time.Now().Add(90 * time.Second)
	}

	var (
		lastErr  error
		lastTurn *openAIWebConversationTurn
	)

	for time.Now().Before(deadline) {
		_, mapping, err := fetchOpenAIWebConversationSnapshot(ctx, client, headers, conversationID)
		if err != nil {
			lastErr = err
			if !isOpenAIImageTransientConversationNotFoundError(err) && !asyncMode {
				break
			}
		} else {
			turn := extractOpenAIWebConversationTurn(mapping, rootMessageID)
			if turn != nil {
				lastTurn = turn
				if turn.IsReady() {
					images, imageErr := downloadOpenAIWebConversationImages(ctx, client, headers, conversationID, turn.PointerInfos)
					if imageErr == nil {
						return &openAIWebResolvedConversationTurn{
							AssistantText:   turn.AssistantText,
							AssistantImages: images,
							LatestNodeID:    turn.LatestNodeID,
							ResponseID:      turn.ResponseID,
						}, nil
					}
					lastErr = imageErr
				}
			}
		}

		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	if lastTurn != nil {
		if lastTurn.IsReady() {
			if len(lastTurn.PointerInfos) > 0 {
				images, imageErr := downloadOpenAIWebConversationImages(ctx, client, headers, conversationID, lastTurn.PointerInfos)
				if imageErr == nil {
					return &openAIWebResolvedConversationTurn{
						AssistantText:   lastTurn.AssistantText,
						AssistantImages: images,
						LatestNodeID:    lastTurn.LatestNodeID,
						ResponseID:      lastTurn.ResponseID,
					}, nil
				}
				if lastErr == nil {
					lastErr = imageErr
				}
			}
			return &openAIWebResolvedConversationTurn{
				AssistantText: lastTurn.AssistantText,
				LatestNodeID:  lastTurn.LatestNodeID,
				ResponseID:    lastTurn.ResponseID,
			}, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("openai web conversation turn not found")
}

func fetchOpenAIWebConversationSnapshot(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	conversationID string,
) ([]byte, map[string]any, error) {
	resp, err := client.R().
		SetContext(ctx).
		SetHeaders(headerToMap(headers)).
		DisableAutoReadResponse().
		Get(fmt.Sprintf("https://chatgpt.com/backend-api/conversation/%s", conversationID))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, newOpenAIImageStatusError(resp, "conversation poll failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return body, nil, err
	}
	mapping, _ := decoded["mapping"].(map[string]any)
	return body, mapping, nil
}

func extractOpenAIWebConversationTurn(mapping map[string]any, rootMessageID string) *openAIWebConversationTurn {
	if len(mapping) == 0 {
		return nil
	}

	subtreeIDs := collectOpenAIWebConversationSubtree(mapping, strings.TrimSpace(rootMessageID))
	if len(subtreeIDs) == 0 {
		if strings.TrimSpace(rootMessageID) != "" {
			return &openAIWebConversationTurn{Pending: true}
		}
		subtreeIDs = make(map[string]struct{}, len(mapping))
		for messageID := range mapping {
			subtreeIDs[messageID] = struct{}{}
		}
	}

	type timedText struct {
		ID         string
		CreateTime float64
		Text       string
	}

	turn := &openAIWebConversationTurn{}
	texts := make([]timedText, 0, 2)
	var (
		latestTime           float64
		hasLatest            bool
		latestAssistantTime  float64
		hasAssistantResponse bool
	)

	for messageID := range subtreeIDs {
		node, _ := mapping[messageID].(map[string]any)
		if node == nil {
			continue
		}
		message, _ := node["message"].(map[string]any)
		if message == nil {
			continue
		}
		author, _ := message["author"].(map[string]any)
		role, _ := author["role"].(string)
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}

		status, _ := message["status"].(string)
		finished := isOpenAIWebConversationMessageFinished(status)
		if role != "user" && !finished {
			turn.Pending = true
		}

		createTime := openAIWebConversationCreateTime(message["create_time"])
		if role != "user" && finished {
			if !hasLatest || createTime >= latestTime {
				hasLatest = true
				latestTime = createTime
				turn.LatestNodeID = messageID
			}
			rawMessage, err := json.Marshal(message)
			if err == nil {
				turn.PointerInfos = mergeOpenAIImagePointerInfos(turn.PointerInfos, collectOpenAIImagePointers(rawMessage))
			}
		}

		if role == "assistant" && finished {
			text := openAIWebExtractAssistantText(message["content"])
			if text != "" {
				texts = append(texts, timedText{
					ID:         messageID,
					CreateTime: createTime,
					Text:       text,
				})
			}
			if !hasAssistantResponse || createTime >= latestAssistantTime {
				hasAssistantResponse = true
				latestAssistantTime = createTime
				turn.ResponseID = openAIWebStringPtr(messageID)
			}
		}
	}

	for _, toolMessage := range extractOpenAIImageToolMessages(mapping) {
		if _, ok := subtreeIDs[toolMessage.MessageID]; !ok {
			continue
		}
		turn.PointerInfos = mergeOpenAIImagePointerInfos(turn.PointerInfos, toolMessage.PointerInfos)
		if !hasLatest || toolMessage.CreateTime >= latestTime {
			hasLatest = true
			latestTime = toolMessage.CreateTime
			turn.LatestNodeID = toolMessage.MessageID
		}
	}

	sort.Slice(texts, func(i, j int) bool {
		if texts[i].CreateTime == texts[j].CreateTime {
			return texts[i].ID < texts[j].ID
		}
		return texts[i].CreateTime < texts[j].CreateTime
	})
	parts := make([]string, 0, len(texts))
	for i := range texts {
		if strings.TrimSpace(texts[i].Text) != "" {
			parts = append(parts, texts[i].Text)
		}
	}
	parts = dedupeStrings(parts)
	turn.AssistantText = strings.TrimSpace(strings.Join(parts, "\n\n"))
	if turn.ResponseID == nil && strings.TrimSpace(turn.LatestNodeID) != "" {
		turn.ResponseID = openAIWebStringPtr(strings.TrimSpace(turn.LatestNodeID))
	}
	return turn
}

func collectOpenAIWebConversationSubtree(mapping map[string]any, rootMessageID string) map[string]struct{} {
	rootMessageID = strings.TrimSpace(rootMessageID)
	if rootMessageID == "" {
		return nil
	}
	if _, ok := mapping[rootMessageID]; !ok {
		return nil
	}

	seen := map[string]struct{}{}
	queue := []string{rootMessageID}
	for len(queue) > 0 {
		messageID := queue[0]
		queue = queue[1:]
		if _, ok := seen[messageID]; ok {
			continue
		}
		seen[messageID] = struct{}{}

		node, _ := mapping[messageID].(map[string]any)
		if node == nil {
			continue
		}
		children, _ := node["children"].([]any)
		for _, child := range children {
			childID, _ := child.(string)
			childID = strings.TrimSpace(childID)
			if childID != "" {
				queue = append(queue, childID)
			}
		}
	}
	return seen
}

func isOpenAIWebConversationMessageFinished(status string) bool {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "", "finished", "finished_successfully", "completed", "done":
		return true
	}
	return strings.Contains(status, "finished") || strings.Contains(status, "complete") || strings.Contains(status, "success")
}

func openAIWebExtractAssistantText(content any) string {
	contentMap, _ := content.(map[string]any)
	if contentMap == nil {
		return ""
	}
	if text, _ := contentMap["text"].(string); strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text)
	}

	parts, _ := contentMap["parts"].([]any)
	if len(parts) == 0 {
		return ""
	}
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		switch value := part.(type) {
		case string:
			if text := strings.TrimSpace(value); text != "" && !strings.HasPrefix(text, "file-service://") && !strings.HasPrefix(text, "sediment://") {
				texts = append(texts, text)
			}
		case map[string]any:
			if text, _ := value["text"].(string); strings.TrimSpace(text) != "" {
				texts = append(texts, strings.TrimSpace(text))
			}
		}
	}
	return strings.TrimSpace(strings.Join(dedupeStrings(texts), "\n\n"))
}

func openAIWebConversationCreateTime(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		number, _ := typed.Float64()
		return number
	default:
		return 0
	}
}

func downloadOpenAIWebConversationImages(
	ctx context.Context,
	client *req.Client,
	headers http.Header,
	conversationID string,
	pointers []openAIImagePointerInfo,
) ([]OpenAIWebThreadMessageImage, error) {
	pointers = preferOpenAIFileServicePointerInfos(mergeOpenAIImagePointerInfos(nil, pointers))
	if len(pointers) == 0 {
		return nil, nil
	}

	images := make([]OpenAIWebThreadMessageImage, 0, len(pointers))
	for _, pointer := range pointers {
		downloadURL, err := fetchOpenAIImageDownloadURL(ctx, client, headers, conversationID, pointer.Pointer)
		if err != nil {
			return nil, err
		}
		data, err := downloadOpenAIImageBytes(ctx, client, headers, downloadURL)
		if err != nil {
			return nil, err
		}
		mimeType, width, height := openAIWebDetectImageMetadata(data)
		images = append(images, OpenAIWebThreadMessageImage{
			DataURL:       fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)),
			MimeType:      mimeType,
			RevisedPrompt: strings.TrimSpace(pointer.Prompt),
			Width:         width,
			Height:        height,
		})
	}
	return images, nil
}

func openAIWebDetectImageMetadata(data []byte) (string, int, int) {
	mimeType := strings.TrimSpace(http.DetectContentType(data))
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = "image/png"
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return mimeType, 0, 0
	}
	if strings.TrimSpace(format) == "webp" {
		mimeType = "image/webp"
	}
	return mimeType, cfg.Width, cfg.Height
}

func openAIWebStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func openAIWebDerefStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func openAIWebSubmissionText(content string, hasAttachments bool) string {
	content = strings.TrimSpace(content)
	if content != "" {
		return content
	}
	if hasAttachments {
		return "Please analyze the attached image."
	}
	return ""
}
