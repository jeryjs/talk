package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type StreamChunk struct {
	Content   string
	Timestamp time.Time
	Source    string
}

type StreamBuffer struct {
	chunks []StreamChunk
	output chan string
	buffer strings.Builder
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewStreamBuffer(ctx context.Context) *StreamBuffer {
	ctx, cancel := context.WithCancel(ctx)
	return &StreamBuffer{
		chunks: make([]StreamChunk, 0),
		output: make(chan string, 100),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (sb *StreamBuffer) Write(content string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	chunk := StreamChunk{
		Content:   content,
		Timestamp: time.Now(),
		Source:    "ai",
	}

	sb.chunks = append(sb.chunks, chunk)
	sb.buffer.WriteString(content)

	select {
	case sb.output <- content:
	case <-sb.ctx.Done():
	}
}

func (sb *StreamBuffer) Read() <-chan string {
	return sb.output
}

func (sb *StreamBuffer) GetComplete() string {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	return sb.buffer.String()
}

func (sb *StreamBuffer) Close() {
	sb.cancel()
	close(sb.output)
}

// AsyncStreamer handles concurrent streaming from multiple sources
type AsyncStreamer struct {
	buffers map[string]*StreamBuffer
	output  chan StreamChunk
	mutex   sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewAsyncStreamer(ctx context.Context) *AsyncStreamer {
	ctx, cancel := context.WithCancel(ctx)
	return &AsyncStreamer{
		buffers: make(map[string]*StreamBuffer),
		output:  make(chan StreamChunk, 1000),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (as *AsyncStreamer) CreateStream(id string) *StreamBuffer {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	buffer := NewStreamBuffer(as.ctx)
	as.buffers[id] = buffer

	// Forward chunks to main output
	go func() {
		for chunk := range buffer.output {
			select {
			case as.output <- StreamChunk{
				Content:   chunk,
				Timestamp: time.Now(),
				Source:    id,
			}:
			case <-as.ctx.Done():
				return
			}
		}
	}()

	return buffer
}

func (as *AsyncStreamer) GetStream(id string) (*StreamBuffer, bool) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	buffer, exists := as.buffers[id]
	return buffer, exists
}

func (as *AsyncStreamer) Output() <-chan StreamChunk {
	return as.output
}

func (as *AsyncStreamer) Close() {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	for _, buffer := range as.buffers {
		buffer.Close()
	}

	as.cancel()
	close(as.output)
}

// ResponseParser uses helper model to parse main model responses
type ResponseParser struct {
	helperModel Provider
	ctx         context.Context
}

func NewResponseParser(helperModel Provider) *ResponseParser {
	return &ResponseParser{
		helperModel: helperModel,
		ctx:         context.Background(),
	}
}

func (rp *ResponseParser) ShouldSaveToMemory(response string) bool {
	if rp.helperModel == nil {
		return false // Fallback to not saving
	}

	prompt := fmt.Sprintf(`<instruction>
Analyze if this response contains information worth saving to user memory.
Only respond "YES" or "NO" - no explanation.

Response to analyze: "%s"

Save to memory if:
- Contains important facts about user
- User preferences or settings
- Key decisions or conclusions
- Important context for future conversations

Do NOT save:
- Simple greetings or acknowledgments  
- Temporary/session-specific info
- General knowledge or explanations
</instruction>`, response)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	stream := make(chan string, 10)
	ctx, cancel := context.WithTimeout(rp.ctx, 5*time.Second)
	defer cancel()

	if err := rp.helperModel.Chat(ctx, messages, stream); err != nil {
		return false // Fail-safe
	}

	result := ""
	for chunk := range stream {
		result += chunk
	}

	return strings.Contains(strings.ToUpper(strings.TrimSpace(result)), "YES")
}

func (rp *ResponseParser) ExtractMemoryContent(response string) string {
	if rp.helperModel == nil {
		return response // Fallback to full response
	}

	prompt := fmt.Sprintf(`<instruction>
Extract only the key information worth saving to memory from this response.
Be concise but preserve important details.

Response: "%s"

Output only the extracted key information:
</instruction>`, response)

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	stream := make(chan string, 100)
	ctx, cancel := context.WithTimeout(rp.ctx, 10*time.Second)
	defer cancel()

	if err := rp.helperModel.Chat(ctx, messages, stream); err != nil {
		return response // Fallback
	}

	result := ""
	for chunk := range stream {
		result += chunk
	}

	return strings.TrimSpace(result)
}
