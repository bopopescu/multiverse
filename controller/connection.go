package controller

import (
	"sort"

	"github.com/tapglue/multiverse/service/app"
	"github.com/tapglue/multiverse/service/connection"
	"github.com/tapglue/multiverse/service/user"
)

// ConnectionFeed is the composite to transport information relevant for
// connections.
type ConnectionFeed struct {
	Connections connection.List
	Users       user.List
	UserMap     user.Map
}

// ConnectionController bundles the business constraints of Connections.
type ConnectionController struct {
	connections connection.Service
	users       user.Service
}

// NewConnectionController returns a controller instance.
func NewConnectionController(
	connections connection.Service,
	users user.Service,
) *ConnectionController {
	return &ConnectionController{
		connections: connections,
		users:       users,
	}
}

// ByState returns all connections for the given origin and state.
func (c *ConnectionController) ByState(
	currentApp *app.App,
	originID uint64,
	state connection.State,
) (*ConnectionFeed, error) {
	switch state {
	case connection.StatePending, connection.StateConfirmed, connection.StateRejected:
		// valid
	default:
		return nil, wrapError(ErrInvalidEntity, "unsupported state %s", string(state))
	}

	ics, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Enabled: &defaultEnabled,
		FromIDs: []uint64{originID},
		States:  []connection.State{state},
	})
	if err != nil {
		return nil, err
	}

	ocs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Enabled: &defaultEnabled,
		States:  []connection.State{state},
		ToIDs:   []uint64{originID},
	})
	if err != nil {
		return nil, err
	}

	um, err := user.MapFromIDs(
		c.users,
		currentApp.Namespace(),
		append(ics.ToIDs(), ocs.FromIDs()...)...,
	)
	if err != nil {
		return nil, err
	}

	return &ConnectionFeed{
		Connections: append(ics, ocs...),
		UserMap:     um,
	}, nil
}

// CreateSocial connects the origin with the users matching the platform ids.
func (c *ConnectionController) CreateSocial(
	currentApp *app.App,
	originID uint64,
	connectionType connection.Type,
	connectionState connection.State,
	platform string,
	connectionIDs ...string,
) (user.List, error) {
	us, err := c.users.Query(currentApp.Namespace(), user.QueryOptions{
		Enabled: &defaultEnabled,
		SocialIDs: map[string][]string{
			platform: connectionIDs,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, u := range us {
		_, err := c.connections.Put(currentApp.Namespace(), &connection.Connection{
			Enabled: true,
			FromID:  originID,
			ToID:    u.ID,
			State:   connectionState,
			Type:    connectionType,
		})
		if err != nil {
			return nil, err
		}

		r, err := queryRelation(c.connections, currentApp, originID, u.ID)
		if err != nil {
			return nil, err
		}

		u.IsFollower = r.isFollower
		u.IsFollowing = r.isFollowing
		u.IsFriend = r.isFriend
	}

	return us, nil
}

// Delete disables the given connection.
func (c *ConnectionController) Delete(
	currentApp *app.App,
	con *connection.Connection,
) error {
	var (
		fromIDs = []uint64{con.FromID}
		toIDs   = []uint64{con.ToID}
	)

	if con.Type == connection.TypeFriend {
		fromIDs = []uint64{con.FromID, con.ToID}
		toIDs = []uint64{con.FromID, con.ToID}
	}

	cs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Enabled: &defaultEnabled,
		FromIDs: fromIDs,
		Limit:   1,
		ToIDs:   toIDs,
		Types:   []connection.Type{con.Type},
	})
	if err != nil {
		return err
	}

	if len(cs) == 0 {
		return nil
	}

	con = cs[0]

	con.Enabled = false

	_, err = c.connections.Put(currentApp.Namespace(), con)

	return err
}

// Followers returns the list of users who follow the origin.
func (c *ConnectionController) Followers(
	currentApp *app.App,
	origin uint64,
	userID uint64,
	opts connection.QueryOptions,
) (*ConnectionFeed, error) {
	cs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Before:  opts.Before,
		Enabled: &defaultEnabled,
		Limit:   opts.Limit,
		ToIDs:   []uint64{userID},
		States:  []connection.State{connection.StateConfirmed},
		Types:   []connection.Type{connection.TypeFollow},
	})
	if err != nil {
		return nil, err
	}

	us, err := user.ListFromIDs(c.users, currentApp.Namespace(), cs.FromIDs()...)
	if err != nil {
		return nil, err
	}

	for _, u := range us {
		err := enrichConnectionCounts(c.connections, c.users, currentApp, u)
		if err != nil {
			return nil, err
		}

		err = enrichRelation(c.connections, currentApp, origin, u)
		if err != nil {
			return nil, err
		}
	}

	return &ConnectionFeed{
		Connections: cs,
		Users:       us,
	}, nil
}

// Followings returns the list of users the origin is following.
func (c *ConnectionController) Followings(
	currentApp *app.App,
	origin uint64,
	userID uint64,
	opts connection.QueryOptions,
) (*ConnectionFeed, error) {
	cs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Before:  opts.Before,
		Enabled: &defaultEnabled,
		FromIDs: []uint64{userID},
		Limit:   opts.Limit,
		States:  []connection.State{connection.StateConfirmed},
		Types:   []connection.Type{connection.TypeFollow},
	})
	if err != nil {
		return nil, err
	}

	us, err := user.ListFromIDs(c.users, currentApp.Namespace(), cs.ToIDs()...)
	if err != nil {
		return nil, err
	}

	for _, u := range us {
		err := enrichConnectionCounts(c.connections, c.users, currentApp, u)
		if err != nil {
			return nil, err
		}

		err = enrichRelation(c.connections, currentApp, origin, u)
		if err != nil {
			return nil, err
		}
	}

	return &ConnectionFeed{
		Connections: cs,
		Users:       us,
	}, nil
}

