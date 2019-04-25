package main

import (
    "os"
    "os/exec"
    "fmt"
    "bufio"
    "io/ioutil"
    "time"
)

type Run struct {
    Start time.Time
    End time.Time
    Config []Directive
    // TODO: IMPLEMENT BETTER NAMING (IDEALLY USING A CHECKSUM) FOR ITERATIONS OF A DIRECTIVE.
    Results map[string][]error
}

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

var Config = []Directive{
    Deb{
        Name: "nginx",
        Install: true,
        Upgrade: true,
    },
    Deb{
        Name: "php5-fpm",
        Install: true,
        Upgrade: true,
    },
    File{
       Path: "/etc/nginx/sites-available/default",
       Owner: 1000,
       Group: 1000,
       Mode: 0664,
       Directory: false,
       Create: true,
       Content: "server {\n\tlisten 80 default_server;\n\troot /usr/share/nginx/html;\n\tindex index.php;\n\tlocation ~ \\.php$ {\n\t\ttry_files $uri =404;\n\t\tfastcgi_split_path_info ^(.+\\.php)(/.+)$;\n\t\tinclude fastcgi_params;\n\t\tfastcgi_pass unix:/var/run/php5-fpm.sock;\n\t\tfastcgi_index index.php;\n\t\tfastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n\t}\n}",
    },
    File{
        Path: "/usr/share/nginx/html/index.php",
        Owner: 1000,
        Group: 1000,
        Mode: 0664,
        Directory: false,
        Create: true,
        Content: "<?php\nheader('Content-Type: text/plain');\necho 'Hello, world!';",
    },
    Service{
        Name: "php5-fpm",
        Running: true,
        Restart: true,
    },
    Service{
        Name: "nginx",
        Running: true,
        Restart: true,
    },
}


/*
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
    fmt.Printf("File contents: %s %s", string(y), "\n")
    if e != nil {
        fmt.Print(e.Error())
    }
    if n != c.Size {
        fmt.Printf("[WARN] Number of bytes read into config array (%s) does not match config file size (%s). Some directives may have been truncated.", string(n), string(c.Size))
    }
    e = yaml.Unmarshal(y, &c.Directives)
    if e != nil {
        fmt.Print(e.Error())
    }
    return
}
*/

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

/*

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

    fmt.Printf("Number of directives to execute: %s %s", strconv.Itoa(len(c.directives)), "\n")
    for n, d := range c.directives {
        r.Results[n] = d.handle()
    }
    return
}

*/

func main() {
    results := make([]error, len(Config))
    for n, d := range Config {
        results[n] = d.handle()
    }
    fmt.Print(results)
}
