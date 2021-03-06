package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/tapglue/multiverse/platform/metrics"
	"github.com/tapglue/multiverse/service/connection"
	"github.com/tapglue/multiverse/service/device"
	"github.com/tapglue/multiverse/service/event"
	"github.com/tapglue/multiverse/service/object"
	"github.com/tapglue/multiverse/service/user"
)

const (
	component        = "sims"
	namespaceService = "service"
	namespaceSource  = "source"
	subsystemErr     = "err"
	subsystemOp      = "op"
	subsystemQueue   = "queue"

	attributeEnabled = "Enabled"
	attributeToken   = "Token"

	fmtURN         = `%s://%s`
	msgAndroid     = `{"GCM": "{\"notification\": {\"title\": \"%s\", \"data\": {\"urn\": \"%s\"}} }"}`
	msgAPNSSandbox = `{"APNS_SANDBOX": "{\"aps\": {\"alert\": \"%s\"}, \"urn\":\"%s\"}" }`
)

// Control flow.
var (
	ErrDeliveryFailure   = errors.New("delivery failed")
	ErrEndpointDisabled  = errors.New("endppint disabled")
	ErrEndpointNotFound  = errors.New("endpoint not found")
	ErrNamespaceNotFound = errors.New("namespace not found")
	ErrPlatformNotFound  = errors.New("platform not found")
)

var (
	defaultDeleted = false
	defaultEnabled = true
	// Set at build time.
	revision = "0000000-dev"
)

// appSchemes maps namespaces to device specific schemes for construction of
// deep links.
var appSchemes = map[string]map[device.Platform]string{
	"app_1_610": map[device.Platform]string{
		device.PlatformIOSSandbox: "simsTest",
		device.PlatformAndroid:    "simsTest",
	},
	"app_1_1147": map[device.Platform]string{
		device.PlatformIOSSandbox: "TapglueSample",
	},
	"app_684_948": map[device.Platform]string{
		device.PlatformIOSSandbox: "uMake",
	},
	"app_684_987": map[device.Platform]string{
		device.PlatformIOSSandbox: "uMake",
	},
	"app_515_922": map[device.Platform]string{
		device.PlatformIOSSandbox: "bkx",
	},
	"app_831_1203": map[device.Platform]string{
		device.PlatformIOSSandbox: "shapeupclub",
	},
}

// arns maps Tapglue app namespaces to SNS ARNs.
var arns = map[string]map[device.Platform]string{
	"app_1_610": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/simsTest",
		device.PlatformAndroid:    "arn:aws:sns:eu-central-1:775034650473:app/GCM/simsTestGCM",
	},
	"app_1_1147": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/tapglue-iOSExample",
	},
	"app_515_922": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/bikestorming-iOSSandbox",
	},
	"app_684_948": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/uMake-iOSSandbox",
	},
	"app_684_987": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/uMake-iOSSandbox",
	},
	"app_831_1203": map[device.Platform]string{
		device.PlatformIOSSandbox: "arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/lifesum-iOSSandbox",
	},
}

// namespaces maps SNS ARNs to Tapglue app namespaces.
var namespaces = map[string]string{
	"arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/simsTest":                "app_1_610",
	"arn:aws:sns:eu-central-1:775034650473:app/GCM/simsTestGCM":                      "app_1_610",
	"arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/uMake-iOSSandbox":        "app_684_948",
	"arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/bikestorming-iOSSandbox": "app_515_922",
	"arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/tapglue-iOSExample":      "app_1_1147",
	"arn:aws:sns:eu-central-1:775034650473:app/APNS_SANDBOX/lifesum-iOSSandbox":      "app_831_1203",
}

