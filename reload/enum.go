package reload

// Max 255 templates. Could bump it up to uint16 if needed...
type TmplName uint8 // template name enum

//go:generate stringer -type=TmplName
const (
	// SiteRouter
	Index TmplName = iota
	About
	Login
	Create
	Signup
	ThankYou
	ErrorPage
)
