package main

import (
    "os"
    "os/exec"
    "fmt"
    "bufio"
    "io/ioutil"
    "time"

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
    file_p, e := os.Open(c.Path)
    if e == nil {
        i, _ := file_p.Stat()
        c.Size = int(i.Size())
    } else {
        fmt.Sprintf("[FATAL] Could not open config file at path %s", c.Path)
        return
    }
    y := make([]byte, c.Size)
    n, _ := file_p.Read(y)
    if n != c.Size {
        fmt.Sprintf("[WARN] Number of bytes read into config array (%s) does not match config file size (%s). Some directives may have been truncated.", string(n), string(c.Size))
    }
    e = c.ParseYaml(y)
    return
}

func (c ConfigFile) ParseYaml(y []byte) (e error) {
    e = yaml.Unmarshal(y, c.Config)
    return
}

func (f File) Handle() (e error) {
    _, e = os.Open(f.Path)
    if os.IsNotExist(e) {
        if f.Create {
            if f.Directory {
                e = os.Mkdir(f.Path, f.Mode)
                if e != nil {
                    return
                }
            } else {
                if len(f.Content) > 0 {
                    e = ioutil.WriteFile(f.Path, f.Content, f.Mode)
                    if e != nil {
                        return
                    }
                } else {
                    _, e = os.Create(f.Path)
                    if e != nil {
                        return
                    }
                }
            }
        } else {
            return fmt.Errorf("File %s does not exist and Create is not set.", f.Path)
        }
    } else if e != nil {
        return
    }
    e = os.Chown(f.Path, f.Owner, f.Group)
    if e != nil {
        return
    }
    e = os.Chmod(f.Path, f.Mode)
    if e != nil {
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
                return
            } else {
                return
            }
        }
        return
    } else if d.Remove {
        if d.CheckDebInstalledStatus() {
            cmd := exec.Command("apt", "remove", "-y", d.Name)
            e = cmd.Run()
            return
        } else {
            e = fmt.Errorf("Error: Package %s already installed", d.Name)
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
        } else {
            c := exec.Command("service", s.Name, "start")
            e = c.Run()
        }
    } else {
        c := exec.Command("service", s.Name, "stop")
        e = c.Run()
    }
    return
}

func (c Config) Execute() (r Run) {
    r = Run{Start: time.Now(), Results: map[string][]error{}}
    file_r := make([]error, len(c.File))
    for n, f := range c.File {
        file_r[n] = f.Handle()
    }
    deb_r := make([]error, len(c.Deb))
    for n, d := range c.Deb {
        deb_r[n] = d.Handle()
    }
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
        fmt.Print(y)
    }
}
