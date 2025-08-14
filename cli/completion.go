package cli

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	TokenNormal    TokenType = "normal"
	TokenExtension TokenType = "extension" // @nero
	TokenCommand   TokenType = "command"   // /help
	TokenResource  TokenType = "resource"  // #terminal
	TokenString    TokenType = "string"    // quoted text
	TokenKeyword   TokenType = "keyword"   // specific words
)

type Token struct {
	Type  TokenType
	Text  string
	Start int
	End   int
	Color Color
}

type SyntaxHighlighter struct {
	extensions []string
	commands   []string
	resources  []string
	keywords   []string
}

func NewSyntaxHighlighter() *SyntaxHighlighter {
	return &SyntaxHighlighter{
		extensions: []string{"nero", "system", "dev", "code"},
		commands:   []string{"/help", "/clear", "/status", "/quit", "/exit", "/config", "/spin", "/reset"},
		resources:  []string{"#terminal", "#screen", "#code", "#memory", "#config"},
		keywords:   []string{"full", "lite", "true", "false"},
	}
}

func (sh *SyntaxHighlighter) Tokenize(input string) []Token {
	var tokens []Token
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		start := i

		// Skip whitespace
		if unicode.IsSpace(runes[i]) {
			for i < len(runes) && unicode.IsSpace(runes[i]) {
				i++
			}
			tokens = append(tokens, Token{
				Type:  TokenNormal,
				Text:  string(runes[start:i]),
				Start: start,
				End:   i,
				Color: Reset,
			})
			continue
		}

		// Extension (@nero)
		if runes[i] == '@' {
			i++
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
				i++
			}
			text := string(runes[start:i])
			color := sh.getExtensionColor(text)
			tokens = append(tokens, Token{
				Type:  TokenExtension,
				Text:  text,
				Start: start,
				End:   i,
				Color: color,
			})
			continue
		}

		// Command (/help)
		if runes[i] == '/' {
			end := i + 1
			for end < len(runes) && (unicode.IsLetter(runes[end]) || unicode.IsDigit(runes[end])) {
				end++
			}
			text := string(runes[start:end])
			color := sh.getCommandColor(text)
			tokens = append(tokens, Token{
				Type:  TokenCommand,
				Text:  text,
				Start: start,
				End:   end,
				Color: color,
			})
			i = end
			continue
		}

		// Resource (#terminal)
		if runes[i] == '#' {
			end := i + 1
			for end < len(runes) && (unicode.IsLetter(runes[end]) || unicode.IsDigit(runes[end])) {
				end++
			}
			text := string(runes[start:end])
			color := sh.getResourceColor(text)
			tokens = append(tokens, Token{
				Type:  TokenResource,
				Text:  text,
				Start: start,
				End:   end,
				Color: color,
			})
			i = end
			continue
		}

		// Quoted strings
		if runes[i] == '"' || runes[i] == '\'' {
			quote := runes[i]
			i++
			for i < len(runes) && runes[i] != quote {
				if runes[i] == '\\' && i+1 < len(runes) {
					i += 2 // Skip escaped character
				} else {
					i++
				}
			}
			if i < len(runes) {
				i++ // Include closing quote
			}
			tokens = append(tokens, Token{
				Type:  TokenString,
				Text:  string(runes[start:i]),
				Start: start,
				End:   i,
				Color: Green,
			})
			continue
		}

		// Keywords and regular text
		if unicode.IsLetter(runes[i]) {
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_') {
				i++
			}
			text := string(runes[start:i])
			tokenType := TokenNormal
			color := Reset

			if sh.isKeyword(text) {
				tokenType = TokenKeyword
				color = Yellow
			}

			tokens = append(tokens, Token{
				Type:  tokenType,
				Text:  text,
				Start: start,
				End:   i,
				Color: color,
			})
			continue
		}

		// Everything else
		i++
		tokens = append(tokens, Token{
			Type:  TokenNormal,
			Text:  string(runes[start:i]),
			Start: start,
			End:   i,
			Color: Reset,
		})
	}

	return tokens
}

