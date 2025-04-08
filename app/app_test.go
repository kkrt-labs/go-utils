package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/kkrt-labs/go-utils/tag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestApp(t *testing.T) *App {
	cfg := new(Config)
	cfg.MainEntrypoint.Net.KeepAlive = "1s"
	cfg.MainEntrypoint.HTTP.ReadTimeout = "1s"
	cfg.MainEntrypoint.HTTP.ReadHeaderTimeout = "1s"
	cfg.MainEntrypoint.HTTP.WriteTimeout = "1s"
	cfg.MainEntrypoint.HTTP.IdleTimeout = "1s"
	cfg.HealthzEntrypoint.Net.KeepAlive = "1s"
	cfg.HealthzEntrypoint.HTTP.ReadTimeout = "1s"
	cfg.HealthzEntrypoint.HTTP.ReadHeaderTimeout = "1s"
	cfg.HealthzEntrypoint.HTTP.WriteTimeout = "1s"
	cfg.HealthzEntrypoint.HTTP.IdleTimeout = "1s"

	app, err := NewApp(
		cfg,
		WithLogger(zap.NewNop()),
		WithName("test"),
		WithVersion("1.0.0"),
	)
	require.NoError(t, err)
	return app
}

func TestAppProvide(t *testing.T) {
	var testCase = []struct {
		desc        string
		constructor func() (any, error)
		expected    any
		expectErr   bool
	}{
		{
			desc: "string",
			constructor: func() (any, error) {
				return "test", nil
			},
			expected: "test",
		},
		{
			desc: "int",
			constructor: func() (any, error) {
				return 1, nil
			},
			expected: 1,
		},
		{
			desc: "nil",
			constructor: func() (any, error) {
				return nil, nil
			},
			expected: nil,
		},
		{
			desc: "error",
			constructor: func() (any, error) {
				return nil, errors.New("error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.desc, func(t *testing.T) {
			app := newTestApp(t)
			res := app.Provide("test", tc.constructor)
			err := app.Error()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}

func TestProvide(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (string, error) {
			return "test", nil
		})
		assert.Equal(t, res, "test")
	})

	t.Run("int", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (int, error) {
			return 1, nil
		})
		assert.Equal(t, res, 1)
	})

	t.Run("*string", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (*string, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})

	t.Run("*string#nil", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (*string, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})

	t.Run("error", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (error, error) {
			return errors.New("error"), nil
		})
		assert.Equal(t, errors.New("error"), res)
	})

	t.Run("interface", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (interface{}, error) {
			return "test", nil
		})
		assert.Equal(t, res, "test")
	})

	t.Run("interface#nil", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (interface{}, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})
}

type testService struct {
	started bool
	stopped bool

	start chan error
	stop  chan error
}

func (s *testService) Start(_ context.Context) error {
	if s.started {
		return errors.New("already started")
	}
	s.started = true
	return <-s.start
}

func (s *testService) Stop(_ context.Context) error {
	if !s.started {
		return errors.New("not started")
	}
	if s.stopped {
		return errors.New("already stopped")
	}
	s.stopped = true
	return <-s.stop
}

func TestAppNoDeps(t *testing.T) {
	start, stop := make(chan error), make(chan error)
	defer close(start)
	defer close(stop)

	testApp := func() *App {
		app := newTestApp(t)
		_ = Provide(app, "test", func() (*testService, error) {
			return &testService{
				start: start,
				stop:  stop,
			}, nil
		})
		return app
	}

	resStart, resStop := make(chan error), make(chan error)
	defer close(resStart)
	defer close(resStop)

	t.Run("no errors", func(t *testing.T) {
		app := testApp()
		require.Equal(t, app.services["test"].Status(), Constructed)

		go func() {
			resStart <- app.Start(context.Background())
		}()
		time.Sleep(100 * time.Millisecond) // wait for the service to start
		assert.Equal(t, app.services["test"].Status(), Starting)

		// Trigger start
		start <- nil
		assert.Nil(t, <-resStart)
		assert.Equal(t, app.services["test"].Status(), Running)

		go func() {
			resStop <- app.Stop(context.Background())
		}()
		time.Sleep(100 * time.Millisecond) // wait for the service to start
		assert.Equal(t, app.services["test"].Status(), Stopping)

		// Trigger stop
		stop <- nil
		assert.Nil(t, <-resStop)
		assert.Equal(t, app.services["test"].Status(), Stopped)
	})

	t.Run("error on start", func(t *testing.T) {
		app := testApp()
		go func() {
			resStart <- app.Start(context.Background())
		}()

		start <- errors.New("error on start")
		startErr := <-resStart
		require.IsType(t, startErr, &ServiceError{})
		assert.Equal(t, startErr.(*ServiceError).directErr, errors.New("error on start"))
		assert.Equal(t, app.services["test"].Status(), Error)
	})

	t.Run("error on stop", func(t *testing.T) {
		app := testApp()
		go func() {
			resStart <- app.Start(context.Background())
		}()
		start <- nil
		<-resStart

		go func() {
			resStop <- app.Stop(context.Background())
		}()
		stop <- errors.New("error on stop")
		stopErr := <-resStop
		require.IsType(t, stopErr, &ServiceError{})
		assert.Equal(t, stopErr.(*ServiceError).directErr, errors.New("error on stop"))
		assert.Equal(t, app.services["test"].Status(), Error)
	})
}

