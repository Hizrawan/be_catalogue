package middlewares

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"be20250107/internal/modules/filestore"

	"be20250107/internal/modules/ratelimiter"

	"be20250107/internal/app"
	httperr "be20250107/internal/errors"
	"be20250107/internal/responses"

	"github.com/charmbracelet/lipgloss"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"golang.org/x/term"
)

func Recover(app *app.Registry) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil && rec != http.ErrAbortHandler {
					details := rec
					if err, ok := rec.(error); ok {
						if errors.Is(err, context.Canceled) {
							return
						}

						if errors.Is(err, httperr.ErrUnauthenticated) {
							responses.Unauthenticated(w)
							return
						}

						if errors.Is(err, httperr.ErrForbidden) {
							responses.Forbidden(w)
							return
						}

						if errors.Is(err, httperr.ErrNotFound) || errors.Is(err, sql.ErrNoRows) || errors.Is(err, filestore.ErrFileNotExist) {
							responses.NotFound(w, nil)
							return
						}

						if err, ok := err.(validation.Errors); ok {
							responses.ValidationError(w, err)
							return
						}

						if errors.Is(err, httperr.ErrMalformedRequest) {
							responses.MalformedRequest(w)
							return
						}

						if err, ok := err.(httperr.ErrUnprocessableEntity); ok {
							responses.UnprocessableEntity(w, err)
							return
						}

						if errors.Is(err, httperr.ErrServiceUnavailable) {
							responses.ServiceUnavailable(w)
							return
						}

						if errors.Is(err, httperr.ErrTooManyRequests) || errors.Is(err, ratelimiter.ErrRateLimited) {
							responses.TooManyRequests(w)
							return
						}

						details = err.Error()
					}

					app.Log.Error(fmt.Sprint(details))

					if app.Config.Debug {
						printStackTrace(rec)
						responses.InternalServerError(w, details)
					} else {
						responses.InternalServerError(w, nil)
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

var timeStyle = lipgloss.NewStyle().
	Padding(0, 2).
	Background(lipgloss.Color("#14161a"))

var panicMsgStyle = lipgloss.NewStyle().
	PaddingLeft(2).
	PaddingRight(1).
	Align(lipgloss.Right).
	Background(lipgloss.Color("#8D0000")).
	Foreground(lipgloss.Color("#FFF"))

var errorMsgStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Foreground(lipgloss.Color("#FFF")).
	Background(lipgloss.Color("#14161a"))

var stackTraceBase = lipgloss.NewStyle().
	Background(lipgloss.Color("#14161a"))

var stackTraceTitle = stackTraceBase.Copy().
	PaddingLeft(2).
	PaddingTop(1).PaddingBottom(1).
	Foreground(lipgloss.Color("#FFF")).
	Background(lipgloss.Color("#14161a"))

var stackTraceLineStyle = stackTraceBase.Copy().
	PaddingLeft(4)

var stackTraceLinkStyle = stackTraceBase.Copy().
	PaddingLeft(8).
	Foreground(lipgloss.Color("#9d9d9d"))

func printStackTrace(err any) {
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	dbg := debug.Stack()
	fmt.Println(timeStyle.Width(w).Render("\n[" + time.Now().Format(time.RFC3339) + "]\n"))
	infoBox := panicMsgStyle.Render("PANIC")
	fmt.Printf("%s%s\n",
		infoBox,
		errorMsgStyle.Width(w-lipgloss.Width(infoBox)).Render(fmt.Sprintf(": %v", err)),
	)

	fmt.Println(stackTraceTitle.Width(w).Render("This error is located at:"))
	stack := strings.Split(string(dbg), "\n")
	panicLine := 0
	for i := 0; i < len(stack); i++ {
		if strings.HasPrefix(stack[i], "panic(") {
			panicLine = i
			break
		}
	}
	stack = stack[panicLine+2:]
	for i, l := range stack {
		l = fmt.Sprintf("%v", l)
		if i%2 == 0 {
			l = stackTraceLineStyle.Width(w).Render(styleFunctionCall(l))
		} else {
			l = stackTraceLinkStyle.Width(w).Render(strings.TrimSpace(l) + "\n")
		}
		fmt.Println(l)
	}
}

var moduleStyle = stackTraceBase.Copy().
	Foreground(lipgloss.Color("#6a93a7"))
var functionStyle = stackTraceBase.Copy().
	Foreground(lipgloss.Color("#e4b46d"))
var paramsStyle = stackTraceBase.Copy().
	Foreground(lipgloss.Color("#6d6d6d"))

func styleFunctionCall(call string) string {
	paramIdx := -1
	modIdx := -1
	for i := len(call) - 1; i > 0; i-- {
		if paramIdx == -1 && call[i] == '(' {
			paramIdx = i
		} else if paramIdx == -1 && call[i] == '.' {
			paramIdx = len(call)
		}
		if modIdx == -1 && call[i] == '/' {
			modIdx = i + 1
		}
		if modIdx != -1 && paramIdx != -1 {
			break
		}
	}
	if modIdx == -1 {
		modIdx = 0
	}
	if paramIdx == -1 {
		paramIdx = len(call)
	}
	module := call[0:modIdx]
	fnName := call[modIdx:paramIdx]
	params := call[paramIdx:]
	return fmt.Sprintf(
		"%s%s%s",
		moduleStyle.Render(module),
		functionStyle.Render(fnName),
		paramsStyle.Render(params),
	)
}

func NotFound(app *app.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u *url.URL
		if app.Config.Debug {
			u = r.URL
		}
		responses.NotFound(w, u)
	}
}