func (sh *SyntaxHighlighter) RenderTokens(tokens []Token) string {
	var result strings.Builder

	for _, token := range tokens {
		result.WriteString(string(token.Color))
		result.WriteString(token.Text)
		result.WriteString(string(Reset))
	}

	return result.String()
}

func (sh *SyntaxHighlighter) HighlightInput(input string) string {
	tokens := sh.Tokenize(input)
	return sh.RenderTokens(tokens)
}

func (sh *SyntaxHighlighter) getExtensionColor(ext string) Color {
	for _, known := range sh.extensions {
		if strings.HasSuffix(ext, known) {
			switch known {
			case "nero":
				return Magenta
			case "system":
				return Blue
			case "dev", "code":
				return Cyan
			default:
				return Yellow
			}
		}
	}
	return Red // Unknown extension
}

func (sh *SyntaxHighlighter) getCommandColor(cmd string) Color {
	for _, known := range sh.commands {
		if cmd == known {
			return Blue
		}
	}
	return Red // Unknown command
}

func (sh *SyntaxHighlighter) getResourceColor(res string) Color {
	for _, known := range sh.resources {
		if res == known {
			switch known {
			case "#terminal":
				return Green
			case "#screen":
				return Cyan
			case "#code":
				return Blue
			case "#memory":
				return Yellow
			case "#config":
				return Magenta
			default:
				return Cyan
			}
		}
	}
	return Red // Unknown resource
}

func (sh *SyntaxHighlighter) isKeyword(word string) bool {
	for _, keyword := range sh.keywords {
		if word == keyword {
			return true
		}
	}
	return false
}

type AutoCompleter struct {
	extensions []string
	commands   []string
	resources  []string
	history    []string
}

func NewAutoCompleter() *AutoCompleter {
	return &AutoCompleter{
		extensions: []string{"@nero", "@system", "@dev", "@code"},
		commands:   []string{"/help", "/clear", "/status", "/quit", "/exit"},
		resources:  []string{"#terminal", "#screen", "#code", "#memory", "#config"},
		history:    make([]string, 0),
	}
}

func (ac *AutoCompleter) GetSuggestions(input string, cursorPos int) []string {
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	// Find the current word being typed
	start := cursorPos
	for start > 0 && !unicode.IsSpace(rune(input[start-1])) {
		start--
	}

	currentWord := input[start:cursorPos]

	var suggestions []string

	// Extension suggestions
	if strings.HasPrefix(currentWord, "@") {
		for _, ext := range ac.extensions {
			if strings.HasPrefix(ext, currentWord) {
				suggestions = append(suggestions, ext)
			}
		}
	}

	// Command suggestions
	if strings.HasPrefix(currentWord, "/") {
		for _, cmd := range ac.commands {
			if strings.HasPrefix(cmd, currentWord) {
				suggestions = append(suggestions, cmd)
			}
		}
	}

	// Resource suggestions
	if strings.HasPrefix(currentWord, "#") {
		for _, res := range ac.resources {
			if strings.HasPrefix(res, currentWord) {
				suggestions = append(suggestions, res)
			}
		}
	}

	return suggestions
}

func (ac *AutoCompleter) AddToHistory(input string) {
	if input != "" {
		ac.history = append(ac.history, input)
		if len(ac.history) > 100 {
			ac.history = ac.history[1:]
		}
	}
}

func (ac *AutoCompleter) ShowInlineSuggestion(input string, cursorPos int) string {
	suggestions := ac.GetSuggestions(input, cursorPos)
	if len(suggestions) > 0 {
		// Show first suggestion as faded text
		suggestion := suggestions[0]
		if cursorPos < len(input) {
			return suggestion[len(input[strings.LastIndex(input[:cursorPos], " ")+1:cursorPos]):]
		}
		return suggestion[len(input[strings.LastIndex(input, " ")+1:]):]
	}
	return ""
}
