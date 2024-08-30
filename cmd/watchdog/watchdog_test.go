package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"hw1/cmd"
	"hw1/models"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchdog(t *testing.T) {
	ctx := context.Background()

	err := os.Setenv("TEST_ENV", "TEST_ENV_DATA")
	require.NoError(t, err)

	var success bool
	var count atomic.Int64

	var handler http.HandlerFunc = func(responseWriter http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)

		defer func() {
			err = request.Body.Close()
			require.NoError(t, err)
		}()

		if string(body) == "COUNTER" {
			count.Add(1)
			return
		}

		require.Equal(t, body, []byte("OK"))
		success = true
	}

	testServer := httptest.NewServer(handler)
	t.Cleanup(func() {
		testServer.Close()
	})

	path, err := cmd.ResolvePath("watchdog_integration_test_server.go")
	require.NoError(t, err)

	binPath, err := cmd.ResolvePath("bin")
	require.NoError(t, err)

	// do not check errors, fs specific, there may be no permissions
	os.MkdirAll(binPath, os.ModePerm)
	os.Chmod(binPath, os.ModePerm)

	var outputPath string

	if runtime.GOOS == "windows" {
		outputPath = filepath.Join(binPath, "watchdog_integration_test_server.exe")
	} else {
		outputPath = filepath.Join(binPath, "watchdog_integration_test_server")
	}

	err = cmd.GoBuild(ctx, path, outputPath)
	require.NoError(t, err)

	os.Args = append(
		os.Args,
		"-"+models.ServerBinaryPathArgName, outputPath,
		"-"+models.ServerProgramArgumentsArgName, "-test_server_result_url "+testServer.URL,
	)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdin = r

	_, err = w.Write([]byte(models.WatchDogCommandSetMacProc + " 32 "))
	require.NoError(t, err)

	_, err = w.Write([]byte(models.WatchDogCommandExit + " "))
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()

	wg.Wait()

	require.True(t, success)

	for {
		prev := count.Load()
		time.Sleep(time.Second)

		cur := count.Load()

		if prev == cur {
			break
		}

		fmt.Println("Wait for killing")
	}
}