// Friends returns the list of users the origin is friends with.
func (c *ConnectionController) Friends(
	currentApp *app.App,
	origin uint64,
	userID uint64,
	opts connection.QueryOptions,
) (*ConnectionFeed, error) {
	fs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Before:  opts.Before,
		Enabled: &defaultEnabled,
		FromIDs: []uint64{userID},
		Limit:   opts.Limit,
		States:  []connection.State{connection.StateConfirmed},
		Types:   []connection.Type{connection.TypeFriend},
	})
	if err != nil {
		return nil, err
	}

	ts, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Before:  opts.Before,
		Enabled: &defaultEnabled,
		Limit:   opts.Limit,
		ToIDs:   []uint64{userID},
		States:  []connection.State{connection.StateConfirmed},
		Types:   []connection.Type{connection.TypeFriend},
	})
	if err != nil {
		return nil, err
	}

	cs := append(fs, ts...)

	sort.Sort(cs)

	if len(cs) > opts.Limit {
		cs = cs[:opts.Limit-1]
	}

	ids := []uint64{}

	for _, con := range cs {
		if con.FromID == userID {
			ids = append(ids, con.ToID)
		} else {
			ids = append(ids, con.FromID)
		}
	}

	us, err := user.ListFromIDs(
		c.users,
		currentApp.Namespace(),
		ids...,
	)
	if err != nil {
		return nil, err
	}

	for _, u := range us {
		err := enrichConnectionCounts(c.connections, c.users, currentApp, u)
		if err != nil {
			return nil, err
		}

		err = enrichRelation(c.connections, currentApp, origin, u)
		if err != nil {
			return nil, err
		}
	}

	return &ConnectionFeed{
		Connections: cs,
		Users:       us,
	}, nil
}

// Update transitions the passed Connection to its new state.
func (c *ConnectionController) Update(
	currentApp *app.App,
	new *connection.Connection,
) (*connection.Connection, error) {
	us, err := c.users.Query(currentApp.Namespace(), user.QueryOptions{
		Enabled: &defaultEnabled,
		IDs: []uint64{
			new.ToID,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(us) != 1 {
		return nil, ErrNotFound
	}

	var (
		fromIDs = []uint64{new.FromID}
		toIDs   = []uint64{new.ToID}
	)

	if new.Type == connection.TypeFriend {
		fromIDs = []uint64{new.FromID, new.ToID}
		toIDs = []uint64{new.FromID, new.ToID}
	}

	cs, err := c.connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Enabled: &defaultEnabled,
		FromIDs: fromIDs,
		Limit:   1,
		ToIDs:   toIDs,
		Types:   []connection.Type{new.Type},
	})
	if err != nil {
		return nil, err
	}

	if len(cs) > 0 && cs[0].State == new.State {
		return cs[0], nil
	}

	var old *connection.Connection

	if len(cs) > 0 {
		old = cs[0]

		new.FromID = old.FromID
		new.ToID = old.ToID
	}

	new.Enabled = true

	if err := validateConTransition(old, new); err != nil {
		return nil, err
	}

	return c.connections.Put(currentApp.Namespace(), new)
}

func validateConTransition(old, new *connection.Connection) error {
	if old == nil {
		return nil
	}

	if old.FromID != new.FromID {
		return wrapError(
			ErrInvalidEntity,
			"from id miss-match %d != %d",
			old.FromID,
			new.FromID,
		)
	}

	if old.ToID != new.ToID {
		return wrapError(
			ErrInvalidEntity,
			"to id miss-match %d != %d",
			old.ToID,
			new.ToID,
		)
	}

	if old.Type != new.Type {
		return wrapError(
			ErrInvalidEntity,
			"type miss-match %s != %s",
			string(old.Type),
			string(new.Type),
		)
	}

	if old.State == new.State {
		return nil
	}

	switch old.State {
	case connection.StatePending:
		switch new.State {
		case connection.StateConfirmed, connection.StateRejected:
			return nil
		}
	case connection.StateConfirmed:
		switch new.State {
		case connection.StateRejected:
			return nil
		}
	}

	return wrapError(
		ErrInvalidEntity,
		"invalid state transition from %s to %s",
		string(old.State),
		string(new.State),
	)
}

type relation struct {
	isFriend    bool
	isFollower  bool
	isFollowing bool
	isSelf      bool
}

func queryRelation(
	connections connection.Service,
	currentApp *app.App,
	origin, userID uint64,
) (*relation, error) {
	if origin == userID {
		return &relation{isSelf: true}, nil
	}

	cs, err := connections.Query(currentApp.Namespace(), connection.QueryOptions{
		Enabled: &defaultEnabled,
		FromIDs: []uint64{
			origin,
			userID,
		},
		States: []connection.State{
			connection.StateConfirmed,
		},
		ToIDs: []uint64{
			origin,
			userID,
		},
	})
	if err != nil {
		return nil, err
	}

	r := &relation{}

	for _, c := range cs {
		if c.Type == connection.TypeFriend {
			r.isFriend = true
		}

		if c.Type == connection.TypeFollow && c.FromID == origin {
			r.isFollowing = true
		}

		if c.Type == connection.TypeFollow && c.ToID == origin {
			r.isFollower = true
		}
	}

	return r, nil
}
