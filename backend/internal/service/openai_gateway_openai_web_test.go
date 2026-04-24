package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveOpenAIWebConversationStateIgnoresLegacySyntheticIDs(t *testing.T) {
	pageSessionID := "page-session-1"
	legacyID := isolateOpenAISessionID(17, pageSessionID)

	conversationID, parentMessageID := resolveOpenAIWebConversationState(&OpenAIWebThread{
		PageSessionID:          pageSessionID,
		UpstreamConversationID: openAIWebStringPtr(legacyID),
		UpstreamSessionID:      openAIWebStringPtr(legacyID),
	}, 17)

	require.Empty(t, conversationID)
	require.Empty(t, parentMessageID)

	conversationID, parentMessageID = resolveOpenAIWebConversationState(&OpenAIWebThread{
		PageSessionID:          pageSessionID,
		UpstreamConversationID: openAIWebStringPtr("conv-real-1"),
		UpstreamSessionID:      openAIWebStringPtr("msg-real-1"),
	}, 17)

	require.Equal(t, "conv-real-1", conversationID)
	require.Equal(t, "msg-real-1", parentMessageID)
}

func TestExtractOpenAIWebConversationTurnParsesAssistantTextAndImages(t *testing.T) {
	mapping := map[string]any{
		"user-1": map[string]any{
			"children": []any{"assistant-1", "tool-1"},
			"message": map[string]any{
				"author":      map[string]any{"role": "user"},
				"status":      "finished_successfully",
				"create_time": 1.0,
				"content": map[string]any{
					"content_type": "text",
					"parts":        []any{"draw a cat"},
				},
			},
		},
		"assistant-1": map[string]any{
			"children": []any{},
			"message": map[string]any{
				"author":      map[string]any{"role": "assistant"},
				"status":      "finished_successfully",
				"create_time": 2.0,
				"content": map[string]any{
					"content_type": "text",
					"parts":        []any{"Here is your image."},
				},
			},
		},
		"tool-1": map[string]any{
			"children": []any{},
			"message": map[string]any{
				"author":      map[string]any{"role": "tool"},
				"status":      "finished_successfully",
				"create_time": 3.0,
				"metadata": map[string]any{
					"async_task_type": "image_gen",
					"image_gen_title": "cat prompt",
				},
				"content": map[string]any{
					"content_type": "multimodal_text",
					"parts": []any{
						map[string]any{
							"asset_pointer": "file-service://file_123",
						},
					},
				},
			},
		},
	}

	turn := extractOpenAIWebConversationTurn(mapping, "user-1")
	require.NotNil(t, turn)
	require.Equal(t, "Here is your image.", turn.AssistantText)
	require.Equal(t, "tool-1", turn.LatestNodeID)
	require.NotNil(t, turn.ResponseID)
	require.Equal(t, "assistant-1", *turn.ResponseID)
	require.False(t, turn.Pending)
	require.Len(t, turn.PointerInfos, 1)
	require.Equal(t, "file-service://file_123", turn.PointerInfos[0].Pointer)
	require.Equal(t, "cat prompt", turn.PointerInfos[0].Prompt)
}

func TestPrepareOpenAIWebUploadsDecodesDataURLs(t *testing.T) {
	uploads, err := prepareOpenAIWebUploads([]OpenAIWebThreadMessageAttachment{
		{
			FileName:    "sample.png",
			ContentType: "image/png",
			DataURL:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO7Z6bQAAAAASUVORK5CYII=",
		},
	})
	require.NoError(t, err)
	require.Len(t, uploads, 1)
	require.Equal(t, "sample.png", uploads[0].FileName)
	require.Equal(t, "image/png", uploads[0].ContentType)
	require.NotEmpty(t, uploads[0].Data)
	require.Equal(t, 1, uploads[0].Width)
	require.Equal(t, 1, uploads[0].Height)
}

func TestBuildOpenAIWebConversationRequestSupportsMixedTextAndImages(t *testing.T) {
	body := buildOpenAIWebConversationRequest(
		"Please describe and refine this image.",
		"gpt-5.4",
		"xhigh",
		"conv-1",
		"parent-1",
		"user-1",
		[]openAIUploadedImage{
			{
				FileID:   "file-123",
				FileName: "reference.png",
				FileSize: 2048,
				MimeType: "image/png",
				Width:    1024,
				Height:   768,
			},
		},
	)

	messages, ok := body["messages"].([]any)
	require.True(t, ok)
	require.Len(t, messages, 1)

	message, ok := messages[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "xhigh", body["reasoning_effort"])

	content, ok := message["content"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "multimodal_text", content["content_type"])

	parts, ok := content["parts"].([]any)
	require.True(t, ok)
	require.Len(t, parts, 2)

	imagePart, ok := parts[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "image_asset_pointer", imagePart["content_type"])
	require.Equal(t, "file-service://file-123", imagePart["asset_pointer"])
	require.Equal(t, "Please describe and refine this image.", parts[1])

	metadata, ok := message["metadata"].(map[string]any)
	require.True(t, ok)
	attachments, ok := metadata["attachments"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, attachments, 1)
	require.Equal(t, "file-123", attachments[0]["id"])
	require.Equal(t, "reference.png", attachments[0]["name"])
	require.Equal(t, "image/png", attachments[0]["mimeType"])
}
