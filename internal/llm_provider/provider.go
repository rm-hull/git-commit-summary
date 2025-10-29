package llmprovider

type Provider interface {
	Call(systemPrompt, userPrompt string) (string, error)
	Model() string
}
