package main

import (
    "os"
    "os/exec"
    "fmt"
    "bufio"
    "io/ioutil"
    "time"
    "strconv"

    "gopkg.in/yaml.v2"
)

type File struct {
    Path string
    Owner int
    Group int
    Mode os.FileMode
    Create bool
    Directory bool
    Content []byte
}

type Deb struct {
    Name string
    Install bool
    Remove bool
    Upgrade bool
}

type Service struct {
    Name string
    Running bool
    Restart bool
}

type Run struct {
    Start time.Time
    End time.Time
    Config *Config
    Results map[string][]error
}

type ConfigFile struct {
    Path string
    Size int
    Config Config
}

type Config struct {
    File []File
    Deb []Deb
    Service []Service
}

func (c ConfigFile) Init() (e error) {
    fmt.Printf("Config File Path: %s %s", c.Path, "\n")
    file_p, e := os.Open(c.Path)
    if e == nil {
        i, e := file_p.Stat()
        if e != nil {
            fmt.Print(e.Error())
        }
        fmt.Printf("Pointer created. File Name (as seen by pointer) is: %s %s", i.Name(), "\n")
        fmt.Printf("Config file size: %s %s", strconv.FormatInt(i.Size(), 10), "\n")
        c.Size = int(i.Size())
    } else {
        fmt.Printf("[FATAL] Could not open config file. %s", e.Error())
        return
    }
    y := make([]byte, c.Size)
    n, e := file_p.Read(y)
    if e != nil {
        fmt.Print(e.Error())
    }
    if n != c.Size {
        fmt.Printf("[WARN] Number of bytes read into config array (%s) does not match config file size (%s). Some directives may have been truncated.", string(n), string(c.Size))
    }
    e = c.ParseYaml(y)
    if e != nil {
        fmt.Print(e.Error())
    }
    return
}

func (c ConfigFile) ParseYaml(y []byte) (e error) {
    e = yaml.Unmarshal(y, c.Config)
    if e != nil {
        fmt.Printf(e.Error())
    }
    return
}

func (f File) Handle() (e error) {
    _, e = os.Open(f.Path)
    if os.IsNotExist(e) {
        if f.Create {
            if f.Directory {
                e = os.Mkdir(f.Path, f.Mode)
                if e != nil {
                    fmt.Print(e.Error())
                    return
                }
            } else {
                if len(f.Content) > 0 {
                    e = ioutil.WriteFile(f.Path, f.Content, f.Mode)
                    if e != nil {
                        fmt.Print(e.Error())
                        return
                    }
                } else {
                    _, e = os.Create(f.Path)
                    if e != nil {
                        fmt.Print(e.Error())
                        return
                    }
                }
            }
        } else {
            return fmt.Errorf("File %s does not exist and Create is not set.", f.Path)
        }
    } else if e != nil {
        fmt.Print(e.Error())
        return
    }
    e = os.Chown(f.Path, f.Owner, f.Group)
    if e != nil {
        fmt.Print(e.Error())
        return
    }
    e = os.Chmod(f.Path, f.Mode)
    if e != nil {
        fmt.Print(e.Error())
        return
    }
    return
}

func (d Deb) Handle() (e error) {
    e = nil
    if d.Install {
        cmd := exec.Command("apt", "install", "-y", d.Name)
        e = cmd.Run()
        if d.Upgrade {
            cmd := exec.Command("apt", "update")
            e = cmd.Run()
            if e == nil {
                cmd = exec.Command("apt", "upgrade", "-y", d.Name)
                e = cmd.Run()
                if e != nil {
                    fmt.Print(e.Error())
                }
                return
            } else {
                fmt.Print(e.Error())
                return
            }
        }
        return
    } else if d.Remove {
        if d.CheckDebInstalledStatus() {
            cmd := exec.Command("apt", "remove", "-y", d.Name)
            e = cmd.Run()
            if e != nil {
                fmt.Print(e.Error())
            }
            return
        } else {
            e = fmt.Errorf("Error: Package %s already installed", d.Name)
            fmt.Print(e.Error())
            return
        }
    }
    return
}

func (d Deb) CheckDebInstalledStatus() (i bool) {
    c := exec.Command("dpkg", "-l")
    o, _ := c.StdoutPipe()
    c.Run()
    s := bufio.NewScanner(o)
    for s.Scan() {
        if s.Text() == d.Name {
            return true
        }
    }
    return false
}

func (s Service) Handle() (e error) {
    if s.Running {
        if s.Restart{
            c := exec.Command("service", s.Name, "restart")
            e = c.Run()
            if e != nil {
                fmt.Print(e.Error())
            }
        } else {
            c := exec.Command("service", s.Name, "start")
            e = c.Run()
            if e != nil {
                fmt.Print(e.Error())
            }
        }
    } else {
        c := exec.Command("service", s.Name, "stop")
        e = c.Run()
        if e != nil {
            fmt.Print(e.Error())
        }
    }
    return
}

func (c Config) Execute() (r Run) {
    r = Run{Start: time.Now(), Results: map[string][]error{}}
    fmt.Printf("Number of Files to be Targeted: %s %s", strconv.Itoa(len(c.File)), "\n")
    file_r := make([]error, len(c.File))
    for n, f := range c.File {
        file_r[n] = f.Handle()
    }
    fmt.Printf("Number of Debian packages to be targeted: %s %s", strconv.Itoa(len(c.Deb)), "\n")
    deb_r := make([]error, len(c.Deb))
    for n, d := range c.Deb {
        deb_r[n] = d.Handle()
    }
    fmt.Printf("Number of system services to be targeted: %s %s", strconv.Itoa(len(c.Service)), "\n")
    service_r := make([]error, len(c.Service))
    for n, s := range c.Service {
        service_r[n] = s.Handle()
    }
    end := time.Now()
    r.Results["file"], r.Results["deb"], r.Results["service"], r.End = file_r, deb_r, service_r, end
    return
}

func main() {
    config_file_path := os.Args[1]
    config_file := ConfigFile{Path: config_file_path}
    e := config_file.Init()
    if e == nil {
        run := config_file.Config.Execute()
        y, _ := yaml.Marshal(run)
        fmt.Print(string(y))
    }
}
