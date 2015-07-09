package paths

// PQDN is the partially-qualified domain name for the site. This is in
// contrast to the FQDN which includes subdomains.
const PQDN = "localhost"

// Different URL paths.
const (
	// Index-rooted paths.
	IndexPath    = "/"
	AboutPath    = "/about/"
	SignupPath   = "/signup/"
	CreatePath   = "/create/"
	LoginPath    = "/login/"
	LogoutPath   = "/logout/"
	ThankYouPath = "/thanks/"

	DashboardPath = "/dashboard/"
)

var (
	TemplatePath = "templates"
)
