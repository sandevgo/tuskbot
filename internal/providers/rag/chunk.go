package rag

import (
	"strings"
	"sync"
	"unicode"

	"github.com/pkoukk/tiktoken-go"
)

var (
	tk     *tiktoken.Tiktoken
	tkOnce sync.Once
)

type Chunk struct {
	Text      string
	TokenSize int
	Index     int
}

type ChunkerConfig struct {
	MaxTokens     int
	OverlapTokens int
}

// E5BaseChunkerConfig config for e5-base-v2 model chunker.
// context size: 512 tokens, dimension: 768
func E5BaseChunkerConfig() ChunkerConfig {
	return ChunkerConfig{
		MaxTokens:     400,
		OverlapTokens: 50,
	}
}

func ChunkText(text string, cfg ChunkerConfig) []Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	// 1. Split into sentences (Unicode-aware)
	sentences := splitSentencesUnicode(text)

	// 2. Build chunks
	var chunks []Chunk
	var currentChunk strings.Builder
	currentTokens := 0
	chunkIndex := 0

	for i, sentence := range sentences {
		sentenceTokens := CountTokensUnicode(sentence)

		// Case A: Sentence is huge (larger than MaxTokens)
		if sentenceTokens > cfg.MaxTokens {
			// Flush current buffer if not empty
			if currentChunk.Len() > 0 {
				chunks = append(chunks, Chunk{
					Text:      strings.TrimSpace(currentChunk.String()),
					TokenSize: currentTokens,
					Index:     chunkIndex,
				})
				chunkIndex++
				currentChunk.Reset()
				currentTokens = 0
			}

			// Split the long sentence using pure token slicing
			subChunks := chunkLongTextUnicode(sentence, cfg.MaxTokens)
			for _, sc := range subChunks {
				chunks = append(chunks, Chunk{
					Text:      strings.TrimSpace(sc.Text),
					TokenSize: sc.TokenSize,
					Index:     chunkIndex,
				})
				chunkIndex++
			}
			continue
		}

		// Case B: Adding sentence exceeds limit -> Flush and start new chunk
		if currentTokens+sentenceTokens > cfg.MaxTokens && currentChunk.Len() > 0 {
			chunks = append(chunks, Chunk{
				Text:      strings.TrimSpace(currentChunk.String()),
				TokenSize: currentTokens,
				Index:     chunkIndex,
			})
			chunkIndex++

			// Overlap: get last N tokens from previous sentences
			overlap := getOverlapFromSentences(sentences, i, cfg.OverlapTokens)
			currentChunk.Reset()
			currentChunk.WriteString(overlap)
			currentTokens = CountTokensUnicode(overlap)
		}

		// Append sentence to buffer
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(sentence)
		currentTokens += sentenceTokens
	}

	// Flush remaining buffer
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			Text:      strings.TrimSpace(currentChunk.String()),
			TokenSize: currentTokens,
			Index:     chunkIndex,
		})
	}

	return chunks
}

// chunkLongTextUnicode splits a long string by encoding to tokens and slicing the array.
// This replaces the need for 'segment'.
func chunkLongTextUnicode(text string, maxTokens int) []Chunk {
	enc := getTokenizer()
	tokens := enc.Encode(text, nil, nil)

	var chunks []Chunk
	numTokens := len(tokens)

	for i := 0; i < numTokens; i += maxTokens {
		end := i + maxTokens
		if end > numTokens {
			end = numTokens
		}

		chunkTokens := tokens[i:end]
		chunkText := enc.Decode(chunkTokens)

		chunks = append(chunks, Chunk{
			Text:      chunkText,
			TokenSize: len(chunkTokens),
			// Index is handled by the caller
		})
	}

	return chunks
}

// splitSentencesUnicode splits text into sentences using Unicode rules.
func splitSentencesUnicode(text string) []string {
	// 1. Сначала разбиваем на параграфы
	paragraphs := splitParagraphs(text)

	// Разделители предложений для разных языков
	sentenceEnders := map[rune]bool{
		'.': true, '!': true, '?': true,
		'。': true, '！': true, '？': true, '．': true, '…': true,
	}

	var sentences []string

	// 2. Каждый параграф разбиваем на предложения
	for _, para := range paragraphs {
		var current strings.Builder
		runes := []rune(para)

		for i, r := range runes {
			current.WriteRune(r)

			if sentenceEnders[r] {
				// Проверяем, есть ли следующий символ и это пробел/конец/CJK
				if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) || isCJK(runes[i+1]) {
					s := strings.TrimSpace(current.String())
					if s != "" {
						sentences = append(sentences, s)
					}
					current.Reset()
				}
			}
		}

		// Остаток параграфа
		if s := strings.TrimSpace(current.String()); s != "" {
			sentences = append(sentences, s)
		}
	}

	// Если не нашли предложений — возвращаем весь текст
	if len(sentences) == 0 && text != "" {
		return []string{text}
	}

	return sentences
}

func splitParagraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n\n")

	var result []string
	for _, p := range parts {
		// Убираем одиночные переносы внутри параграфа (soft wrap)
		p = strings.ReplaceAll(p, "\n", " ")
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func getTokenizer() *tiktoken.Tiktoken {
	tkOnce.Do(func() {
		var err error
		tk, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			panic("failed to load tiktoken: " + err.Error())
		}
	})
	return tk
}

func CountTokensUnicode(text string) int {
	if text == "" {
		return 0
	}
	tokenIds := getTokenizer().Encode(text, nil, nil)
	return len(tokenIds)
}

func getOverlapFromSentences(sentences []string, currentIdx int, targetTokens int) string {
	if currentIdx == 0 {
		return ""
	}

	var overlap []string
	tokens := 0

	for i := currentIdx - 1; i >= 0 && tokens < targetTokens; i-- {
		sentTokens := CountTokensUnicode(sentences[i])
		overlap = append([]string{sentences[i]}, overlap...)
		tokens += sentTokens
	}

	return strings.Join(overlap, " ")
}

// isCJK — проверяет, является ли руна CJK символом
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}
