/*
	LoginData is here because of import loops. Stupid, yes, but I couldn't
	find a better way to do it.
*/
package datatypes

// Data returned to /login/ after validation.
type LoginData struct {
	Host string // FQDN

	CSRF []byte

	// Form validation errors
	Error string

	// Redirection
	Redir string
}
