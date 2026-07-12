package controller

// Docs renders the AI-oriented architecture instruction (GOLANGARCH.md content).
func (s *Controller) Docs() string {
	return s.templater.Run()
}