type ackFunc func() error
type channelFunc func(string, *message) error
type createEndpointFunc func(platformARN, token string) (string, error)
type disableDeviceFunc func(platformARN, endpointARN string) error
type fetchFollowerIDsFunc func(namespace string, origin uint64) ([]uint64, error)
type fetchFriendIDsFunc func(namespace string, origin uint64) ([]uint64, error)
type fetchObjectFunc func(namespace string, id uint64) (*object.Object, error)
type fetchUserFunc func(namespace string, id uint64) (*user.User, error)
type fetchUsersFunc func(namespace string, ids ...uint64) (user.List, error)
type findDevicesFunc func(namespace string, userID uint64, platforms ...device.Platform) (device.List, error)
type getEndpointFunc func(arn string) (string, error)
type getNamespaceFunc func(arn string) (string, error)
type getPlatformARNFunc func(namespace string, platform device.Platform) (string, error)
type getPlatformNameFunc func(namespace string, platform device.Platform) (string, error)
type getUserDevicesFunc func(namespace string, userID uint64) (device.List, error)
type prepareDeviceEndpointFunc func(namespace string, d *device.Device) (*device.Device, error)
type pushAndroidFunc func(arn, platformName string, message *message) error
type pushAPNSSandboxFunc func(arn, platformName string, message *message) error
type updateTokenFunc func(arn, token string) error

type batch struct {
	ackFunc   ackFunc
	messages  []*message
	namespace string
}

type message struct {
	message   string
	recipient uint64
	urn       string
}

