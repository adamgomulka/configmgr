package main

import (
    "os"
    "os/exec"
    "fmt"
    "bufio"
    "io/ioutil"
)

type File struct {
    Path string
    Owner int
    Group int
    Mode int
    Directory bool
    Create bool
    Content string
}

type Deb struct {
    Name string
    Install bool
    Upgrade bool
}

type Service struct {
    Name string
    Running bool
    Restart bool
}

type Directive interface {
    handle() error
}

func (f File) handle() (e error) {
    _, e = os.Open(f.Path)
    if os.IsNotExist(e) || e == nil {
        if f.Create {
            if f.Directory {
                e = os.Mkdir(f.Path, os.FileMode(f.Mode))
                if e != nil {
                    fmt.Print(e.Error())
                    return
                }
            } else {
                if len(f.Content) > 0 {
                    e = ioutil.WriteFile(f.Path, []byte(f.Content), os.FileMode(f.Mode))
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

func main() {
    results := make([]error, len(Config))
    for n, d := range Config {
        results[n] = d.handle()
    }
    fmt.Print(results)
}