func TestAppWithDeps(t *testing.T) {
	app := newTestApp(t)
	startMain, stopMain, startDep, stopDep := make(chan error), make(chan error), make(chan error), make(chan error)
	defer close(startMain)
	defer close(stopMain)
	defer close(startDep)
	defer close(stopDep)

	_ = Provide(app, "main", func() (*testService, error) {
		_ = Provide(app, "dep", func() (*testService, error) {
			return &testService{
				start: startDep,
				stop:  stopDep,
			}, nil
		})
		return &testService{
			start: startMain,
			stop:  stopMain,
		}, nil
	})

	// Test dependency tree
	assert.Equal(t, app.services["main"].deps["dep"], app.services["dep"])
	assert.Equal(t, app.services["dep"].depsOf["main"], app.services["main"])

	resStart, resStop := make(chan error), make(chan error)
	defer close(resStart)
	defer close(resStop)

	go func() {
		resStart <- app.Start(context.Background())
	}()
	startDep <- nil
	startMain <- nil
	assert.Nil(t, <-resStart)

	go func() {
		resStop <- app.Stop(context.Background())
	}()
	stopMain <- nil
	stopDep <- nil
	assert.Nil(t, <-resStop)
}

func TestAppWithMultipleTopConstruct(t *testing.T) {
	app := newTestApp(t)
	startTop1, stopTop1, startTop2, stopTop2, startDep, stopDep := make(chan error), make(chan error), make(chan error), make(chan error), make(chan error), make(chan error)
	defer close(startTop1)
	defer close(stopTop1)
	defer close(startTop2)
	defer close(stopTop2)
	defer close(startDep)
	defer close(stopDep)

	provideDep := func() *testService {
		return Provide(app, "dep", func() (*testService, error) {
			return &testService{
				start: startDep,
				stop:  stopDep,
			}, nil
		})
	}

	// Provide top1
	dep := provideDep()

	top1 := Provide(app, "top1", func() (*testService, error) {
		provideDep()
		return &testService{
			start: startTop1,
			stop:  stopTop1,
		}, nil
	})

	// Provide top2
	top2 := Provide(app, "top2", func() (*testService, error) {
		provideDep()
		return &testService{
			start: startTop2,
			stop:  stopTop2,
		}, nil
	})

	// Test dependency tree
	assert.Equal(t, app.services["top1"].deps["dep"], app.services["dep"])
	assert.Equal(t, app.services["top2"].deps["dep"], app.services["dep"])
	assert.Equal(t, app.services["dep"].depsOf["top1"], app.services["top1"])
	assert.Equal(t, app.services["dep"].depsOf["top2"], app.services["top2"])

	resStart := make(chan error)
	defer close(resStart)
	go func() {
		resStart <- app.Start(context.Background())
	}()

	// Sleeps to ensure that deps as started at this point tops should be waiting for the dep to start
	time.Sleep(100 * time.Millisecond)
	assert.True(t, dep.started, "#1 dep should be started")
	assert.False(t, top1.started, "#1 top1 should not be started")
	assert.False(t, top2.started, "#1 top2 should not be started")

	// Succeed dep start which should trigger top starts
	startDep <- nil
	time.Sleep(100 * time.Millisecond)
	assert.True(t, top1.started, "#2 top1 should be started")
	assert.True(t, top2.started, "#2 top2 should be started")

	// Succeed tops start
	startTop1 <- nil
	startTop2 <- nil
	assert.Nil(t, <-resStart)

	//
	resStop := make(chan error)
	defer close(resStop)
	go func() {
		resStop <- app.Stop(context.Background())
	}()

	// Sleep ensure that service are stopping, at this point top1 and top2 should be stopped, but dep should not
	time.Sleep(100 * time.Millisecond)
	assert.True(t, top1.stopped, "#3 top1 should be stopped")
	assert.True(t, top2.stopped, "#3 top2 should be stopped")
	assert.False(t, dep.stopped, "#3 dep should not be stopped")

	// Succeeds top 1 stop, then dep should still not be stopped
	stopTop1 <- nil
	time.Sleep(100 * time.Millisecond)
	assert.False(t, dep.stopped, "#4 dep should not be stopped")

	// Succeeds top 2 stop, then dep should now be stopped
	stopTop2 <- nil
	time.Sleep(100 * time.Millisecond)
	assert.True(t, dep.stopped, "#5 dep should be stopped")

	// Succeeds dep stop, then app.Stop should return
	stopDep <- nil
	assert.Nil(t, <-resStop)
}

