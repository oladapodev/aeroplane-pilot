package llm

// Generator is the minimal text-generation interface used by the core router.
type Generator interface {
	Generate(prompt string) (string, error)
}
