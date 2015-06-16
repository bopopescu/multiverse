/**
 * @author Florin Patan <florinpatan@gmail.com>
 */

package kinesis

import (
	"github.com/tapglue/backend/context"
	"github.com/tapglue/backend/errors"
	"github.com/tapglue/backend/v02/core"
	"github.com/tapglue/backend/v02/errmsg"
	"github.com/tapglue/backend/v02/server"
)

type (
	application struct {
		storage core.Application
	}
)

func (app *application) Read(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) Update(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) Delete(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) Create(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) List(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) PopulateContext(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

func (app *application) PopulateContextFromID(ctx *context.Context) (err []errors.Error) {
	return []errors.Error{errmsg.ErrServerNotImplementedYet}
}

// NewApplication returns a new application route handler
func NewApplication(storage core.Application) server.Application {
	return &application{
		storage: storage,
	}
}
