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

type Run struct {
    Start time.Time
    End time.Time
    Config *ConfigFile
    // TODO: IMPLEMENT BETTER NAMING (IDEALLY USING A CHECKSUM) FOR ITERATIONS OF A DIRECTIVE.
    Results map[string][]error
}

type config interface {
    init() error
    Execute() Run
}

type File struct {
    Path string `yaml:"path"`
    Owner int `yaml:"owner"`
    Group int `yaml:"group"`
    Mode int `yaml:"mode"`
    Directory bool `yaml:"directory"`
    Create bool `yaml:"create"`
    Content []byte `yaml:"content",omitempty`
}

type Deb struct {
    Name string `yaml:"name"`
    Install bool `yaml:"install"`
    Upgrade bool `yaml:"upgrade"`
}

type Service struct {
    Name string `yaml:"name"`
    Running bool `yaml:"running"`
    Restart bool `yaml:"restart"`
}

type ConfigFile struct {
    Path string
    Size int
    Directives struct {
        Files []File `yaml:"files"`
        Debs []Deb `yaml:"debs"`
        Services []Service `yaml:"services"`
    } `yaml:"directives"`
}

func (c ConfigFile) init() (e error) {
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
    e = yaml.Unmarshal(y, &c)
    if e != nil {
        fmt.Print(e.Error())
    }
    return
}

func (f File) handle() (e error) {
    _, e = os.Open(f.Path)
    if os.IsNotExist(e) {
        if f.Create {
            if f.Directory {
                e = os.Mkdir(f.Path, os.FileMode(f.Mode))
                if e != nil {
                    fmt.Print(e.Error())
                    return
                }
            } else {
                if len(f.Content) > 0 {
                    e = ioutil.WriteFile(f.Path, f.Content, os.FileMode(f.Mode))
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
    e = os.Chmod(f.Path, os.FileMode(f.Mode))
    if e != nil {
        fmt.Print(e.Error())
        return
    }
    return
}

func (d Deb) handle() (e error) {
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
    } else if !(d.Install) {
        if d.checkDebInstalledStatus() {
            cmd := exec.Command("apt", "remove", "-y", d.Name)
            e = cmd.Run()
            if e != nil {
                fmt.Print(e.Error())
            }
            return
        } else {
            e = fmt.Errorf("Error: Package %s marked for removal, but is not present on the system.", d.Name)
            fmt.Print(e.Error())
            return
        }
    }
    return
}

func (d Deb) checkDebInstalledStatus() (i bool) {
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

func (s Service) handle() (e error) {
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

func (c ConfigFile) Execute() (r Run) {
    r = Run{Start: time.Now(), Results: map[string][]error{}, Config: &c}
    fmt.Printf("Number of Files to be Targeted: %s %s", strconv.Itoa(len(c.Directives.Files)), "\n")
    file_r := make([]error, len(c.Directives.Files))
    for n, f := range c.Directives.Files {
        file_r[n] = f.handle()
    }
    fmt.Printf("Number of Debian packages to be targeted: %s %s", strconv.Itoa(len(c.Directives.Debs)), "\n")
    deb_r := make([]error, len(c.Directives.Debs))
    for n, d := range c.Directives.Debs {
        deb_r[n] = d.handle()
    }
    fmt.Printf("Number of system services to be targeted: %s %s", strconv.Itoa(len(c.Directives.Services)), "\n")
    service_r := make([]error, len(c.Directives.Services))
    for n, s := range c.Directives.Services {
        service_r[n] = s.handle()
    }
    r.Results["file"], r.Results["deb"], r.Results["service"], r.End = file_r, deb_r, service_r, time.Now()

/*
    fmt.Printf("Number of directives to execute: %s %s", strconv.Itoa(len(c.directives)), "\n")
    for n, d := range c.directives {
        r.Results[n] = d.handle()
    }
*/
    return
}

func main() {
    config_file_path := os.Args[1]
    config_file := ConfigFile{Path: config_file_path}
    e := config_file.init()
    if e == nil {
        run := config_file.Execute()
        y, _ := yaml.Marshal(run)
        fmt.Print(string(y))
    }
}
