package core

type InternalSource string

const (
	AppSource      InternalSource = "App"
	RendererSource InternalSource = "Renderer"
	InputSource    InternalSource = "Input"
)