func TestServiceError(t *testing.T) {
	app := newTestApp(t)

	Provide(app, "svc", func() (any, error) {
		_ = Provide(app, "dep1", func() (any, error) {
			_ = Provide(app, "dep11", func() (any, error) { return nil, fmt.Errorf("error on dep11") })
			_ = Provide(app, "dep12", func() (any, error) { return nil, nil })
			return nil, nil
		})
		_ = Provide(app, "dep2", func() (any, error) {
			_ = Provide(app, "dep21", func() (any, error) { return nil, nil })
			_ = Provide(app, "dep22", func() (any, error) { return nil, nil })
			return nil, fmt.Errorf("error on dep2")
		})
		return nil, fmt.Errorf("error on svc")
	})

	expected := `service "svc": error on svc
>service "dep1"
>>service "dep11": error on dep11
>service "dep2": error on dep2`
	assert.Equal(t, expected, app.Error().Error())
}

func TestAppServers(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.Error())

	app.Provide("top", func() (any, error) {
		app.EnableMainEntrypoint()
		app.EnableHealthzEntrypoint()
		return nil, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	// Check main server is running
	require.NotNil(t, app.main)
	mainAddr := app.main.Addr()
	require.NotEmpty(t, mainAddr)

	conn, err := net.Dial("tcp", mainAddr)
	require.NoError(t, err)
	conn.Close()

	// Check main server is running
	require.NotNil(t, app.healthz)
	healthzAddr := app.healthz.Addr()
	require.NotEmpty(t, healthzAddr)

	conn, err = net.Dial("tcp", healthzAddr)
	require.NoError(t, err)
	conn.Close()

	// Check healthz server is running
	err = app.Stop(context.Background())
	require.NoError(t, err)
}

type checkableService struct {
	err error
}

func (s *checkableService) Ready(_ context.Context) error {
	return s.err
}

func TestHealthChecks(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.Error())

	checkable := new(checkableService)
	Provide(app, "checkable", func() (*checkableService, error) {
		app.EnableHealthzEntrypoint()
		return checkable, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	require.NotNil(t, app.healthz)
	healthAddr := app.healthz.Addr()
	require.NotEmpty(t, healthAddr)

	// Test live check
	req, err := http.NewRequest("GET", "http://"+healthAddr+"/live", http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test ready check
	req, err = http.NewRequest("GET", "http://"+healthAddr+"/ready", http.NoBody)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test ready check with error
	checkable.err = errors.New("test error")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusServiceUnavailable)
}

type metricsService struct {
	name  string
	count prometheus.Counter
}

func (s *metricsService) SetMetrics(appName, subsystem string, _ ...*tag.Tag) {
	s.count = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: appName,
		Subsystem: subsystem,
		Name:      fmt.Sprintf("%s_count", s.name),
	})
}

func (s *metricsService) incr() {
	s.count.Inc()
}

func (s *metricsService) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.count.Desc()
}

func (s *metricsService) Collect(ch chan<- prometheus.Metric) {
	ch <- s.count
}

func TestMetrics(t *testing.T) {
	app := newTestApp(t)
	_ = WithName("testApp")(app)

	require.NoError(t, app.Error())

	metrics := &metricsService{
		name: "A",
	}
	app.Provide("test-wo-cfg", func() (any, error) {
		app.EnableHealthzEntrypoint()
		app.Provide(
			"test-w-cfg",
			func() (any, error) {
				return &metricsService{
					name: "B",
				}, nil
			},
			WithComponentName("subsystem"),
		)
		return metrics, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	require.NotNil(t, app.healthz)
	healthAddr := app.healthz.Addr()
	require.NotEmpty(t, healthAddr)

	// Test metrics endpoint
	req, err := http.NewRequest("GET", "http://"+healthAddr+"/metrics", http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test collectors are registered with correct labels
	families, err := app.prometheus.Gather()
	require.NoError(t, err)
	familyCount := len(families)
	assert.GreaterOrEqual(t, familyCount, 2)
	assert.Equal(t, "testApp_subsystem_B_count", families[familyCount-2].GetName())
	assert.Equal(t, "testApp_test_wo_cfg_A_count", families[familyCount-1].GetName())

	// Test metrics are updated
	assert.Equal(t, float64(0), families[familyCount-1].GetMetric()[0].GetCounter().GetValue())
	metrics.incr()
	metrics.incr()
	metrics.incr()

	families, err = app.prometheus.Gather()
	require.NoError(t, err)
	assert.Equal(t, float64(3), families[familyCount-1].GetMetric()[0].GetCounter().GetValue())

	err = app.Stop(context.Background())
	require.NoError(t, err)
}
