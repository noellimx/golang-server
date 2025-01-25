// Package main
package mux

import (
	"net/http"

	"golang-server/cmd/servers/auth/handlers"
	"golang-server/src/domain/auth"
	"golang-server/src/infrastructure/messaging/email"
)

type AuthMuxOpts struct {
	Auth auth.AuthService
	Mail email.EmailService
}

type MuxOpts struct {
	*AuthMuxOpts
}

// The closes entry point sans sockets.
func NewMux(opts *MuxOpts) *http.ServeMux {
	if opts == nil {
		opts = &MuxOpts{
			AuthMuxOpts: nil,
		}
	}

	mux := http.NewServeMux()

	{

		helloMux := http.NewServeMux()

		hello := handlers.Hello()
		helloMux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/" {

				http.NotFound(writer, request)
			} else {
				hello(writer, request)
			}
		})

		mux.Handle("/", helloMux)
	}

	if opts.AuthMuxOpts != nil {
		authHandlers, err := handlers.NewAuthHandler(opts.Auth, opts.Mail)
		if err != nil {
			panic(err)
		}
		mux.HandleFunc("POST /register/", authHandlers.RegisterUsernamePassword())
		mux.HandleFunc("POST /token/", authHandlers.AuthByUsernamePassword())
		mux.HandleFunc("POST /otp/", authHandlers.SubmitOtp())

		mustAuth := MiddleWare(nil).Wrap(func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				next(writer, request)
			}
		})
		mux.HandleFunc("PATCH /user/", mustAuth(authHandlers.PatchUser()))
	}

	return mux
}

type MiddleWare func(http.HandlerFunc) http.HandlerFunc

// Wrap
// .Wrap(f1,f2,f3) => f1 => f2 => f3
func (mw MiddleWare) Wrap(nexts ...MiddleWare) MiddleWare {
	for _, next := range nexts {
		mw = mw.wrap(next)
	}
	return mw
}

func (mw MiddleWare) wrap(next MiddleWare) MiddleWare {
	if mw == nil {
		return next
	}
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return next(mw(handlerFunc))
	}
}
