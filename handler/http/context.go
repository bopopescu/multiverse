package http

import (
	"golang.org/x/net/context"

	v04_entity "github.com/tapglue/multiverse/v04/entity"
)

const (
	ctxKeyApp       = "app"
	ctxKeyLogger    = "logger"
	ctxKeyRoute     = "route"
	ctxKeyTokenType = "tokenType"
	ctxKeyUser      = "user"
	ctxKeyVersion   = "version"

	tokenApplication = "application"
	tokenBackend     = "backend"
)

func appFromContext(ctx context.Context) *v04_entity.Application {
	return ctx.Value(ctxKeyApp).(*v04_entity.Application)
}

func appInContext(ctx context.Context, app *v04_entity.Application) context.Context {
	return context.WithValue(ctx, ctxKeyApp, app)
}

func routeFromContext(ctx context.Context) string {
	return ctx.Value(ctxKeyRoute).(string)
}

func routeInContext(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, ctxKeyRoute, route)
}

func tokenFromContext(ctx context.Context) string {
	return ctx.Value(ctxKeyTokenType).(string)
}

func tokenTypeInContext(ctx context.Context, tokenType string) context.Context {
	return context.WithValue(ctx, ctxKeyTokenType, tokenType)
}

func userFromContext(ctx context.Context) *v04_entity.ApplicationUser {
	return ctx.Value(ctxKeyUser).(*v04_entity.ApplicationUser)
}

func userInContext(ctx context.Context, user *v04_entity.ApplicationUser) context.Context {
	return context.WithValue(ctx, ctxKeyUser, user)
}

func versionFromContext(ctx context.Context) string {
	return ctx.Value(ctxKeyVersion).(string)
}

func versionInContext(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, ctxKeyVersion, version)
}