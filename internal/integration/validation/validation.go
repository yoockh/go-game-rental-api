package integration

import (
	"fmt"
	"os"
)

// Validate performs fail-fast validation of critical integration environment variables
func Validate() error {
	var missing []string

	// Supabase Storage
	if os.Getenv("SUPABASE_URL") == "" {
		missing = append(missing, "SUPABASE_URL")
	}
	if os.Getenv("SUPABASE_SERVICE_KEY") == "" {
		missing = append(missing, "SUPABASE_SERVICE_KEY")
	}
	if os.Getenv("SUPABASE_STORAGE_BUCKET") == "" {
		missing = append(missing, "SUPABASE_STORAGE_BUCKET")
	}

	// SendGrid
	if os.Getenv("SENDGRID_API_KEY") == "" {
		missing = append(missing, "SENDGRID_API_KEY")
	}
	if os.Getenv("SENDGRID_FROM_EMAIL") == "" {
		missing = append(missing, "SENDGRID_FROM_EMAIL")
	}

	// Midtrans
	if os.Getenv("MIDTRANS_SERVER_KEY") == "" {
		missing = append(missing, "MIDTRANS_SERVER_KEY")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing critical environment variables: %v", missing)
	}

	return nil
}
