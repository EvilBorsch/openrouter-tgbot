package bot

import (
	"strings"
)

// formatForMarkdownV2 processes text to be compatible with Telegram's MarkdownV2 format
func (b *Bot) formatForMarkdownV2(text string) string {
	// Preserve text structure and only escape truly problematic characters

	// Convert tables first
	text = b.convertTablesToMarkdownV2(text)

	// Convert headers to bold formatting
	text = b.convertHeadersToMarkdownV2(text)

	// Only escape characters that are actually breaking parsing
	// Be very conservative to preserve text flow and structure
	text = b.escapeConservatively(text)

	return text
}

// escapeConservatively only escapes characters that actually cause parsing failures
func (b *Bot) escapeConservatively(text string) string {
	// MINIMAL escaping - only handle characters that cause actual errors
	// Preserve all punctuation and structure for natural text flow

	// Only escape characters that are confirmed to break parsing
	// Based on actual error: "Character '-' is reserved"
	text = strings.ReplaceAll(text, "-", "\\-")

	// Escape backslashes to prevent double escaping issues
	// Note: Do this after other escaping to avoid double-escaping
	text = strings.ReplaceAll(text, "\\\\-", "\\-") // Fix double escaping of dashes

	return text
}

// convertTablesToMarkdown converts tables for regular Markdown (no escaping needed)
func (b *Bot) convertTablesToMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Simple table detection: starts and ends with |
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && len(trimmed) > 2 {
			// Check if this might be a header (next line has dashes)
			isHeader := false
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.Contains(nextLine, "-") && strings.Contains(nextLine, "|") {
					isHeader = true
				}
			}

			// Convert to bullet point format
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			var cleanCells []string

			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if cell != "" {
					if isHeader {
						cleanCells = append(cleanCells, "**"+cell+"**") // Use ** for bold in regular Markdown
					} else {
						cleanCells = append(cleanCells, cell)
					}
				}
			}

			if len(cleanCells) > 0 {
				result = append(result, "• "+strings.Join(cleanCells, " | "))
			}
		} else if strings.Contains(trimmed, "|") && strings.Contains(trimmed, "-") && len(strings.TrimSpace(trimmed)) > 0 {
			// Skip separator lines
			continue
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// convertTablesToMarkdownV2 converts markdown tables to a format that works better in Telegram
func (b *Bot) convertTablesToMarkdownV2(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Simple table detection: starts and ends with |
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && len(trimmed) > 2 {
			// Check if this might be a header (next line has dashes)
			isHeader := false
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.Contains(nextLine, "-") && strings.Contains(nextLine, "|") {
					isHeader = true
				}
			}

			// Convert to bullet point format
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			var cleanCells []string

			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if cell != "" {
					if isHeader {
						cleanCells = append(cleanCells, "*"+cell+"*")
					} else {
						cleanCells = append(cleanCells, cell)
					}
				}
			}

			if len(cleanCells) > 0 {
				result = append(result, "• "+strings.Join(cleanCells, " — "))
			}
		} else if strings.Contains(trimmed, "|") && strings.Contains(trimmed, "-") && len(strings.TrimSpace(trimmed)) > 0 {
			// Skip separator lines
			continue
		} else {
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

// createSystemMessageForMarkdown creates an appropriate system message for regular Markdown formatting
func (b *Bot) createSystemMessageForMarkdown() string {
	return `You are a helpful assistant. Format your responses for Telegram using simple, clean formatting:

FORMATTING:
- Use **bold text** for important points and headers  
- Use *italic text* for emphasis
- Use ` + "`inline code`" + ` for commands and technical terms
- Use ` + "```code blocks```" + ` for longer code

STRUCTURE:
- Use bullet points (•) for lists
- Keep paragraphs short and readable
- Use blank lines to separate sections

FOR TABLES: Use simple bullet-point format instead of markdown tables:
• **Column 1** | **Column 2** | **Column 3**  
• Data 1 | Data 2 | Data 3
• Data A | Data B | Data C

Keep responses clear and mobile-friendly. Preserve line breaks and structure.`
}

// createSystemMessageForMarkdownV2 creates an appropriate system message for MarkdownV2 formatting
func (b *Bot) createSystemMessageForMarkdownV2() string {
	return `You are a helpful assistant. Format your responses for Telegram using simple, clean formatting:

FORMATTING:
- Use *bold text* for important points and headers  
- Use _italic text_ for emphasis
- Use ` + "`inline code`" + ` for commands and technical terms
- Use ` + "```code blocks```" + ` for longer code

STRUCTURE:
- Use bullet points (•) for lists
- Keep paragraphs short and readable
- Use blank lines to separate sections

FOR TABLES: Use simple bullet-point format instead of markdown tables:
• *Column 1* — *Column 2* — *Column 3*  
• Data 1 — Data 2 — Data 3
• Data A — Data B — Data C

Keep responses clear and mobile-friendly. Avoid complex formatting.`
}
