package openai

import "fmt"

// based on https://wfhbrian.com/the-best-way-to-summarize-a-paragraph-using-gpt-3/
const summaryPrompt = `
We introduce Extreme TLDR generation, a new form of extreme summarization for paragraphs. TLDR generation involves high source compression, removes stop words and summarizes the paragraph whilst retaining meaning. The result is the shortest possible summary that retains all of the original meaning and context of the paragraph.
Don't show "Translated from" part of the response. Should keep feeling context.
Example
Paragraph:
%s
Extreme TLDR:
`

func summarize(message string) (string, error) {
	messages := []GPTMessage{
		{
			Role:    "user",
			Content: fmt.Sprintf(summaryPrompt, message),
		},
	}
	resps, err := completionAPI(messages, "gpt-3.5-turbo")
	if err != nil {
		return "", err
	}

	return resps.Choices[0].Message.Content, nil
}
