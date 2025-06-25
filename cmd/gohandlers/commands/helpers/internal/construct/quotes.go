package construct

import "fmt"

func quotes(s string) string {
	return fmt.Sprintf("%q", s)
}
