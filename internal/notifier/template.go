package notifier

import (
	"fmt"
	"strings"

	"jobradar/internal/model"
)

// FormatTelegramMessage formats a job notification for Telegram
func FormatTelegramMessage(matched *model.MatchedJob) string {
	job := matched.Job

	var sb strings.Builder

	sb.WriteString("🔔 *New Job Match\\!*\n\n")
	sb.WriteString(fmt.Sprintf("📋 *%s*\n", escapeMD(job.Title)))
	sb.WriteString(fmt.Sprintf("💰 %s\n", escapeMD(job.BudgetDisplay())))

	if job.Proposals != nil {
		sb.WriteString(fmt.Sprintf("👥 Proposals: %d\n", *job.Proposals))
	} else {
		sb.WriteString("👥 Proposals: N/A\n")
	}

	sb.WriteString(fmt.Sprintf("⏰ Posted: %s\n", escapeMD(job.PostedAgo())))

	if len(job.Skills) > 0 {
		skills := job.Skills
		if len(skills) > 5 {
			skills = skills[:5]
		}
		sb.WriteString(fmt.Sprintf("🏷️ Skills: %s\n", escapeMD(strings.Join(skills, ", "))))
	}

	// Description summary (max 200 chars)
	desc := job.Description
	if len(desc) > 200 {
		desc = desc[:200] + "..."
	}
	sb.WriteString(fmt.Sprintf("\n📝 %s\n", escapeMD(desc)))

	sb.WriteString(fmt.Sprintf("\n🔗 [View Job](%s)\n", job.URL))
	sb.WriteString(fmt.Sprintf("\n\\-\\-\\-\n✅ Matched: %s", escapeMD(strings.Join(matched.MatchedKeywords, ", "))))

	return sb.String()
}

// escapeMD escapes special characters for Telegram MarkdownV2
func escapeMD(text string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range chars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
}

// FormatEmailSubject formats the email subject
func FormatEmailSubject(matched *model.MatchedJob) string {
	title := matched.Job.Title
	if len(title) > 50 {
		title = title[:50] + "..."
	}
	return fmt.Sprintf("[JobRadar] New Match: %s", title)
}

// FormatEmailBody formats the email body as HTML
func FormatEmailBody(matched *model.MatchedJob) string {
	job := matched.Job

	skillsHTML := ""
	if len(job.Skills) > 0 {
		skillsHTML = fmt.Sprintf("<strong>🏷️ Skills:</strong> %s<br/>", strings.Join(job.Skills, ", "))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        h2 { color: #2c5282; }
        h3 { color: #4a5568; margin-bottom: 10px; }
        .info { background: #f7fafc; padding: 15px; border-radius: 8px; margin: 15px 0; }
        .description { background: #fff; border-left: 4px solid #4299e1; padding: 15px; margin: 15px 0; }
        .button { display: inline-block; background: #4299e1; color: white; padding: 12px 24px; 
                  text-decoration: none; border-radius: 6px; margin: 15px 0; }
        .footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #e2e8f0; 
                  font-size: 12px; color: #718096; }
    </style>
</head>
<body>
    <div class="container">
        <h2>🔔 New Job Match!</h2>
        
        <h3>%s</h3>
        
        <div class="info">
            <strong>💰 Budget:</strong> %s<br/>
            <strong>👥 Proposals:</strong> %s<br/>
            <strong>⏰ Posted:</strong> %s<br/>
            %s
        </div>
        
        <div class="description">
            <strong>📝 Description:</strong><br/>
            %s
        </div>
        
        <a href="%s" class="button">🔗 View Job on Upwork</a>
        
        <div class="footer">
            <p>Matched keywords: %s</p>
            <p>This notification was sent by JobRadar.</p>
        </div>
    </div>
</body>
</html>`,
		escapeHTML(job.Title),
		escapeHTML(job.BudgetDisplay()),
		formatProposals(job.Proposals),
		escapeHTML(job.PostedAgo()),
		skillsHTML,
		escapeHTML(truncate(job.Description, 500)),
		job.URL,
		escapeHTML(strings.Join(matched.MatchedKeywords, ", ")),
	)
}

// formatProposals formats the proposal count
func formatProposals(p *int) string {
	if p == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *p)
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// FormatTestMessage creates a test notification message
func FormatTestMessage() string {
	return `🔔 *JobRadar Test Notification*

This is a test message to verify your notification settings are working correctly\.

If you received this message, your configuration is correct\!

✅ Telegram Bot: Connected
✅ Chat ID: Verified`
}
