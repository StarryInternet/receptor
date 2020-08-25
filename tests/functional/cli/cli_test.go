package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func ConfirmListening(pid int) (bool, error) {
	pidString := "pid=" + strconv.Itoa(pid)
	var out bytes.Buffer
	out = bytes.Buffer{}
	ssCmd := exec.Command("ss", "-tulnp")
	ssCmd.Stdout = &out
	err := ssCmd.Run()
	if err != nil {
		return false, err
	}
	if strings.Contains(out.String(), pidString) {
		return true, nil
	}
	return false, nil
}

func TestHelp(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("receptor", "--help")
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestListeners(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		listener string
	}{
		{"--tcp-listener"},
		{"--ws-listener"},
		{"--udp-listener"},
	}
	for _, data := range testTable {
		listener := data.listener
		t.Run(listener, func(t *testing.T) {
			t.Parallel()
			receptorStdOut := bytes.Buffer{}
			cmd := exec.Command("receptor", "--node", "id=test", listener, "port=0")
			cmd.Stdout = &receptorStdOut
			err := cmd.Start()
			if err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Wait()
			defer cmd.Process.Kill()

			for timeout := 2 * time.Second; timeout > 0; {
				listening, err := ConfirmListening(cmd.Process.Pid)
				if err != nil {
					t.Fatal(err)
				}
				if listening {
					return
				}
				time.Sleep(100 * time.Millisecond)
				timeout -= 100 * time.Millisecond
			}
			t.Fatalf("Timed out while waiting for backend to start:\n%s", receptorStdOut.String())
		})
	}
}

func TestNegativeCost(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		listener string
	}{
		{"--tcp-listener"},
		{"--ws-listener"},
		{"--udp-listener"},
	}
	for _, data := range testTable {
		listener := data.listener
		t.Run(listener, func(t *testing.T) {
			t.Parallel()
			receptorStdOut := bytes.Buffer{}
			cmd := exec.Command("receptor", "--node", "id=test", listener, "port=0", "cost=-1")
			cmd.Stdout = &receptorStdOut
			err := cmd.Start()
			if err != nil {
				t.Fatal(err)
			}

			// Wait for our process to hopefully run and quit
			time.Sleep(100 * time.Millisecond)

			cmd.Process.Kill()
			cmd.Process.Wait()
			if receptorStdOut.String() != "Error: connection cost must be positive\n" {
				t.Fatalf("Expected stdout: Error: connection cost must be positive, actual stdout: %s", receptorStdOut.String())
			}
		})
	}
}

func TestCostMap(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		listener string
		costMaps []string
	}{
		{"--tcp-listener", []string{"{}", "{\"a\": 1}", "{\"a\": 1.1}", "{\"a\": 1.3, \"b\": 5.6, \"c\": 0.2}"}},
		{"--ws-listener", []string{"{}", "{\"a\": 1}", "{\"a\": 1.1}", "{\"a\": 1.3, \"b\": 5.6, \"c\": 0.2}"}},
		{"--udp-listener", []string{"{}", "{\"a\": 1}", "{\"a\": 1.1}", "{\"a\": 1.3, \"b\": 5.6, \"c\": 0.2}"}},
	}
	for _, data := range testTable {
		listener := data.listener
		costMaps := make([]string, len(data.costMaps))
		copy(costMaps, data.costMaps)
		t.Run(listener, func(t *testing.T) {
			t.Parallel()
			for _, costMap := range costMaps {
				costMapCopy := costMap
				t.Run(costMapCopy, func(t *testing.T) {
					t.Parallel()
					receptorStdOut := bytes.Buffer{}
					cmd := exec.Command("receptor", "--node", "id=test", listener, "port=0", fmt.Sprintf("nodecost=%s", costMapCopy))
					cmd.Stdout = &receptorStdOut
					err := cmd.Start()
					if err != nil {
						t.Fatal(err)
					}
					defer cmd.Process.Wait()
					defer cmd.Process.Kill()

					for timeout := 2 * time.Second; timeout > 0; {
						listening, err := ConfirmListening(cmd.Process.Pid)
						if err != nil {
							t.Fatal(err)
						}
						if listening {
							return
						}
						time.Sleep(100 * time.Millisecond)
						timeout -= 100 * time.Millisecond
					}
					t.Fatalf("Timed out while waiting for backend to start:\n%s", receptorStdOut.String())
				})
			}
		})
	}
}

func TestCosts(t *testing.T) {
	t.Parallel()
	testTable := []struct {
		listener string
		costs    []string
	}{
		{"--tcp-listener", []string{"1", "1.5", "1.0", "0.2", "52", "23"}},
		{"--ws-listener", []string{"1", "1.5", "1.0", "0.2", "52", "23"}},
		{"--udp-listener", []string{"1", "1.5", "1.0", "0.2", "52", "23"}},
	}
	for _, data := range testTable {
		listener := data.listener
		costs := make([]string, len(data.costs))
		copy(costs, data.costs)
		t.Run(listener, func(t *testing.T) {
			t.Parallel()
			for _, cost := range costs {
				costCopy := cost
				t.Run(costCopy, func(t *testing.T) {
					t.Parallel()
					receptorStdOut := bytes.Buffer{}
					cmd := exec.Command("receptor", "--node", "id=test", listener, "port=0", fmt.Sprintf("cost=%s", costCopy))
					cmd.Stdout = &receptorStdOut
					err := cmd.Start()
					if err != nil {
						t.Fatal(err)
					}
					defer cmd.Process.Wait()
					defer cmd.Process.Kill()

					for timeout := 2 * time.Second; timeout > 0; {
						listening, err := ConfirmListening(cmd.Process.Pid)
						if err != nil {
							t.Fatal(err)
						}
						if listening {
							return
						}
						time.Sleep(100 * time.Millisecond)
						timeout -= 100 * time.Millisecond
					}
					t.Fatalf("Timed out while waiting for backend to start:\n%s", receptorStdOut.String())
				})
			}
		})
	}
}