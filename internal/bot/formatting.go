package bot

import (
	"regexp"
	"strings"
)

// formatForMarkdownV2 processes text to be compatible with Telegram's MarkdownV2 format
func (b *Bot) formatForMarkdownV2(text string) string {
	// First, let's handle code blocks to protect them from escaping
	codeBlockRegex := regexp.MustCompile("```([^`]*)```")
	codeBlocks := make(map[string]string)
	blockIndex := 0
	
	// Extract and store code blocks
	text = codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := "___CODEBLOCK_" + string(rune(blockIndex)) + "___"
		codeBlocks[placeholder] = match
		blockIndex++
		return placeholder
	})
	
	// Handle inline code to protect it from escaping
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	inlineCodes := make(map[string]string)
	codeIndex := 0
	
	text = inlineCodeRegex.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := "___INLINECODE_" + string(rune(codeIndex)) + "___"
		inlineCodes[placeholder] = match
		codeIndex++
		return placeholder
	})
	
	// Convert markdown tables to better format for Telegram
	text = b.convertTablesToMarkdownV2(text)
	
	// Convert headers to bold formatting
	text = b.convertHeadersToMarkdownV2(text)
	
	// Escape special characters for MarkdownV2
	text = b.escapeMarkdownV2(text)
	
	// Restore code blocks and inline codes
	for placeholder, original := range codeBlocks {
		// Escape backticks and backslashes in code blocks
		escaped := strings.ReplaceAll(original, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		text = strings.ReplaceAll(text, placeholder, escaped)
	}
	
	for placeholder, original := range inlineCodes {
		// Escape backticks and backslashes in inline code
		escaped := strings.ReplaceAll(original, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")
		text = strings.ReplaceAll(text, placeholder, escaped)
	}
	
	return text
}

// escapeMarkdownV2 escapes special characters for MarkdownV2
func (b *Bot) escapeMarkdownV2(text string) string {
	// Characters that need escaping in MarkdownV2:
	// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	
	return text
}

// convertTablesToMarkdownV2 converts markdown tables to a format that works better in Telegram
func (b *Bot) convertTablesToMarkdownV2(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inTable := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if this is a table line (starts and ends with |)
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && len(trimmed) > 2 {
			// Check if next line is a separator (contains | and -)
			isHeader := false
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.Contains(nextLine, "-") && strings.Contains(nextLine, "|") {
					isHeader = true
				}
			}
			
			if !inTable {
				inTable = true
				result = append(result, "") // Add blank line before table
			}
			
			// Convert table row to formatted text
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			var formattedCells []string
			
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if isHeader && cell != "" {
					formattedCells = append(formattedCells, "*"+cell+"*") // Bold headers
				} else {
					formattedCells = append(formattedCells, cell)
				}
			}
			
			result = append(result, "• "+strings.Join(formattedCells, " | "))
			
			// Skip the separator line for headers
			if isHeader && i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.Contains(nextLine, "-") && strings.Contains(nextLine, "|") {
					i++ // Skip separator line
					continue
				}
			}
		} else {
			if inTable {
				inTable = false
				result = append(result, "") // Add blank line after table
			}
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

// convertHeadersToMarkdownV2 converts markdown headers to bold text
func (b *Bot) convertHeadersToMarkdownV2(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Handle headers (# ## ### etc.)
		if strings.HasPrefix(trimmed, "#") {
			// Count the number of # characters
			headerLevel := 0
			for _, char := range trimmed {
				if char == '#' {
					headerLevel++
				} else {
					break
				}
			}
			
			// Extract header text
			headerText := strings.TrimSpace(trimmed[headerLevel:])
			
			if headerText != "" {
				// Convert to bold formatting based on level
				switch headerLevel {
				case 1:
					result = append(result, "*"+headerText+"*")
				case 2:
					result = append(result, "*"+headerText+"*")
				default:
					result = append(result, "*"+headerText+"*")
				}
			}
		} else {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

// createSystemMessageForMarkdownV2 creates an appropriate system message for MarkdownV2 formatting
func (b *Bot) createSystemMessageForMarkdownV2() string {
	return `You are a helpful assistant. Format your responses appropriately for Telegram using these guidelines:

FORMATTING RULES:
- Use *bold text* for emphasis and headers
- Use _italic text_ for subtle emphasis  
- Use ` + "`inline code`" + ` for commands, variables, file names
- Use ` + "```code blocks```" + ` for multi-line code
- Use >quoted text for important notes or citations

TABLES: Instead of markdown tables, use formatted lists:
• *Header 1* | *Header 2* | *Header 3*
• Value 1 | Value 2 | Value 3
• Value A | Value B | Value C

LISTS: Use bullet points (•) or numbers (1.)

AVOID: Don't use these characters unnecessarily: # + - = | { } . ! [ ] ( ) ~ 

Keep responses clear, well-structured, and easy to read on mobile devices.`
} 