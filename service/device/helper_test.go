package device

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/tapglue/multiverse/platform/generate"
)

type prepareFunc func(t *testing.T, namespace string) Service

func testServicePut(t *testing.T, p prepareFunc) {
	var (
		device    = testDevice()
		namespace = "service_put"
		service   = p(t, namespace)
	)

	created, err := service.Put(namespace, device)
	if err != nil {
		t.Fatal(err)
	}

	list, err := service.Query(namespace, QueryOptions{
		IDs: []uint64{
			created.ID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(list), 1; have != want {
		t.Fatalf("have %v, want %v", have, want)
	}
	if have, want := list[0], created; !reflect.DeepEqual(have, want) {
		t.Errorf("have %v, want %v", have, want)
	}

	list[0].Token = generate.RandomString(18)

	updated, err := service.Put(namespace, list[0])
	if err != nil {
		t.Fatal(err)
	}

	list, err = service.Query(namespace, QueryOptions{
		IDs: []uint64{
			updated.ID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := list[0], updated; !reflect.DeepEqual(have, want) {
		t.Errorf("have %v, want %v", have, want)
	}
}

func testServiceQuery(t *testing.T, p prepareFunc) {
	var (
		deleted   = false
		namespace = "service_query"
		service   = p(t, namespace)
	)

	ds, err := service.Query(namespace, QueryOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 0; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	created, err := service.Put(namespace, testDevice())
	if err != nil {
		t.Fatal(err)
	}

	// QueryOptions{
	// 	Deleted: false,
	// DevicIDs
	// 	IDs: []uint64{},
	// 	UserIDs: []uint64{},
	// }

	for _, d := range testList() {
		_, err := service.Put(namespace, d)
		if err != nil {
			t.Fatal(err)
		}
	}

	ds, err = service.Query(namespace, QueryOptions{
		Deleted: &deleted,
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 8; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	ds, err = service.Query(namespace, QueryOptions{
		DeviceIDs: []string{
			created.DeviceID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 1; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	ds, err = service.Query(namespace, QueryOptions{
		IDs: []uint64{
			created.ID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 1; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	ds, err = service.Query(namespace, QueryOptions{
		Platforms: []Platform{
			PlatformIOSSandbox,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 13; have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	ds, err = service.Query(namespace, QueryOptions{
		UserIDs: []uint64{
			created.UserID,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(ds), 1; have != want {
		t.Errorf("have %v, want %v", have, want)
	}
}

func testDevice() *Device {
	return &Device{
		Deleted:     false,
		DeviceID:    generate.RandomString(24),
		EndpointARN: generate.RandomString(32),
		Language:    DefaultLanguage,
		Platform:    PlatformIOSSandbox,
		Token:       generate.RandomString(18),
		UserID:      uint64(rand.Int63()),
	}
}

func testList() List {
	ds := List{}

	for i := 0; i < 7; i++ {
		ds = append(ds, testDevice())
	}

	for i := 0; i < 5; i++ {
		d := testDevice()
		d.Deleted = true

		ds = append(ds, d)
	}

	return ds
}