func main() {
	var (
		begin = time.Now()

		awsID         = flag.String("aws.id", "", "Identifier for AWS requests")
		awsRegion     = flag.String("aws.region", "us-east-1", "AWS region to operate in")
		awsSecret     = flag.String("aws.secret", "", "Identification secret for AWS requests")
		postgresURL   = flag.String("postgres.url", "", "Postgres URL to connect to")
		telemetryAddr = flag.String("telemetry.addr", ":9001", "Address to expose telemetry on")
	)
	flag.Parse()

	logger := log.NewContext(
		log.NewJSONLogger(os.Stdout),
	).With(
		"caller", log.Caller(3),
		"component", component,
		"revision", revision,
	)

	hostname, err := os.Hostname()
	if err != nil {
		logger.Log("err", err, "lifecycle", "abort")
	}

	logger = log.NewContext(logger).With("host", hostname)

	// Setup instrumenation
	go func(addr string, begin time.Time, logger log.Logger) {
		http.Handle("/metrics", prometheus.Handler())

		logger = log.NewContext(logger).With(
			"listen", addr,
			"sub", "telemetry",
		)

		_ = logger.Log(
			"duration", time.Now().Sub(begin).Nanoseconds(),
			"lifecycle", "start",
		)

		err := http.ListenAndServe(addr, nil)
		if err != nil {
			logger.Log(
				"err", err,
				"lifecycle", "abort",
			)
		}
	}(*telemetryAddr, begin, logger)

	serviceFieldKeys := []string{
		metrics.FieldComponent,
		metrics.FieldMethod,
		metrics.FieldNamespace,
		metrics.FieldService,
		metrics.FieldStore,
	}

	serviceErrCount := kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: namespaceService,
		Subsystem: subsystemErr,
		Name:      "count",
		Help:      "Number of failed service operations",
	}, serviceFieldKeys)

	serviceOpCount := kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: namespaceService,
		Subsystem: subsystemOp,
		Name:      "count",
		Help:      "Number of service operations performed",
	}, serviceFieldKeys)

	serviceOpLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceService,
			Subsystem: subsystemOp,
			Name:      "latency_seconds",
			Help:      "Distribution of service op duration in seconds",
		},
		serviceFieldKeys,
	)
	prometheus.MustRegister(serviceOpLatency)

	sourceFieldKeys := []string{
		metrics.FieldComponent,
		metrics.FieldMethod,
		metrics.FieldNamespace,
		metrics.FieldSource,
		metrics.FieldStore,
	}

	sourceErrCount := kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: namespaceSource,
		Subsystem: subsystemErr,
		Name:      "count",
		Help:      "Number of failed source operations",
	}, sourceFieldKeys)

	sourceOpCount := kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: namespaceSource,
		Subsystem: subsystemOp,
		Name:      "count",
		Help:      "Number of source operations performed",
	}, sourceFieldKeys)

	sourceOpLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceSource,
			Subsystem: subsystemOp,
			Name:      "latency_seconds",
			Help:      "Distribution of source op duration in seconds",
			Buckets:   metrics.BucketsQueue,
		},
		sourceFieldKeys,
	)
	prometheus.MustRegister(sourceOpLatency)

	sourceQueueLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceSource,
			Subsystem: subsystemQueue,
			Name:      "latency_seconds",
			Help:      "Distribution of message queue latency in seconds",
			Buckets:   metrics.BucketsQueue,
		},
		sourceFieldKeys,
	)
	prometheus.MustRegister(sourceQueueLatency)

	aSession := awsSession.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(*awsID, *awsSecret, ""),
		Region:      aws.String(*awsRegion),
	})

	db, err := sqlx.Connect("postgres", *postgresURL)
	if err != nil {
		logger.Log("err", err, "lifecycle", "abort")
		os.Exit(1)
	}

	var connections connection.Service
	connections = connection.NewPostgresService(db)
	connections = connection.InstrumentServiceMiddleware(component, "postgres", serviceErrCount, serviceOpCount, serviceOpLatency)(connections)
	connections = connection.LogServiceMiddleware(logger, "postgres")(connections)

	var devices device.Service
	devices = device.PostgresService(db)
	devices = device.InstrumentServiceMiddleware(component, "postgres", serviceErrCount, serviceOpCount, serviceOpLatency)(devices)
	devices = device.LogServiceMiddleware(logger, "postgres")(devices)

	var objects object.Service
	objects = object.NewPostgresService(db)
	objects = object.InstrumentServiceMiddleware(component, "postgres", serviceErrCount, serviceOpCount, serviceOpLatency)(objects)
	objects = object.LogServiceMiddleware(logger, "postgres")(objects)

	var users user.Service
	users = user.NewPostgresService(db)
	users = user.InstrumentMiddleware(component, "postgres", serviceErrCount, serviceOpCount, serviceOpLatency)(users)
	users = user.LogMiddleware(logger, "postgres")(users)

	snsService := sns.New(aSession)

	conSource, err := connection.SQSSource(sqs.New(aSession))
	if err != nil {
		logger.Log("err", err, "lifecycle", "abort")
		os.Exit(1)
	}

	conSource = connection.InstrumentSourceMiddleware(
		component,
		"sqs",
		sourceErrCount,
		sourceOpCount,
		sourceOpLatency,
		sourceQueueLatency,
	)(conSource)
	conSource = connection.LogSourceMiddleware("sqs", logger)(conSource)

	eventSource, err := event.SQSSource(sqs.New(aSession))
	if err != nil {
		logger.Log("err", err, "lifecycle", "abort")
		os.Exit(1)
	}

	eventSource = event.InstrumentSourceMiddleware(
		component,
		"sqs",
		sourceErrCount,
		sourceOpCount,
		sourceOpLatency,
		sourceQueueLatency,
	)(eventSource)
	eventSource = event.LogSourceMiddleware("sqs", logger)(eventSource)

	objectSource, err := object.SQSSource(sqs.New(aSession))
	if err != nil {
		logger.Log("err", err, "lifecycle", "abort")
		os.Exit(1)
	}

	objectSource = object.InstrumentSourceMiddleware(
		component,
		"sqs",
		sourceErrCount,
		sourceOpCount,
		sourceOpLatency,
		sourceQueueLatency,
	)(objectSource)
	objectSource = object.LogSourceMiddleware("sqs", logger)(objectSource)

	var createEndpoint createEndpointFunc
	var disableDevice disableDeviceFunc
	var fetchFollowerIDs fetchFollowerIDsFunc
	var fetchFriendIDs fetchFriendIDsFunc
	var fetchObject fetchObjectFunc
	var fetchUser fetchUserFunc
	var fetchUsers fetchUsersFunc
	var getUserDevices getUserDevicesFunc
	var getEndpoint getEndpointFunc
	var getNamespace getNamespaceFunc
	var getPlatformARN getPlatformARNFunc
	var getPlatformName getPlatformNameFunc
	var prepareDeviceEndpoint prepareDeviceEndpointFunc
	var pushAndroid pushAndroidFunc
	var pushAPNSSandbox pushAPNSSandboxFunc
	var updateToken updateTokenFunc

	createEndpoint = func(pARN, token string) (string, error) {
		r, err := snsService.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
			PlatformApplicationArn: aws.String(pARN),
			Token: aws.String(token),
		})
		if err != nil {
			return "", err
		}

		return *r.EndpointArn, nil
	}

	disableDevice = func(pARN, eARN string) error {
		ns, err := getNamespace(pARN)
		if err != nil {
			return err
		}

		ds, err := devices.Query(ns, device.QueryOptions{
			Deleted: &defaultDeleted,
			EndpointARNs: []string{
				eARN,
			},
		})
		if err != nil {
			return err
		}

		if len(ds) == 0 {
			return nil
		}

		d := ds[0]
		d.Disabled = true

		_, err = devices.Put(ns, d)

		return err
	}

	fetchFollowerIDs = func(ns string, origin uint64) ([]uint64, error) {
		fs, err := connections.Query(ns, connection.QueryOptions{
			Enabled: &defaultEnabled,
			States: []connection.State{
				connection.StateConfirmed,
			},
			ToIDs: []uint64{
				origin,
			},
			Types: []connection.Type{
				connection.TypeFollow,
			},
		})
		if err != nil {
			return nil, err
		}

		return fs.FromIDs(), nil
	}

	fetchFriendIDs = func(ns string, origin uint64) ([]uint64, error) {
		fs, err := connections.Query(ns, connection.QueryOptions{
			Enabled: &defaultEnabled,
			FromIDs: []uint64{
				origin,
			},
			States: []connection.State{
				connection.StateConfirmed,
			},
			Types: []connection.Type{
				connection.TypeFriend,
			},
		})
		if err != nil {
			return nil, err
		}

		ts, err := connections.Query(ns, connection.QueryOptions{
			Enabled: &defaultEnabled,
			ToIDs: []uint64{
				origin,
			},
			States: []connection.State{
				connection.StateConfirmed,
			},
			Types: []connection.Type{
				connection.TypeFriend,
			},
		})
		if err != nil {
			return nil, err
		}

		return append(fs.ToIDs(), ts.FromIDs()...), nil
	}

	fetchObject = func(ns string, id uint64) (*object.Object, error) {
		os, err := objects.Query(ns, object.QueryOptions{
			ID: &id,
		})
		if err != nil {
			return nil, err
		}

		if len(os) != 1 {
			return nil, object.ErrNotFound
		}

		return os[0], nil
	}

	fetchUser = func(ns string, id uint64) (*user.User, error) {
		us, err := users.Query(ns, user.QueryOptions{
			Enabled: &defaultEnabled,
			IDs: []uint64{
				id,
			},
		})
		if err != nil {
			return nil, err
		}

		if len(us) == 0 {
			return nil, fmt.Errorf("user '%d' not found", id)
		}

		return us[0], nil
	}

	fetchUsers = func(ns string, ids ...uint64) (user.List, error) {
		if len(ids) == 0 {
			return user.List{}, nil
		}

		return users.Query(ns, user.QueryOptions{
			Enabled: &defaultEnabled,
			IDs:     ids,
		})
	}

	getUserDevices = func(ns string, userID uint64) (device.List, error) {
		ds, err := devices.Query(ns, device.QueryOptions{
			Deleted:  &defaultDeleted,
			Disabled: &defaultDeleted,
			Platforms: []device.Platform{
				device.PlatformIOSSandbox,
				device.PlatformIOS,
				device.PlatformAndroid,
			},
			UserIDs: []uint64{
				userID,
			},
		})
		if err != nil {
			return nil, err
		}

		es := device.List{}

		for _, d := range ds {
			_, err := prepareDeviceEndpoint(ns, d)
			if isEndpointDisabled(err) || isNamespaceNotFound(err) || isPlatformNotFound(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			es = append(es, d)
		}

		return es, nil
	}

	getEndpoint = func(arn string) (string, error) {
		r, err := snsService.GetEndpointAttributes(&sns.GetEndpointAttributesInput{
			EndpointArn: aws.String(arn),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok && awsErr.StatusCode() == 404 {
				return "", ErrEndpointNotFound
			}

			return "", err
		}

		enabled := *r.Attributes[attributeEnabled]

		if enabled == "false" {
			return "", ErrEndpointDisabled
		}

		return *r.Attributes[attributeToken], nil
	}

	getNamespace = func(arn string) (string, error) {
		ns, ok := namespaces[arn]
		if !ok {
			return "", ErrNamespaceNotFound
		}

		return ns, nil
	}

	getPlatformARN = func(ns string, platform device.Platform) (string, error) {
		cs, ok := arns[ns]
		if !ok {
			return "", ErrNamespaceNotFound
		}

		arn, ok := cs[platform]
		if !ok {
			return "", ErrPlatformNotFound
		}

		return arn, nil
	}

	getPlatformName = func(ns string, platform device.Platform) (string, error) {
		cs, ok := appSchemes[ns]
		if !ok {
			return "", ErrNamespaceNotFound
		}

		name, ok := cs[platform]
		if !ok {
			return "", ErrPlatformNotFound
		}

		return name, nil
	}

	prepareDeviceEndpoint = func(ns string, d *device.Device) (*device.Device, error) {
		pARN, err := getPlatformARN(ns, d.Platform)
		if err != nil {
			return nil, err
		}

		if d.EndpointARN == "" {
			arn, err := createEndpoint(pARN, d.Token)
			if err != nil {
				return nil, err
			}

			d.EndpointARN = arn

			d, err = devices.Put(ns, d)
			if err != nil {
				return nil, err
			}

			return d, nil
		}

		token, err := getEndpoint(d.EndpointARN)
		if !isEndpointNotFound(err) {
			return nil, err
		}

		if isEndpointNotFound(err) {
			arn, err := createEndpoint(pARN, d.Token)
			if err != nil {
				return nil, err
			}

			d.EndpointARN = arn

			d, err = devices.Put(ns, d)
			if err != nil {
				return nil, err
			}

			token = d.Token
		}

		if token != d.Token {
			err := updateToken(d.EndpointARN, d.Token)
			if err != nil {
				return nil, err
			}
		}

		return d, nil
	}

	pushAndroid = func(arn, platformName string, msg *message) error {
		var (
			urn = fmt.Sprintf(fmtURN, platformName, msg.urn)
			m   = fmt.Sprintf(msgAndroid, msg.message, urn)
		)

		_, err := snsService.Publish(&sns.PublishInput{
			Message:          aws.String(m),
			MessageStructure: aws.String("json"),
			TargetArn:        aws.String(arn),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				if awsErr.StatusCode() == 400 {
					return ErrDeliveryFailure
				}
			}
		}

		return nil
	}

	pushAPNSSandbox = func(arn, platformName string, msg *message) error {
		var (
			urn = fmt.Sprintf(fmtURN, platformName, msg.urn)
			m   = fmt.Sprintf(msgAPNSSandbox, msg.message, urn)
		)

		_, err = snsService.Publish(&sns.PublishInput{
			Message:          aws.String(m),
			MessageStructure: aws.String("json"),
			TargetArn:        aws.String(arn),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.RequestFailure); ok {
				if awsErr.StatusCode() == 400 {
					return ErrDeliveryFailure
				}
			}
		}
		return nil
	}

	updateToken = func(arn, token string) error {
		_, err := snsService.SetEndpointAttributes(&sns.SetEndpointAttributesInput{
			Attributes: map[string]*string{
				attributeToken: aws.String(token),
			},
			EndpointArn: aws.String(arn),
		})
		return err
	}

	logger.Log(
		"duration", time.Now().Sub(begin).Nanoseconds(),
		"lifecycle", "start",
		"sub", "worker",
	)

	changec := make(chan endpointChange)

	go func() {
		err := consumeEndpointChange(
			sqs.New(aSession),
			"https://sqs.eu-central-1.amazonaws.com/775034650473/endpoint-changes",
			changec,
		)
		if err != nil {
			logger.Log("err", err, "lifecycle", "abort")
			os.Exit(1)
		}
	}()

	go func() {
		for c := range changec {
			err := updateEndpoint(c, disableDevice)
			if err != nil {
				_ = logger.Log("err", err, "lifecycle", "abort")
				os.Exit(1)
			}
		}
	}()

	batchc := make(chan batch)

	go func() {
		err := consumeConnection(
			conSource,
			batchc,
			conRuleFollower(fetchUser),
			conRuleFriendConfirmed(fetchUser),
			conRuleFriendRequest(fetchUser),
		)
		if err != nil {
			logger.Log("err", err, "lifecycle", "abort")
			os.Exit(1)
		}
	}()

	go func() {
		err := consumeEvent(
			eventSource,
			batchc,
			eventRuleLikeCreated(fetchFollowerIDs, fetchFriendIDs, fetchObject, fetchUser, fetchUsers),
		)
		if err != nil {
			logger.Log("err", err, "lifecycle", "abort")
			os.Exit(1)
		}
	}()

	go func() {
		err := consumeObject(
			objectSource,
			batchc,
			objectRuleCommentCreated(fetchFollowerIDs, fetchFriendIDs, fetchObject, fetchUser, fetchUsers),
			objectRulePostCreated(fetchFollowerIDs, fetchFriendIDs, fetchUser, fetchUsers),
		)
		if err != nil {
			logger.Log("err", err, "lifecycle", "abort")
			os.Exit(1)
		}
	}()

	cs := []channelFunc{
		channelPush(getUserDevices, getPlatformName, pushAndroid, pushAPNSSandbox),
	}

	for batch := range batchc {
		for _, msg := range batch.messages {
			for _, channel := range cs {
				err := channel(batch.namespace, msg)
				if err != nil {
					logger.Log("err", err, "lifecycle", "abort")
					os.Exit(1)
				}
			}
		}

		err = batch.ackFunc()
		if err != nil {
			logger.Log("err", err, "lifecycle", "abort")
			os.Exit(1)
		}
	}

	logger.Log("lifecycle", "stop")
}

func channelPush(
	getUserDevices getUserDevicesFunc,
	getPlatformName getPlatformNameFunc,
	pushAndroid pushAndroidFunc,
	pushAPNSSandbox pushAPNSSandboxFunc,
) channelFunc {
	return func(ns string, msg *message) error {
		// find devices
		ds, err := getUserDevices(ns, msg.recipient)
		if err != nil {
			return err
		}
		if len(ds) == 0 {
			return nil
		}

		// publish to devices
		for _, d := range ds {
			switch d.Platform {
			case device.PlatformIOSSandbox:
				name, err := getPlatformName(ns, d.Platform)
				if err != nil {
					return err
				}

				err = pushAPNSSandbox(d.EndpointARN, name, msg)
				if err != nil {
					if isDeliveryFailure(err) {
						return nil
					}

					return err
				}
			case device.PlatformAndroid:
				name, err := getPlatformName(ns, d.Platform)
				if err != nil {
					return err
				}

				err = pushAndroid(d.EndpointARN, name, msg)
				if err != nil {
					if isDeliveryFailure(err) {
						return nil
					}

					return err
				}
			default:
				return fmt.Errorf("platform not supported")
			}
		}

		return nil
	}
}

func isDeliveryFailure(err error) bool {
	return err == ErrDeliveryFailure
}

func isEndpointDisabled(err error) bool {
	return err == ErrEndpointDisabled
}

func isEndpointNotFound(err error) bool {
	return err == ErrEndpointNotFound
}

func isNamespaceNotFound(err error) bool {
	return err == ErrNamespaceNotFound
}

func isPlatformNotFound(err error) bool {
	return err == ErrPlatformNotFound
}
