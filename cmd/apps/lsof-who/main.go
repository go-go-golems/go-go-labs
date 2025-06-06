package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/charmbracelet/huh"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type ProcessInfo struct {
	PID        int32             `json:"pid"`
	Name       string            `json:"name"`
	Cmdline    string            `json:"cmdline,omitempty"`
	Exe        string            `json:"exe,omitempty"`
	Username   string            `json:"username,omitempty"`
	ParentPID  int32             `json:"parent_pid,omitempty"`
	CPUPercent float64           `json:"cpu_percent,omitempty"`
	RSS        uint64            `json:"rss_bytes,omitempty"`
	StartTime  int64             `json:"create_time,omitempty"`
	OpenFiles  int               `json:"open_files,omitempty"`
	ParentTree []ProcessAncestor `json:"parent_tree,omitempty"`
}

type ProcessAncestor struct {
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	Cmdline  string `json:"cmdline,omitempty"`
	Username string `json:"username,omitempty"`
}

func findProcessByPort(port uint32) (*process.Process, error) {
	conns, err := net.Connections("inet")
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		if c.Status == "LISTEN" && c.Laddr.Port == port {
			return process.NewProcess(c.Pid)
		}
	}
	return nil, fmt.Errorf("no process listening on port %d", port)
}

func buildParentTree(p *process.Process) ([]ProcessAncestor, error) {
	var ancestors []ProcessAncestor
	current := p

	for {
		parentPID, err := current.Ppid()
		if err != nil || parentPID <= 1 {
			break
		}

		parent, err := process.NewProcess(parentPID)
		if err != nil {
			break
		}

		name, _ := parent.Name()
		cmdline, _ := parent.Cmdline()
		username, _ := parent.Username()

		ancestor := ProcessAncestor{
			PID:      parentPID,
			Name:     name,
			Cmdline:  cmdline,
			Username: username,
		}

		ancestors = append(ancestors, ancestor)
		current = parent

		// Prevent infinite loops
		if len(ancestors) > 20 {
			break
		}
	}

	return ancestors, nil
}

func gatherInfo(p *process.Process) (*ProcessInfo, error) {
	name, _ := p.Name()
	cmdline, _ := p.Cmdline()
	exe, _ := p.Exe()
	username, _ := p.Username()
	parentPID, _ := p.Ppid()
	cpuPercent, _ := p.CPUPercent()
	memInfo, _ := p.MemoryInfo()
	openfiles, _ := p.OpenFiles()

	timestamp, err := p.CreateTime()
	if err != nil {
		log.Fatalln(err)
	}

	parentTree, _ := buildParentTree(p)

	info := &ProcessInfo{
		PID:        p.Pid,
		Name:       name,
		Cmdline:    cmdline,
		Exe:        exe,
		Username:   username,
		ParentPID:  parentPID,
		CPUPercent: cpuPercent,
		RSS:        memInfo.RSS,
		StartTime:  timestamp,
		OpenFiles:  len(openfiles),
		ParentTree: parentTree,
	}
	return info, nil
}

func killProcessBySignal(pid int32, signalType string) error {
	var sig syscall.Signal
	switch strings.ToUpper(signalType) {
	case "TERM":
		sig = syscall.SIGTERM
	case "KILL":
		sig = syscall.SIGKILL
	case "INT":
		sig = syscall.SIGINT
	default:
		sig = syscall.SIGTERM
	}

	return syscall.Kill(int(pid), sig)
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func main() {
	var port uint32
	var outputJSON bool
	var killProcess bool
	var signalType string
	var interactive bool

	rootCmd := &cobra.Command{
		Use:   "portinfo [flags]",
		Short: "Show details of the process listening on a port",
		Run: func(cmd *cobra.Command, args []string) {
			p, err := findProcessByPort(port)
			if err != nil {
				log.Fatalln(err)
			}
			info, err := gatherInfo(p)
			if err != nil {
				log.Fatalln(err)
			}

			// If kill flag is set, kill immediately without prompting
			if killProcess {
				fmt.Printf("Killing process %d (%s) with signal %s\n", info.PID, info.Name, signalType)
				err := killProcessBySignal(info.PID, signalType)
				if err != nil {
					log.Printf("Failed to kill process: %v\n", err)
					os.Exit(1)
				} else {
					fmt.Printf("Signal %s sent to process %d successfully\n", signalType, info.PID)
				}
				return
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(info)
			} else {
				fmt.Printf("PID: %d\n", info.PID)
				fmt.Printf("Name: %s\n", info.Name)
				fmt.Printf("Cmdline: %s\n", info.Cmdline)
				fmt.Printf("Executable: %s\n", info.Exe)
				fmt.Printf("Username: %s\n", info.Username)
				fmt.Printf("Parent PID: %d\n", info.ParentPID)
				fmt.Printf("CPU %%: %.2f\n", info.CPUPercent)
				fmt.Printf("RSS: %d bytes\n", info.RSS)
				fmt.Printf("Start Time (ms since epoch): %d\n", info.StartTime)
				fmt.Printf("Open Files: %d\n", info.OpenFiles)

				if len(info.ParentTree) > 0 {
					fmt.Printf("\nParent Process Tree:\n")
					for i, ancestor := range info.ParentTree {
						indent := strings.Repeat("  ", i+1)
						fmt.Printf("%s└─ PID: %d, Name: %s", indent, ancestor.PID, ancestor.Name)
						if ancestor.Username != "" {
							fmt.Printf(", User: %s", ancestor.Username)
						}
						if ancestor.Cmdline != "" {
							fmt.Printf(", Cmd: %s", ancestor.Cmdline)
						}
						fmt.Printf("\n")
					}
				}
			}

			// Only show interactive form if we're on a TTY or --interactive flag is set
			if interactive || isTerminal() {
				// Create a form to ask the user if they want to kill the process and select a signal
				var interactiveKill bool
				var interactiveSignal string
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[bool]().
							Title("Do you want to kill the process?").
							Options(
								huh.NewOption("Yes", true),
								huh.NewOption("No", false),
							).
							Value(&interactiveKill),
						huh.NewSelect[string]().
							Title("Select signal type").
							Options(
								huh.NewOption("TERM", "TERM"),
								huh.NewOption("KILL", "KILL"),
								huh.NewOption("INT", "INT"),
							).
							Value(&interactiveSignal),
					),
				)

				// Run the form
				err = form.Run()
				if err != nil {
					log.Fatal(err)
				}

				// Implement logic based on user input
				if interactiveKill {
					fmt.Printf("Killing process %d (%s) with signal %s\n", info.PID, info.Name, interactiveSignal)
					err := killProcessBySignal(info.PID, interactiveSignal)
					if err != nil {
						log.Printf("Failed to kill process: %v\n", err)
					} else {
						fmt.Printf("Signal %s sent to process %d successfully\n", interactiveSignal, info.PID)
					}
				} else {
					fmt.Println("Process will not be killed.")
				}
			}
		},
	}

	rootCmd.Flags().Uint32VarP(&port, "port", "p", 0, "port number to inspect")
	rootCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output as JSON")
	rootCmd.Flags().BoolVarP(&killProcess, "kill", "k", false, "kill the process immediately without prompting")
	rootCmd.Flags().StringVarP(&signalType, "signal", "s", "TERM", "signal to send (TERM, KILL, INT)")
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "force interactive mode even when not on a TTY")
	rootCmd.MarkFlagRequired("port")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
