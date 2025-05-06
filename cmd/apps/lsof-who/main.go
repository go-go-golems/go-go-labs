package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/shirou/gopsutil/v3/net"
    "github.com/shirou/gopsutil/v3/process"
    "github.com/spf13/cobra"
)

type ProcessInfo struct {
    PID        int32   `json:"pid"`
    Name       string  `json:"name"`
    Cmdline    string  `json:"cmdline,omitempty"`
    Exe        string  `json:"exe,omitempty"`
    Username   string  `json:"username,omitempty"`
    ParentPID  int32   `json:"parent_pid,omitempty"`
    CPUPercent float64 `json:"cpu_percent,omitempty"`
    RSS        uint64  `json:"rss_bytes,omitempty"`
    StartTime  int64   `json:"create_time,omitempty"`
    OpenFiles  int     `json:"open_files,omitempty"`
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
    }
    return info, nil
}

func main() {
    var port uint32
    var outputJSON bool

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
            }
        },
    }

    rootCmd.Flags().Uint32VarP(&port, "port", "p", 0, "port number to inspect")
    rootCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "output as JSON")
    rootCmd.MarkFlagRequired("port")

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

